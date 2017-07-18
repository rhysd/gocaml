// Package mir provides definition of MIR and converter from AST.
//
// MIR is an abbreviation of GoCaml Intermediate Language.
// It's an original intermediate language to fill the gap between machine code and
// syntax tree.
// MIR is a SSA form and K-normalized, and has high-level type information.
//
// It discards many things from syntax tree because it's no longer needed.
// For example, position of nodes, display name of symbols and nested tree structure are discarded.
//
// MIR consists of block (basic block), instruction and value.
// There is a one root block. Block contains sequence of instructions.
// Instruction contains a bound identifier name and its value.
// Some value (`if`, `fun`, ...) contains recursive blocks.
//
// Please see spec file in the gocaml repository.
//
// https://github.com/rhysd/gocaml/blob/master/mir/README.md
//
// You can see its string representation by command
//
//		gocaml -mir test.ml
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
package mir

import (
	"github.com/rhysd/locerr"
)

// Block struct represents basic block.
// It has a name and instruction sequence to execute.
// Note that top and bottom of the sequence are always NOP instruction in order to
// make modifying instructions easy.
type Block struct {
	Top    *Insn
	Bottom *Insn
	Name   string
}

func NewBlock(name string, top, bottom *Insn) *Block {
	start := &Insn{"", NOPVal, top, nil, locerr.Pos{}}
	top.Prev = start
	end := &Insn{"", NOPVal, nil, bottom, locerr.Pos{}}
	bottom.Next = end
	return &Block{start, end, name}
}

func NewEmptyBlock(name string) *Block {
	start := &Insn{"", NOPVal, nil, nil, locerr.Pos{}}
	end := &Insn{"", NOPVal, nil, nil, locerr.Pos{}}
	start.Next = end
	end.Prev = start
	return &Block{start, end, name}
}

func NewBlockFromArray(name string, insns []*Insn) *Block {
	if len(insns) == 0 {
		panic("Block must contain at least one instruction")
	}

	top := insns[0]
	bottom := top
	for _, insn := range insns[1:] {
		insn.Prev = bottom
		bottom.Next = insn
		bottom = insn
	}

	return NewBlock(name, top, bottom)
}

func (b *Block) Prepend(i *Insn) {
	i.Next = b.Top.Next
	i.Prev = b.Top
	b.Top.Next.Prev = i
	b.Top.Next = i
}

func (b *Block) Append(i *Insn) {
	i.Next = b.Bottom
	i.Prev = b.Bottom.Prev
	b.Bottom.Prev.Next = i
	b.Bottom.Prev = i
}

// Returns range [begin, end)
func (b *Block) WholeRange() (begin *Insn, end *Insn) {
	begin = b.Top.Next
	end = b.Bottom
	return
}

// Instruction.
// Its form is always `ident = val`
type Insn struct {
	Ident string
	Val   Val
	Next  *Insn
	Prev  *Insn
	Pos   locerr.Pos
}

func (insn *Insn) Last() *Insn {
	i := insn
	for i.Next != nil {
		i = i.Next
	}
	return i
}

func (insn *Insn) Append(other *Insn) {
	if other == nil {
		return
	}
	last := insn.Last()
	last.Next = other
	other.Prev = last
}

func (insn *Insn) RemoveFromList() {
	insn.Next.Prev = insn.Prev
	insn.Prev.Next = insn.Next
}

func NewInsn(n string, v Val, pos locerr.Pos) *Insn {
	return &Insn{n, v, nil, nil, pos}
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
