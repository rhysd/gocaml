// Package gcil provides definition of GCIL and converter from AST.
//
// GCIL is an abbreviation of GoCaml Intermediate Language.
// It's an original intermediate language to fill the gap between machine code and
// syntax tree.
// GCIL is a SSA form and K-normalized, and has high-level type information.
//
// It discards many things from syntax tree because it's no longer needed.
// For example, position of nodes, display name of symbols and nested tree structure are discarded.
//
// GCIL consists of block (basic block), instruction and value.
// There is a one root block. Block contains sequence of instructions.
// Instruction contains a bound identifier name and its value.
// Some value (`if`, `fun`, ...) contains recursive blocks.
//
// Please see spec file in the gocaml repository.
//
// https://github.com/rhysd/gocaml/blob/master/gcil/README.md
//
// You can see its string representation by command
//
//		gocaml -gcil test.ml
//
// e.g.
//
//		let x = 1 in
//		let rec f a b = if a < 0 then a + b - x else x in
//		if true then print_int (f 3 4) else ()
//
//		root:
//		x$t1 = int 1
//		f$t2 = fun a$t3,b$t4
//		  $k1 = int 0
//		  $k2 = less a$t3 $k1
//		  $k3 = if $k2
//		  then:
//		    $k4 = add $at3 $bt4
//		    $k5 = sub $k4 x$t1
//		  else:
//		    $k6 = ref x$t1
//		$k7 = bool true
//		$k8 = if $k7
//		  then:
//		    $k9 = xref print_int
//		    $k10 = ref f$t2
//		    $k11 = int 3
//		    $k12 = int 4
//		    $k13 = app $k10 $k11,$k12
//		    $k14 = app $k9 $k13
//		  else:
//		    $k15 = unit
//
package gcil

import (
	"fmt"
	"github.com/rhysd/gocaml/typing"
	"io"
)

type Block struct {
	Top    *Insn
	Bottom *Insn
	Name   string
}

func (bl *Block) Println(out io.Writer) {
	fmt.Fprintf(out, "BEGIN: %s\n", bl.Name)
	for i := bl.Top; i != nil; i = i.Next {
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
	Prev  *Insn
}

func (insn *Insn) Println(out io.Writer) {
	fmt.Fprintf(out, "%s = ", insn.Ident)
	insn.Val.Print(out)
	var s string
	if insn.Ty == nil {
		s = "(unknown)"
	} else {
		s = insn.Ty.String()
	}
	fmt.Fprintf(out, " ; type=%s\n", s)
}

func (insn *Insn) Last() *Insn {
	i := insn
	for i.Next != nil {
		i = i.Next
	}
	return i
}

func (insn *Insn) Append(other *Insn) {
	last := insn.Last()
	last.Next = other
	if other != nil {
		other.Prev = last
	}
}

func NewInsn(n string, t typing.Type, v Val) *Insn {
	return &Insn{n, t, v, nil, nil}
}

func Concat(a, b *Insn) *Insn {
	a.Append(b)
	return a
}

// Reverse the instruction list. `insn` is assumed to point head of the list
func Reverse(insn *Insn) *Insn {
	i := insn
	for {
		i.Next, i.Prev = i.Prev, i.Next
		if i.Prev == nil {
			return i
		}
		i = i.Prev
	}
}
