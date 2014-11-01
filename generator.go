package guillotine

import (
	"math/rand"
	"sort"
)

type MeasuredBoards struct {
	Boards    []Board
	CumArea   []int
	TotalArea int
}

func (mb *MeasuredBoards) Calc() {
	var cumArea int
	for i := 0; i < len(mb.Boards); i++ {
		area := mb.Boards[i].Area()
		cumArea += int(area * area)
		mb.CumArea[i] = cumArea
	}
	mb.TotalArea = int(cumArea)
}

func NewRandomSpec(nboards, width, height int, r *rand.Rand) (spec *CutSpec) {
	boards := make([]Board, 0, nboards)
	cumArea := make([]int, nboards)
	mb := &MeasuredBoards{Boards: boards, CumArea: cumArea}
	mb.Boards = append(mb.Boards, Board{Width: uint(width), Height: uint(height)})
	for len(mb.Boards) < nboards {
		mb.Calc()
		target := r.Intn(mb.TotalArea)
		i := sort.Search(len(mb.Boards), func(j int) bool { return mb.CumArea[j] > target })
		toSplit := mb.Boards[i]
		if b1, b2, didSplit := splitBoard(toSplit, r); didSplit {
			mb.Boards[i] = b1
			mb.Boards = append(mb.Boards, b2)
		}
	}
	return &CutSpec{Boards: mb.Boards}
}

func splitBoard(b Board, r *rand.Rand) (b1, b2 Board, didSplit bool) {
	if b.Height < 2 || b.Width < 2 {
		return
	}
	if b.Height > b.Width {
		b1, b2 = b.Hsplit(rUintn(r, 1, b.Height))
		return b1, b2, true
	} else {
		b1, b2 = b.Vsplit(rUintn(r, 1, b.Width))
		return b1, b2, true
	}
}

func rUintn(r *rand.Rand, min, max uint) uint {
	return uint(r.Intn(int(max-min))) + min
}
