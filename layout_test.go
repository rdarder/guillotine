package guillotine

import "testing"
import "fmt"

var _ = fmt.Println

func wrongArea(t *testing.T, expected, got int) {
	t.Errorf("Expected area to be [%v], got [%v]", expected, got)
}

func TestJoinBits(t *testing.T) {
	if j := JOIN.direct(HORIZONTAL); j.direction() != HORIZONTAL {
		t.Error("Expected horizontal Join")
	}
	if j := JOIN.direct(VERTICAL); j.direction() != VERTICAL {
		t.Error("Expected vertical Join")
	}
	j := JOIN.direct(VERTICAL).direct(HORIZONTAL).irotated().jrotated()
	if j.direction() != HORIZONTAL {
		t.Error("Expected horizontal Join")
	} else if !j.irot() {
		t.Error("Expected i rotated Join")
	} else if !j.jrot() {
		t.Error("Expected j rotated Join")
	}
	if j := JOIN.irotated().istraight(); j.irot() {
		t.Error("Expected i straight Join")
	}
	if j := JOIN.jrotated().jstraight(); j.jrot() {
		t.Error("Expected j straight Join")
	}

}

func TestTwoBoards(t *testing.T) {
	spec := newCutSpec().Add(1, 6).Add(4, 5)
	lt := emptyLayoutTree(2)
	lt.take(0, 1, JOIN.direct(HORIZONTAL))
	state := make([]Board, 1, 1)
	if area, err := lt.area(spec, state); err != nil {
		t.Error(err)
	} else if area != 30 {
		wrongArea(t, 30, area)
	}

}
func TestThreeBoards(t *testing.T) {
	spec := newCutSpec().Add(1, 6).Add(4, 5).Add(5, 2)
	lt := emptyLayoutTree(3)
	lt.take(0, 1, JOIN.direct(HORIZONTAL))
	lt.take(0, 2, JOIN.direct(VERTICAL))
	state := make([]Board, 2, 2)
	if area, err := lt.area(spec, state); err != nil {
		t.Error(err)
	} else if area != 40 {
		wrongArea(t, 40, area)
	}
}

func TestBiggerTree(t *testing.T) {
	spec := newCutSpec().Add(1, 2).Add(2, 2).Add(1, 5).Add(3, 1)
	spec = spec.Add(2, 7).Add(4, 2).Add(5, 3).Add(2, 6)
	lt := emptyLayoutTree(8)
	lt.take(0, 5, JOIN.direct(HORIZONTAL))
	lt.take(1, 5, JOIN.direct(HORIZONTAL).irotated())
	lt.take(2, 3, JOIN.direct(VERTICAL).jrotated())
	lt.take(4, 5, JOIN.direct(HORIZONTAL))
	lt.take(6, 7, JOIN.direct(VERTICAL).irotated().jrotated())
	lt.take(6, 3, JOIN.direct(VERTICAL))
	lt.take(3, 5, JOIN.direct(HORIZONTAL))
	state := make([]Board, 7, 7)
	if area, err := lt.area(spec, state); err != nil {
		t.Error(err)
	} else if area != 225 {
		wrongArea(t, 225, area)
	}
}
