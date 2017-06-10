package closure

import (
	"fmt"
	"github.com/rhysd/gocaml/mir"
)

type freeVarsGatherer struct {
	found     nameSet
	transform *transformWithKFO
}

func (fvg *freeVarsGatherer) add(name string) {
	fvg.found[name] = struct{}{}
}

func (fvg *freeVarsGatherer) exploreBlock(block *mir.Block) {
	// Traverse instructions in the block in reverse order.
	// First and last instructions are NOP, so skipped.
	for i := block.Bottom.Prev; i.Prev != nil; i = i.Prev {
		fvg.exploreInsn(i)
	}
}

func (fvg *freeVarsGatherer) exploreTillTheEnd(insn *mir.Insn) {
	end := insn
	for end.Next.Next != nil {
		// Find the last instruction before NOP
		end = end.Next
	}
	for i := end; i != insn.Prev; i = i.Prev {
		fvg.exploreInsn(i)
	}
}

func (fvg *freeVarsGatherer) exploreInsn(insn *mir.Insn) {
	switch val := insn.Val.(type) {
	case *mir.Unary:
		fvg.add(val.Child)
	case *mir.Binary:
		fvg.add(val.LHS)
		fvg.add(val.RHS)
	case *mir.Ref:
		fvg.add(val.Ident)
	case *mir.If:
		fvg.add(val.Cond)
		fvg.exploreBlock(val.Then)
		fvg.exploreBlock(val.Else)
	case *mir.App:
		// Should not add val.Callee to free variables if it is not a closure
		// because a normal function is treated as label, not a variable
		// (label is a constant).
		// `_, ok := fvg.transform.closures[val.Callee]; ok` cannot be used
		// because callee may be a function variable, which also must be treated
		// as closure call.
		if _, ok := fvg.transform.knownFuns[val.Callee]; !ok && val.Kind != mir.EXTERNAL_CALL {
			fvg.add(val.Callee)
		}
		for _, a := range val.Args {
			fvg.add(a)
		}
	case *mir.Tuple:
		for _, e := range val.Elems {
			fvg.add(e)
		}
	case *mir.Array:
		fvg.add(val.Size)
		fvg.add(val.Elem)
	case *mir.ArrLit:
		for _, e := range val.Elems {
			fvg.add(e)
		}
	case *mir.TplLoad:
		fvg.add(val.From)
	case *mir.ArrLoad:
		fvg.add(val.From)
		fvg.add(val.Index)
	case *mir.ArrStore:
		fvg.add(val.To)
		fvg.add(val.Index)
		fvg.add(val.RHS)
	case *mir.ArrLen:
		fvg.add(val.Array)
	case *mir.Some:
		fvg.add(val.Elem)
	case *mir.IsSome:
		fvg.add(val.OptVal)
	case *mir.DerefSome:
		fvg.add(val.SomeVal)
	case *mir.Fun:
		make, ok := fvg.transform.replacedFuns[insn]
		if !ok {
			panic(fmt.Sprintf("Visiting function '%s' for gathering free vars is not visit by transformWithKFO: %v", insn.Ident, val))
		}
		if make == nil {
			// The function is not a closure. Need not to be visit because it is
			// simply moved to toplevel
			break
		}
		fv, ok := fvg.transform.closureBlockFreeVars[make.Fun]
		if !ok {
			panic(fmt.Sprintf("Applying unknown closure '%s'", insn.Ident))
		}
		for v := range fv {
			fvg.add(v)
		}
		for _, v := range make.Vars {
			fvg.add(v)
		}
		delete(fvg.found, make.Fun)
	case *mir.MakeCls:
		panic("unreachable")
	}

	// Note:
	// Functions in tree will be moved to toplevel. So they should be ignored here.

	delete(fvg.found, insn.Ident)
}

func gatherFreeVars(block *mir.Block, trans *transformWithKFO) nameSet {
	v := &freeVarsGatherer{map[string]struct{}{}, trans}
	v.exploreBlock(block)
	return v.found
}

func gatherFreeVarsTillTheEnd(insn *mir.Insn, trans *transformWithKFO) nameSet {
	v := &freeVarsGatherer{map[string]struct{}{}, trans}
	v.exploreTillTheEnd(insn)
	return v.found
}
