// Package closure provides closure transform for GCIL representation.
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
package closure

import (
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/typing"
)

type nameSet map[string]struct{}

// Do closure transform with known functions optimization
type transformWithKFO struct {
	replacedFuns map[string]*gcil.MakeCls // nil means it's a known normal function
	promotedApps []*gcil.App
	closures map[string][]string // Mapping function name to free variables
	closureBlockFreeVars map[string]nameSet // Known free variables of closures' blocks
}

func (trans *transformWithKFO) dup() *transformWithKFO{
	funs := make(map[string]*gcil.MakeCls, len(trans.replacedFuns))
	for f,v := range trans.replacedFuns {
		funs[f] = v
	}
	apps := make([]*gcil.App, 0, len(trans.promotedApps))
	for _,a := range trans.promotedApps {
		apps = append(apps, a)
	}
	clss := make(map[string][]string, len(trans.closures))
	for f,fv := range trans.closures {
		clss[f] = fv
	}
	blks := make(map[string]nameSet, len(trans.closureBlockFreeVars))
	for f,fv := range trans.closureBlockFreeVars {
		blks[f] = fv
	}
	return &transformWithKFO {funs, apps, clss, blks}
}

func Transform(ir *gcil.Block, env *typing.Env) {
	// TODO
}
