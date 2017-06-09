package mir

import (
	"github.com/rhysd/locerr"
	"testing"
)

func TestLast(t *testing.T) {
	i1 := &Insn{"test1", nil, nil, nil, locerr.Pos{}}
	i2 := &Insn{"test1", nil, i1, nil, locerr.Pos{}}
	i1.Prev = i2

	if i1 != i2.Last() {
		t.Errorf("last node is actually not last")
	}

	if i1 != i1.Last() {
		t.Errorf("last of last node should be itself")
	}
}

func TestInsnAppend(t *testing.T) {
	i1 := &Insn{"test1", nil, nil, nil, locerr.Pos{}}
	i2 := &Insn{"test2", nil, i1, nil, locerr.Pos{}}
	i1.Prev = i2

	i3 := &Insn{"test3", nil, nil, nil, locerr.Pos{}}
	i4 := &Insn{"test4", nil, i3, nil, locerr.Pos{}}
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
	i1 := &Insn{"test1", nil, nil, nil, locerr.Pos{}}
	i2 := &Insn{"test2", nil, i1, nil, locerr.Pos{}}
	i1.Prev = i2

	i3 := &Insn{"test3", nil, nil, nil, locerr.Pos{}}
	i4 := &Insn{"test4", nil, i3, nil, locerr.Pos{}}
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
	i1 := &Insn{"test1", nil, nil, nil, locerr.Pos{}}
	i2 := &Insn{"test1", nil, i1, nil, locerr.Pos{}}
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

func TestEmptyArrayFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Trying to create empty block should make panic")
		}
	}()
	NewBlockFromArray("test", []*Insn{})
}

func TestBlockPrepend(t *testing.T) {
	i := NewInsn("$k1", UnitVal, locerr.Pos{})
	j := NewInsn("$k2", UnitVal, locerr.Pos{})
	b := NewBlockFromArray("test", []*Insn{i})
	b.Prepend(j)
	if j.Next != i || i.Prev != j || j.Prev != b.Top || b.Top.Next != j {
		t.Fatalf("Instruction was not prepended correctly")
	}
}

func TestBlockAppend(t *testing.T) {
	i := NewInsn("$k1", UnitVal, locerr.Pos{})
	j := NewInsn("$k2", UnitVal, locerr.Pos{})
	b := NewBlockFromArray("test", []*Insn{i})
	b.Append(j)
	if i.Next != j || j.Prev != i || b.Bottom.Prev != j || j.Next != b.Bottom {
		t.Fatalf("Instruction was not appended correctly")
	}
}
