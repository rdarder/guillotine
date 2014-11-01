package guillotine

import "fmt"

var _ = fmt.Println

type Direction bool

const (
	HORIZONTAL = false
	VERTICAL   = true
)

type Join uint8

const JOIN Join = 0

func (j Join) direct(d Direction) Join {
	if d == VERTICAL {
		return j | DIRECTION_MASK
	} else {
		return j &^ DIRECTION_MASK
	}
}
func (j Join) irotated() Join {
	return j | IROT_MASK
}
func (j Join) istraight() Join {
	return j &^ IROT_MASK
}

func (j Join) jrotated() Join {
	return j | JROT_MASK
}

func (j Join) jstraight() Join {
	return j &^ JROT_MASK
}

const (
	DIRECTION_MASK = 1 << iota
	IROT_MASK
	JROT_MASK
)

func (c Join) direction() Direction {
	return (c & DIRECTION_MASK) != 0
}
func (c Join) irot() bool {
	return (c & IROT_MASK) != 0
}
func (c Join) jrot() bool {
	return (c & JROT_MASK) != 0
}

/*This tree design ended out very error prone, due to the indexing
scheme. When the time comes for a refactor consider the following
requirements that led to this design.
- the layout is expressed as a binary tree.
- LayoutTree := pick(board, rot) | join(lt, lt, orientation)
- the tree builder refer to disjoint subsets by one of its leafs,
beforehand if the tree root will be a leaf or a node.
- leafs and nodes carry different data.
- fast union-find with find compression.
- the tree evaluator traverses the tree in post order.
- similart to a reverse polish notation, we attempt to keep the
evaluator state in a stack. in this case it's a slice which length is
the node count
*/

type PickLeaf struct {
	//1-based node index
	//0 means no parent
	Parent uint16
	Rot    bool
}

type StackNode struct {
	// 0-based mixed index.
	//0 to n-1 => leaf index
	//n to 2n-2 => node index
	Left, Right uint16
	//1-based node index
	//0 means no parent
	Parent    uint16
	Direction Direction `json:"Vertical"`
}

type LayoutTree struct {
	Picks    []PickLeaf  //size N for N boards
	Stacks   []StackNode //size N-1 for N boards
	nextNode uint16      //0-based node index
	nboards  uint16
}

func emptyLayoutTree(n uint16) *LayoutTree {
	return &LayoutTree{
		make([]PickLeaf, n, n), make([]StackNode, n-1, n-1), 0, n}
}

//Join two boards if they're not already connected together.
//Two boards can be already connected either directly or through
//other Joins. They're connected if they belong to the same tree
//component.
//Boards must be referred to by their index in the CutSpec.
//config is a Join configuration, which specifies whether the
//boards are rotated or not, and whether the join is vertical or
//horizontal. Rotation configuration is only considered if the board
//to be joined hasn't been picked before. The first pick determines
//rotation
func (t *LayoutTree) take(i, j uint16, config Join) bool {
	iRoot := t.getLeafRoot(i)
	jRoot := t.getLeafRoot(j)
	if iRoot == jRoot {
		return false
	} else {
		t.setNode(t.nextNode, iRoot, jRoot, config)
		t.nextNode += 1
		return true
	}
}

func (t *LayoutTree) setNode(i, left, right uint16, config Join) {
	node := &t.Stacks[i]
	node.Direction = config.direction()
	node.Left = left
	node.Right = right
	t.setChild(left, i, config.irot())
	t.setChild(right, i, config.jrot())
}

func (t *LayoutTree) setChild(i, parent uint16, rot bool) {
	if i < t.nboards {
		t.Picks[i].Parent = parent + 1
		t.Picks[i].Rot = rot
	} else {
		t.Stacks[i-t.nboards].Parent = parent + 1
	}

}

type Fitness func(t *LayoutTree, spec *CutSpec) uint

func (t *LayoutTree) areaWalk(spec *CutSpec) []Board {
	state := make([]Board, len(spec.Boards)-1)
	//state is an [n-1]Board for n boards
	nboards := len(t.Picks)
	if len(state) < nboards-1 {
		panic("State too small")
	}
	for i, stack := range t.Stacks {
		first := t.getBoard(stack.Left, spec.Boards, state)
		second := t.getBoard(stack.Right, spec.Boards, state)
		switch stack.Direction {
		case VERTICAL:
			state[i] = first.Vstack(second)
		case HORIZONTAL:
			state[i] = first.Hstack(second)
		}
	}
	return state
}

func (t *LayoutTree) Area(spec *CutSpec) uint {
	return t.areaWalk(spec)[len(spec.Boards)-2].Area()
}

type Rect struct {
	X, Y, Width, Height uint
}
type Drawer struct {
	lt    *LayoutTree
	state []Board
	spec  *CutSpec
}

func NewDrawer(lt *LayoutTree, spec *CutSpec) *Drawer{
	return &Drawer{lt: lt, state: lt.areaWalk(spec), spec: spec}	
}
//needs cleanup
func (d *Drawer) Draw() []Rect {
	nboards := len(d.spec.Boards)
	boxes := make([]Rect, nboards)
	for i, board := range d.spec.Boards {
		if d.lt.Picks[i].Rot {
			board = board.rotated()
		}
		boxes[i].Width = board.Width
		boxes[i].Height = board.Height
	}
	d.DrawWithOffset(2*nboards-2, Board{0, 0}, boxes)
	return boxes
}

//throw away and redo
func (d *Drawer) DrawWithOffset(i int, offset Board, boxes []Rect) {
	nboards := len(d.spec.Boards)
	if i < nboards {
		boxes[i].X = offset.Width
		boxes[i].Y = offset.Height
	} else {
		stack := d.lt.Stacks[i-nboards]
		d.DrawWithOffset(int(stack.Left), offset, boxes)
		leftOffset := d.lt.getBoard(stack.Left, d.spec.Boards, d.state)
		if stack.Direction == VERTICAL {
			offset = Board{offset.Width, offset.Height + leftOffset.Height}
		} else {
			offset = Board{offset.Width + leftOffset.Width, offset.Height}
		}
		d.DrawWithOffset(int(stack.Right), offset, boxes)
	}
}

var _ Fitness = (*LayoutTree).Area

func (t *LayoutTree) getBoard(i uint16, orig []Board, state []Board) Board {
	if i < t.nboards {
		if t.Picks[i].Rot {
			return orig[i].rotated()
		} else {
			return orig[i]
		}
	} else {
		return state[i-t.nboards]
	}
}

//Gets the root of a leaf's tree. The root is encoded as:
//0 if the leaf is th root
//j+1 if the node j is the root
//The leaf is identified by its index in the layoutTree picks slice.
func (t *LayoutTree) getLeafRoot(i uint16) uint16 {
	if pick := &t.Picks[i]; pick.Parent == 0 {
		return i
	} else {
		pick.Parent = t.getNodeRoot(pick.Parent)
		return pick.Parent + t.nboards - 1
	}
}

//Gets the root of a node's tree. The node is identified by
//it's 1-based index in the layoutTree stacks slice.
//returns the 1-based index of root node
//returns: 1-based node-index
func (t *LayoutTree) getNodeRoot(i uint16) uint16 {
	if node := &t.Stacks[i-1]; node.Parent == 0 {
		return i
	} else {
		node.Parent = t.getNodeRoot(node.Parent)
		return node.Parent
	}
}
