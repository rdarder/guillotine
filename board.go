package guillotine

func max(a, b uint) uint {
	if a >= b {
		return a
	} else {
		return b
	}
}

type Board struct {
	Width, Height uint
}

func (b Board) rotated() Board {
	return Board{b.Height, b.Width}
}

func (left Board) Hstack(right Board) Board {
	return Board{left.Width + right.Width, max(left.Height, right.Height)}
}

func (top Board) Vstack(bottom Board) Board {
	return Board{max(top.Width, bottom.Width), top.Height + bottom.Height}
}

func (b Board) Hsplit(y uint) (b1, b2 Board) {
	if y > b.Height {
		panic("invalid split position")
	}
	return Board{b.Width, y}, Board{b.Width, b.Height - y}
}

func (b Board) Vsplit(x uint) (b1, b2 Board) {
	if x > b.Width {
		panic("invalid split position")
	}
	return Board{x, b.Height}, Board{b.Width - x, b.Height}
}

func (board Board) Area() uint {
	return board.Width * board.Height
}

type CutSpec struct {
	Boards []Board
}

func (spec *CutSpec) Add(width, height uint) *CutSpec {
	spec.Boards = append(spec.Boards, Board{width, height})
	return spec
}

func newCutSpec() *CutSpec {
	return &CutSpec{make([]Board, 0, 10)}
}
