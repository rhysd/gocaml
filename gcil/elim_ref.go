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
	types *typing.Env
}

func newElimRefVisitor(types *typing.Env) *elimRefVisitor {
	return &elimRefVisitor{map[string]refEntry{}, types}
}

func (vis *elimRefVisitor) elimRef(ident string) string {
	if entry, ok := vis.refs[ident]; ok {
		i := entry.insn
		i.Prev.Next = i.Next
		i.Next.Prev = i.Prev
		delete(vis.types.Table, i.Ident)
		return entry.ident
	}
	return ident
}

func (vis *elimRefVisitor) Visit(insn *Insn) Visitor {
	if ref, ok := insn.Val.(*Ref); ok {
		vis.refs[insn.Ident] = refEntry{insn, ref.Ident}
		return vis
	}

	switch val := insn.Val.(type) {
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
