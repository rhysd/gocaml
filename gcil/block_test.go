package gcil

import (
	"testing"
)

func TestHasNext(t *testing.T) {
	i1 := &Insn{
		"test1",
		nil,
		UnitVal,
		nil,
	}
	i2 := &Insn{
		"test1",
		nil,
		UnitVal,
		i1,
	}

	if !i2.HasNext() {
		t.Errorf("Head should have next")
	}
	if i1.HasNext() {
		t.Errorf("Tail should not have next")
	}
}

func TestLast(t *testing.T) {
	i1 := &Insn{
		"test1",
		nil,
		UnitVal,
		nil,
	}
	i2 := &Insn{
		"test1",
		nil,
		UnitVal,
		i1,
	}

	if i1 != i2.Last() {
		t.Errorf("last node is actually not last")
	}

	if i1 != i1.Last() {
		t.Errorf("last of last node should be itself")
	}
}

func TestReverse(t *testing.T) {
	i1 := &Insn{
		"test1",
		nil,
		UnitVal,
		nil,
	}
	i2 := &Insn{
		"test1",
		nil,
		UnitVal,
		i1,
	}

	i3 := reverseDirection(i2)
	if i1 != i3 {
		t.Errorf("previous bottom must be head of reversed list")
	}

	if i3.Next != i2 {
		t.Errorf("list direction must be reversed but actually %v", i3.Next)
	}

	if i2.Next != nil {
		t.Errorf("bottom of list must be ended with nil but actually %v", i2.Next)
	}
}
