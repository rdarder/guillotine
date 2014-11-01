package guillotine

import (
	"math/rand"
	"sort"
	"math"
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

func AreaDimensions(area float64, r *rand.Rand) (width, height uint){
	mean := math.Sqrt(area)
	stddev := mean/4
	fwidth := (r.NormFloat64() * stddev + mean)
	if fwidth > area/10{
		fwidth = area/10
	} else if fwidth < 10 {
		fwidth = 10
	}
	return uint(fwidth), uint(area/fwidth)	
}

func MaxWidthDimensions(maxWidth int, r *rand.Rand) (width,height uint) {
		width = uint(maxWidth)
		height = uint(r.NormFloat64() * float64(maxWidth) + float64(4*maxWidth))	
		return
}

func NewRandomSpec(nboards int, width, height uint, r *rand.Rand, limitWidth bool) (spec *CutSpec) {
	boards := make([]Board, 0, nboards)
	cumArea := make([]int, nboards)
	mb := &MeasuredBoards{Boards: boards, CumArea: cumArea}
	mb.Boards = append(mb.Boards, Board{Width: width, Height: height})
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
	spec = &CutSpec{Boards: mb.Boards}
	if limitWidth{
		spec.MaxWidth = width
	}
	return spec
}

func splitBoard(b Board, r *rand.Rand) (b1, b2 Board, didSplit bool) {
	if b.Height < 2 || b.Width < 2 {
		return
	}
	if b.Height > b.Width && r.Float32() > 0.3 {
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
