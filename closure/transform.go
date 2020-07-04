// Package closure provides closure transform for MIR representation.
//
// Closure transform is a process to move all functions to toplevel of program.
// If a function does not contain any free variables, it can be moved to toplevel simply.
// But when containing any free variables, the function must take a closure struct as
// hidden parameter. And need to insert a code to make a closure at the definition
// point of the function.
//
// In closure transform, it visits function's body assuming the function is a normal function.
// As the result of the visit, if some free variables found, it means that the function
// is actually not a normal function, but a closure. So restore the state and retry
// visiting its body after adding the function to closures list.
//
// Note that applied normal functions are not free variables, but applied closures are
// free variables. Normal function is not a value but closure is a value.
// So, considering recursive functions, before visiting function's body, the function must
// be determined to normal function or closure. That's the reason to assume function is a
// normal function at first and then backtrack after if needed.
//
package closure

import (
	"fmt"
	"github.com/rhysd/gocaml/mir"
	"sort"
)

type nameSet map[string]struct{}

func (set nameSet) toSortedArray() []string {
	ns := make([]string, 0, len(set))
	for n := range set {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	return ns
}

// Do closure transform with known functions optimization
type transformWithKFO struct {
	knownFuns            nameSet
	replacedFuns         map[*mir.Insn]*mir.MakeCls // nil means simply removing the function
	closures             mir.Closures               // Mapping function name to free variables
	closureBlockFreeVars map[string]nameSet         // Known free variables of closures' blocks
}

func (trans *transformWithKFO) duplicate() *transformWithKFO {
	known := make(map[string]struct{}, len(trans.knownFuns))
	for k := range trans.knownFuns {
		known[k] = struct{}{}
	}
	funs := make(map[*mir.Insn]*mir.MakeCls, len(trans.replacedFuns))
	for f, v := range trans.replacedFuns {
		funs[f] = v
	}
	clss := make(map[string][]string, len(trans.closures))
	for f, fv := range trans.closures {
		clss[f] = fv
	}
	blks := make(map[string]nameSet, len(trans.closureBlockFreeVars))
	for f, fv := range trans.closureBlockFreeVars {
		blks[f] = fv
	}
	return &transformWithKFO{
		known,
		funs,
		clss,
		blks,
	}
}

func (trans *transformWithKFO) block(block *mir.Block) {
	// Skip first NOP instruction
	trans.insn(block.Top.Next)
}

func (trans *transformWithKFO) insn(insn *mir.Insn) {
	if insn.Next == nil {
		// Reaches bottom of the block
		return
	}

	switch val := insn.Val.(type) {
	case *mir.Fun:
		// Assume the function is not a closure and try to transform its body
		dup := trans.duplicate()
		dup.knownFuns[insn.Ident] = struct{}{}
		dup.block(val.Body)
		// Check there is no free variable actually
		fv := gatherFreeVars(val.Body, dup)
		for _, p := range val.Params {
			delete(fv, p)
		}
		if len(fv) != 0 {
			// Assumed the function is not a closure. But there are actually some
			// free variables. It means that the function is actually a closure.
			// Discard 'dup' and retry visiting its body with adding it to closures.
			trans.block(val.Body)
			fv = gatherFreeVars(val.Body, trans)
			for _, p := range val.Params {
				delete(fv, p)
			}
			if _, ok := fv[insn.Ident]; ok {
				// When the closure itself is used in its body (recursive function), it must prepare
				// the closure object in its body to use itself in its body.
				val.IsRecursive = true
				delete(fv, insn.Ident)
			}
			trans.closures[insn.Ident] = fv.toSortedArray()
		} else {
			// When the function is actually not a closure, continue to use 'dup' as current visitor
			*trans = *dup
		}

		// Visit recursively
		trans.insn(insn.Next)

		// Visit rest block of the 'fun' instruction
		if cache, ok := trans.closureBlockFreeVars[insn.Ident]; ok {
			fv = cache
		} else {
			fv = gatherFreeVarsTillTheEnd(insn.Next, trans)
		}
		trans.closureBlockFreeVars[insn.Ident] = fv

		var replaced *mir.MakeCls
		if _, ok := fv[insn.Ident]; ok {
			vars, ok := trans.closures[insn.Ident]
			if !ok {
				// When the function is used as a variable, it must have an empty
				// closure even if there is no free variable for the function.
				// It's because we can't know a passed function variable is a closure or not.
				vars = []string{}
				trans.closures[insn.Ident] = vars
				delete(trans.knownFuns, insn.Ident)
			}
			// If the function is referred from somewhere, we need to  make a closure.
			replaced = &mir.MakeCls{vars, insn.Ident}
		}
		trans.replacedFuns[insn] = replaced
	case *mir.If:
		trans.block(val.Then)
		trans.block(val.Else)
		trans.insn(insn.Next)
	default:
		trans.insn(insn.Next)
	}
}

// Transform executes closure transform.
// The result is a representation of the program. It contains toplevel functions,
// entry point and closure information.
// All nested function was moved to toplevel.
func Transform(ir *mir.Block) *mir.Program {
	t := &transformWithKFO{
		map[string]struct{}{},
		map[*mir.Insn]*mir.MakeCls{},
		map[string][]string{},
		map[string]nameSet{},
	}
	t.block(ir)

	// Move all functions to toplevel and put closure instance if needed
	toplevel := mir.NewToplevel()
	for insn, make := range t.replacedFuns {
		f, ok := insn.Val.(*mir.Fun)
		if !ok {
			panic(fmt.Sprintf("Replaced function '%s' is actually not a function: %v", insn.Ident, insn.Val))
		}
		toplevel.Add(insn.Ident, f, insn.Pos)

		if make == nil {
			// It's not a closure. Simply remove 'fun' instruction from list
			insn.RemoveFromList()
		} else {
			// Replace 'fun' with 'makecls' to make a closure instead of defining the function
			insn.Val = make
		}
	}

	prog := &mir.Program{toplevel, t.closures, ir}
	fixAppsInProg(prog)
	return prog
}
