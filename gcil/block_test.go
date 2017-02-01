package gcil

import (
	"testing"
)

func TestLast(t *testing.T) {
	i1 := &Insn{"test1", nil, UnitVal, nil, nil}
	i2 := &Insn{"test1", nil, UnitVal, i1, nil}
	i1.Prev = i2

	if i1 != i2.Last() {
		t.Errorf("last node is actually not last")
	}

	if i1 != i1.Last() {
		t.Errorf("last of last node should be itself")
	}
}

func TestAppend(t *testing.T) {
	i1 := &Insn{"test1", nil, UnitVal, nil, nil}
	i2 := &Insn{"test2", nil, UnitVal, i1, nil}
	i1.Prev = i2

	i3 := &Insn{"test3", nil, UnitVal, nil, nil}
	i4 := &Insn{"test4", nil, UnitVal, i3, nil}
	i3.Prev = i4

	i2.Append(i4)

	strings := []string{"test2", "test1", "test4", "test3"}

	insn := i2
	for i, s := range strings {
		if insn.Ident != s {
			t.Errorf("While forwarding list %dth insn must be '%s' but actually '%s'", i, s, insn.Ident)
		}
		if insn.Next != nil && insn.Next.Prev != insn {
			t.Errorf("Prev does not point previous node properly. Expected %v but actually %v", insn, insn.Next.Prev)
		}
		insn = insn.Next
	}
}

func TestConcat(t *testing.T) {
	i1 := &Insn{"test1", nil, UnitVal, nil, nil}
	i2 := &Insn{"test2", nil, UnitVal, i1, nil}
	i1.Prev = i2

	i3 := &Insn{"test3", nil, UnitVal, nil, nil}
	i4 := &Insn{"test4", nil, UnitVal, i3, nil}
	i3.Prev = i4

	i5 := Concat(i2, i4)

	strings := []string{"test2", "test1", "test4", "test3"}

	insn := i5
	for i, s := range strings {
		if insn.Ident != s {
			t.Errorf("While forwarding list %dth insn must be '%s' but actually '%s'", i, s, insn.Ident)
		}
		if insn.Next != nil && insn.Next.Prev != insn {
			t.Errorf("Prev does not point previous node properly. Expected %v but actually %v", insn, insn.Next.Prev)
		}
		insn = insn.Next
	}
}

func TestReverse(t *testing.T) {
	i1 := &Insn{"test1", nil, UnitVal, nil, nil}
	i2 := &Insn{"test1", nil, UnitVal, i1, nil}
	i1.Prev = i2

	i3 := Reverse(i2)
	if i1 != i3 {
		t.Errorf("previous bottom must be head of reversed list")
	}

	if i3.Next != i2 {
		t.Errorf("list direction must be reversed but actually %v", i3.Next)
	}
	if i3.Prev != nil {
		t.Errorf("prev of top of reversed list must be null but %v", i3.Prev)
	}

	if i2.Next != nil {
		t.Errorf("bottom of list must be ended with nil but actually %v", i2.Next)
	}
	if i2.Prev != i1 {
		t.Errorf("prev of bottom of reversed list must be i1 but %v", i2.Prev)
	}
}
