package closure

import (
	"fmt"
	"github.com/rhysd/gocaml/gcil"
)

type freeVarsGatherer struct {
	found     nameSet
	transform *transformWithKFO
}

func (fvg *freeVarsGatherer) add(name string) {
	fvg.found[name] = struct{}{}
}

func (fvg *freeVarsGatherer) exploreBlock(block *gcil.Block) {
	// Traverse instructions in the block in reverse order.
	// First and last instructions are NOP, so skipped.
	for i := block.Bottom.Prev; i.Prev != nil; i = i.Prev {
		fvg.exploreInsn(i)
	}
}

func (fvg *freeVarsGatherer) exploreTillTheEnd(insn *gcil.Insn) {
	end := insn
	for end.Next.Next != nil {
		// Find the last instruction before NOP
		end = end.Next
	}
	for i := end; i != insn; i = i.Prev {
		fvg.exploreInsn(i)
	}
}

func (fvg *freeVarsGatherer) exploreInsn(insn *gcil.Insn) {
	switch val := insn.Val.(type) {
	case *gcil.Unary:
		fvg.add(val.Child)
	case *gcil.Binary:
		fvg.add(val.Lhs)
		fvg.add(val.Rhs)
	case *gcil.Ref:
		fvg.add(val.Ident)
	case *gcil.If:
		fvg.add(val.Cond)
		fvg.exploreBlock(val.Then)
		fvg.exploreBlock(val.Else)
	case *gcil.App:
		// Should not add val.Callee to free variables if it is not a closure
		// because a normal function is treated as label, not a variable
		// (label is a constant).
		if _, ok := fvg.transform.knownFuns[val.Callee]; !ok {
			fvg.add(val.Callee)
		}
		for _, a := range val.Args {
			fvg.add(a)
		}
	case *gcil.Tuple:
		for _, e := range val.Elems {
			fvg.add(e)
		}
	case *gcil.Array:
		fvg.add(val.Size)
		fvg.add(val.Elem)
	case *gcil.TplLoad:
		fvg.add(val.From)
	case *gcil.ArrLoad:
		fvg.add(val.From)
		fvg.add(val.Index)
	case *gcil.ArrStore:
		fvg.add(val.To)
		fvg.add(val.Index)
		fvg.add(val.Rhs)
	case *gcil.MakeCls:
		fv, ok := fvg.transform.closureBlockFreeVars[val.Fun]
		if !ok {
			panic(fmt.Sprintf("Applying unknown closure '%s'", insn.Ident))
		}
		for v := range fv {
			fvg.add(v)
		}
		for _, v := range val.Vars {
			fvg.add(v)
		}
		delete(fvg.found, val.Fun)
	}

	// Note:
	// Functions in tree will be moved to toplevel. So they should be ignored here.

	delete(fvg.found, insn.Ident)
}

func gatherFreeVars(block *gcil.Block, trans *transformWithKFO) nameSet {
	v := &freeVarsGatherer{map[string]struct{}{}, trans}
	v.exploreBlock(block)
	return v.found
}

func gatherFreeVarsTillTheEnd(insn *gcil.Insn, trans *transformWithKFO) nameSet {
	v := &freeVarsGatherer{map[string]struct{}{}, trans}
	v.exploreTillTheEnd(insn)
	return v.found
}
