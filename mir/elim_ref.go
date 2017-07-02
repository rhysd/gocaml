package mir

import (
	"github.com/rhysd/gocaml/types"
)

type refEntry struct {
	insn  *Insn
	ident string
}

type elimRef struct {
	refs  map[string]refEntry
	xrefs map[string]refEntry
	types *types.Env
}

func newElimRef(types *types.Env) *elimRef {
	return &elimRef{
		map[string]refEntry{},
		map[string]refEntry{},
		types,
	}
}

func (elim *elimRef) elimRef(ident string) string {
	entry, ok := elim.refs[ident]
	if !ok {
		return ident
	}

	i := entry.insn
	i.RemoveFromList()
	delete(elim.types.DeclTable, i.Ident)
	return entry.ident
}

func (elim *elimRef) elimXRef(app *App) {
	entry, ok := elim.xrefs[app.Callee]
	if !ok {
		return
	}
	entry.insn.RemoveFromList()
	app.Callee = entry.ident
	app.Kind = EXTERNAL_CALL
}

func (elim *elimRef) insn(insn *Insn) {
	switch val := insn.Val.(type) {
	case *Ref:
		elim.refs[insn.Ident] = refEntry{insn, val.Ident}
	case *XRef:
		elim.xrefs[insn.Ident] = refEntry{insn, val.Ident}
	case *Unary:
		val.Child = elim.elimRef(val.Child)
	case *Binary:
		val.LHS = elim.elimRef(val.LHS)
		val.RHS = elim.elimRef(val.RHS)
	case *If:
		val.Cond = elim.elimRef(val.Cond)
		elim.block(val.Then)
		elim.block(val.Else)
	case *Fun:
		elim.block(val.Body)
	case *App:
		val.Callee = elim.elimRef(val.Callee)
		for i, a := range val.Args {
			val.Args[i] = elim.elimRef(a)
		}
		elim.elimXRef(val)
	case *Tuple:
		for i, e := range val.Elems {
			val.Elems[i] = elim.elimRef(e)
		}
	case *Array:
		val.Size = elim.elimRef(val.Size)
		val.Elem = elim.elimRef(val.Elem)
	case *ArrLit:
		for i, e := range val.Elems {
			val.Elems[i] = elim.elimRef(e)
		}
	case *TplLoad:
		val.From = elim.elimRef(val.From)
	case *ArrLoad:
		val.From = elim.elimRef(val.From)
		val.Index = elim.elimRef(val.Index)
	case *ArrStore:
		val.To = elim.elimRef(val.To)
		val.Index = elim.elimRef(val.Index)
		val.RHS = elim.elimRef(val.RHS)
	case *ArrLen:
		val.Array = elim.elimRef(val.Array)
	case *Some:
		val.Elem = elim.elimRef(val.Elem)
	case *IsSome:
		val.OptVal = elim.elimRef(val.OptVal)
	case *DerefSome:
		val.SomeVal = elim.elimRef(val.SomeVal)
	}
}

func (elim *elimRef) block(block *Block) {
	begin, end := block.WholeRange()
	for i := begin; i != end; i = i.Next {
		elim.insn(i)
	}
}

// Removes unnecessary 'ref' instruction.
// This optimization needs to be executed before closure transform because
// functions referenced by variable must be a closure. So it is important
// to remove unnecessary references to functions here.
func ElimRefs(b *Block, env *types.Env) {
	e := newElimRef(env)
	e.block(b)
}
