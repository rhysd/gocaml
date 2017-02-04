package gcil

import (
	"testing"
)

type testInsnCounter struct {
	count uint
}

func (c *testInsnCounter) Visit(_ *Insn) Visitor {
	c.count++
	return c
}

type testCancelVisitor struct {
}

func (v *testCancelVisitor) Visit(i *Insn) Visitor {
	if i.Ident == "$k1" {
		return nil
	}
	return v
}

func TestVisitor(t *testing.T) {
	root := NewBlockFromArray("program", []*Insn{
		NewInsn(
			"$k1",
			&Int{42},
		),
		NewInsn(
			"$k2",
			&Unary{NEG, "$k1"},
		),
		NewInsn(
			"$k3",
			&If{
				"b",
				NewBlockFromArray("then", []*Insn{
					NewInsn(
						"$k4",
						&Int{0},
					),
				}),
				NewBlockFromArray("else", []*Insn{
					NewInsn(
						"$k5",
						&Fun{
							[]string{"a", "b"},
							NewBlockFromArray("body ($k5)", []*Insn{
								NewInsn(
									"$k6",
									&Int{42},
								),
								NewInsn(
									"$k7",
									&Unary{NEG, "$k1"},
								),
							}),
						},
					),
					NewInsn(
						"$k8",
						&App{
							"$k5",
							[]string{"$k1", "$k2"},
						},
					),
				}),
			},
		),
	})

	c := &testInsnCounter{0}
	cancelled := Visit(c, root)

	if cancelled {
		t.Errorf("Visitor was unexpectedly cancelled: %v", *c)
	}

	if c.count != 8 {
		t.Errorf("misamtch number of instructions in test code (8 was expected but actually %d)", c.count)
	}
}

func TestCancel(t *testing.T) {
	root := NewBlockFromArray("program", []*Insn{
		NewInsn(
			"$k1",
			&Int{42},
		),
	})

	v := &testCancelVisitor{}

	if !Visit(v, root) {
		t.Errorf("Should have been cancelled but actually not cancelled")
	}

	// First instruction is 'NOP'
	root.Top.Next.Ident = "$k2"

	if Visit(v, root) {
		t.Errorf("Should not have been cancelled but actually cancelled")
	}
}
