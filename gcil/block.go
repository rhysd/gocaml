package gcil

import (
	"fmt"
	"github.com/rhysd/gocaml/typing"
	"io"
)

// GCIL is an abbreviation of GoCaml Intermediate Language.
// It's an original intermediate language to fill the gap between machine code and
// syntax tree.
// GCIL is a SSA form and K-normalized, and has high-level type information.
//
// It discards many things from syntax tree because it's no longer needed.
//   - Position of nodes
//   - Display name of symbols
//   - Nested tree structure
//
// GCIL consists of block, instruction, value, branch and function.
// All 'let' nodes are flatten.

// e.g.
//
// let x = 1 in
// let rec f a b = if a < 0 then a + b - x else x in
// if true then print_int (f 3 4) else ()
//
// x$t1 = int 1
// f$t2 = fun a$t3,b$t4
//   $k1 = int 0
//   $k2 = less a$t3 $k1
//   $k3 = if $k2
//   then:
//     $k4 = add $at3 $bt4
//     $k5 = sub $k4 x$t1
//   else:
//     $k6 = ref x$t1
// $k7 = bool true
// $k8 = if $k7
//   then:
//	   $k9 = xref print_int
//     $k10 = ref f$t2
//     $k11 = int 3
//     $k12 = int 4
//     $k13 = app $k10 $k11,$k12
//     $k14 = app $k9 $k13
//   else:
//     $k15 = unit

type Block struct {
	Insns *Insn
	Name  string
}

func (bl *Block) LastInsn() *Insn {
	if bl.Insns == nil {
		panic("Block is empty")
	}
	return bl.Insns.Last()
}

func (bl *Block) Println(out io.Writer) {
	fmt.Fprintf(out, "BEGIN: %s\n", bl.Name)
	for i := bl.Insns; i != nil; i = i.Next {
		i.Println(out)
	}
	fmt.Fprintf(out, "END: %s\n", bl.Name)
}

// Instruction.
// Its form is always `ident = val`
type Insn struct {
	Ident string
	Ty    typing.Type
	Val   Val
	Next  *Insn
}

func (insn *Insn) Println(out io.Writer) {
	fmt.Fprintf(out, "%s = ", insn.Ident)
	insn.Val.Print(out)
	fmt.Fprintf(out, " ; type=%s\n", insn.Ty.String())
}

func (insn *Insn) HasNext() bool {
	return insn.Next != nil
}

func (insn *Insn) Last() *Insn {
	i := insn
	for i.HasNext() {
		i = i.Next
	}
	return i
}

// Reverse the instruction list. `insn` is assumed to point head of the list
func Reverse(insn *Insn) *Insn {
	i, j := insn, insn.Next
	for j != nil {
		tmp := j.Next
		j.Next = i
		i, j = j, tmp
	}
	insn.Next = nil
	return i
}
