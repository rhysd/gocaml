package gcil

import (
	"github.com/rhysd/gocaml/typing"
)

type refEntry struct {
	insn  *Insn
	ident string
}

type elimRefVisitor struct {
	refs  map[string]refEntry
	xrefs map[string]refEntry
	types *typing.Env
}

func newElimRefVisitor(types *typing.Env) *elimRefVisitor {
	return &elimRefVisitor{
		map[string]refEntry{},
		map[string]refEntry{},
		types,
	}
}

func (vis *elimRefVisitor) elimRef(ident string) string {
	entry, ok := vis.refs[ident]
	if !ok {
		return ident
	}

	i := entry.insn
	i.RemoveFromList()
	delete(vis.types.Table, i.Ident)
	return entry.ident
}

func (vis *elimRefVisitor) elimXRef(app *App) {
	entry, ok := vis.xrefs[app.Callee]
	if !ok {
		return
	}
	entry.insn.RemoveFromList()
	app.Callee = entry.ident
	app.Kind = EXTERNAL_CALL
}

func (vis *elimRefVisitor) Visit(insn *Insn) Visitor {
	switch val := insn.Val.(type) {
	case *Ref:
		vis.refs[insn.Ident] = refEntry{insn, val.Ident}
	case *XRef:
		vis.xrefs[insn.Ident] = refEntry{insn, val.Ident}
	case *Unary:
		val.Child = vis.elimRef(val.Child)
	case *Binary:
		val.Lhs = vis.elimRef(val.Lhs)
		val.Rhs = vis.elimRef(val.Rhs)
	case *If:
		val.Cond = vis.elimRef(val.Cond)
		Visit(vis, val.Then)
		Visit(vis, val.Else)
	case *Fun:
		Visit(vis, val.Body)
	case *App:
		val.Callee = vis.elimRef(val.Callee)
		for i, a := range val.Args {
			val.Args[i] = vis.elimRef(a)
		}
		vis.elimXRef(val)
	case *Tuple:
		for i, e := range val.Elems {
			val.Elems[i] = vis.elimRef(e)
		}
	case *Array:
		val.Size = vis.elimRef(val.Size)
		val.Elem = vis.elimRef(val.Elem)
	case *TplLoad:
		val.From = vis.elimRef(val.From)
	case *ArrLoad:
		val.From = vis.elimRef(val.From)
		val.Index = vis.elimRef(val.Index)
	case *ArrStore:
		val.To = vis.elimRef(val.To)
		val.Index = vis.elimRef(val.Index)
		val.Rhs = vis.elimRef(val.Rhs)
	}

	return vis
}

// Removes unnecessary 'ref' instruction.
// This optimization needs to be executed before closure transform because
// functions referenced by variable must be a closure. So it is important
// to remove unnecessary references to functions here.
func ElimRefs(b *Block, env *typing.Env) {
	e := newElimRefVisitor(env)
	Visit(e, b)
}
