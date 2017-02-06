package closure

import (
	"fmt"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/typing"
)

type freeVars struct {
	names       []string
	appearInRef bool
}

func newFreeVars() *freeVars {
	return &freeVars{[]string{}, false}
}

type freeVarAnalysis struct {
	funcs   map[string]freeVars
	current map[string]struct{}
	refs    map[string]struct{}
	types   *typing.Env
}

func newFreeVarAnalysis(types *typing.Env) *freeVarAnalysis {
	return &freeVarAnalysis{
		map[string]freeVars{},
		map[string]struct{}{},
		map[string]struct{}{},
		types,
	}
}

func (fva *freeVarAnalysis) add(name string) {
	fva.current[name] = struct{}{}
}

func (fva *freeVarAnalysis) saveFreeVars(ident string) {
	delete(fva.current, ident)
	names := make([]string, 0, len(fva.current))
	for name := range fva.current {
		names = append(names, name)
	}
	_, appearInBody := fva.refs[ident]
	fva.funcs[ident] = freeVars{names, appearInBody}
}

func (fva *freeVarAnalysis) analyzeBlock(b *gcil.Block) {
	// Traverse instructions in reverse order.
	// First and last instructions are NOP, so skipped.
	for i := b.Bottom.Prev; i.Prev != nil; i = i.Prev {
		fva.analyzeInsn(i)
	}
}

func (fva *freeVarAnalysis) analyzeInsn(insn *gcil.Insn) {
	switch val := insn.Val.(type) {
	case *gcil.Unary:
		fva.add(val.Child)
	case *gcil.Binary:
		fva.add(val.Lhs)
		fva.add(val.Rhs)
	case *gcil.Ref:
		fva.add(val.Ident)
		if t, ok := fva.types.Table[val.Ident]; ok {
			if _, ok := t.(*typing.Fun); ok {
				fva.refs[val.Ident] = struct{}{}
			}
		}
	case *gcil.If:
		fva.add(val.Cond)
		fva.analyzeBlock(val.Then)
		fva.analyzeBlock(val.Else)
	case *gcil.Fun:
		fva.analyzeBlock(val.Body)
		for _, p := range val.Params {
			delete(fva.current, p)
		}
		fva.saveFreeVars(insn.Ident)
	case *gcil.App:
		fva.add(val.Callee)
		for _, a := range val.Args {
			fva.add(a)
		}
	case *gcil.Tuple:
		for _, e := range val.Elems {
			fva.add(e)
		}
	case *gcil.Array:
		fva.add(val.Size)
		fva.add(val.Elem)
	case *gcil.TplLoad:
		fva.add(val.From)
	case *gcil.ArrLoad:
		fva.add(val.From)
		fva.add(val.Index)
	case *gcil.ArrStore:
		fva.add(val.To)
		fva.add(val.Index)
		fva.add(val.Rhs)
	}

	delete(fva.current, insn.Ident)
}

type freeVarTransformer struct {
	funcs    map[string]freeVars
	closures map[string]string // function variable name -> its closure varible name
	root     *gcil.Block
}

func (fvt *freeVarTransformer) genClosureId(fun string) string {
	id := fmt.Sprintf("$c$%s", fun)
	fvt.closures[fun] = id
	return id
}

func (fvt *freeVarTransformer) Visit(insn *gcil.Insn) gcil.Visitor {
	switch val := insn.Val.(type) {
	case *gcil.Fun:
		// Move 'fun' instruction into toplevel
		prev := insn.Prev
		insn.RemoveFromList()
		fvt.root.Prepend(insn)
		fv, ok := fvt.funcs[insn.Ident]
		if !ok {
			panic(fmt.Sprintf("Function '%s' was not analyzed (%v)", insn.Ident, fvt.funcs))
		}
		if len(fv.names) == 0 && !fv.appearInRef {
			// When a function contains no free var and it does not appear in 'ref'
			// instruction in its body, it can be moved to toplevel straightforwardly.
			break
		}
		id := fvt.genClosureId(insn.Ident)
		make := gcil.NewInsn(id, &gcil.MakeCls{fv.names, insn.Ident})
		// Insert MakeCls instruction
		make.Prev = prev
		make.Next = prev.Next
		make.Prev.Next = make
		make.Next.Prev = make
	case *gcil.App:
		cls, found := fvt.closures[val.Callee]
		if !found {
			break
		}
		// Promote 'app' to 'appcls'
		app := gcil.NewInsn(insn.Ident, &gcil.AppCls{val.Callee, val.Args, cls})
		// Replace 'app' with promoted 'appcls' in instruction sequence
		app.Prev = insn.Prev
		app.Next = insn.Next
		insn.Prev.Next = app
		insn.Next.Prev = app
	}
	return fvt
}

func Transform(root *gcil.Block, types *typing.Env) {
	// First pass
	firstPass := newFreeVarAnalysis(types)
	firstPass.analyzeBlock(root)

	// Second pass
	secondPass := &freeVarTransformer{firstPass.funcs, map[string]string{}, root}
	gcil.Visit(secondPass, root)
}
