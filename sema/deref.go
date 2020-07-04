package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	. "github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
)

type typeVarDereferencer struct {
	err      *locerr.Error
	env      *Env
	inferred InferredTypes
	schemes  schemes
	insts    refInsts
}

func (d *typeVarDereferencer) unwrapVar(v *Var) (Type, bool) {
	if v.Ref != nil {
		return d.unwrap(v.Ref)
	}

	if v.IsGeneric() {
		// Note:
		// If `d.isGeneralized(v.ID)` is true here, it meands that this type variable will be instantiated later.
		// e.g.
		//   let o = None in o = Some 42; o = Some true
		// In this example, type of `o` and `None` is 'a and will be instantiated `int option` and
		// `bool option` later.
		return v, true
	}

	return nil, false
}

func (d *typeVarDereferencer) unwrapFun(fun *Fun) (Type, bool) {
	r, ok := d.unwrap(fun.Ret)
	if !ok {
		return nil, false
	}
	fun.Ret = r
	for i, param := range fun.Params {
		p, ok := d.unwrap(param)
		if !ok {
			return nil, false
		}
		fun.Params[i] = p
	}
	return fun, true
}

func (d *typeVarDereferencer) unwrap(target Type) (Type, bool) {
	switch t := target.(type) {
	case *Fun:
		return d.unwrapFun(t)
	case *Tuple:
		for i, elem := range t.Elems {
			e, ok := d.unwrap(elem)
			if !ok {
				return nil, false
			}
			t.Elems[i] = e
		}
	case *Array:
		e, ok := d.unwrap(t.Elem)
		if !ok {
			return nil, false
		}
		t.Elem = e
	case *Option:
		e, ok := d.unwrap(t.Elem)
		if !ok {
			return nil, false
		}
		t.Elem = e
	case *Var:
		return d.unwrapVar(t)
	}
	return target, true
}

func (d *typeVarDereferencer) errIn(node ast.Expr, msg string) {
	if d.err == nil {
		d.err = locerr.ErrorIn(node.Pos(), node.End(), msg)
	} else {
		d.err = d.err.NoteAt(node.Pos(), msg)
	}
}

func (d *typeVarDereferencer) errMsg(msg string) {
	if d.err == nil {
		d.err = locerr.NewError(msg)
	} else {
		d.err = d.err.Note(msg)
	}
}

func (d *typeVarDereferencer) derefSym(node ast.Expr, sym *ast.Symbol) {
	symType, ok := d.env.DeclTable[sym.Name]
	if !ok {
		panic("FATAL: Cannot dereference unknown symbol: " + sym.Name)
	}

	if sym.IsIgnored() {
		// Parser expands `foo; bar` to `let $unused = foo in bar`. In this situation, type of the
		// variable will never be determined because it's unused.
		// So skipping it in order to avoid unknown type error for the unused variable.
		if v, ok := symType.(*Var); ok && v.Ref == nil && !v.IsGeneric() {
			// $unused variables are never be used. So its type may not be determined. In the case,
			// it's type should be fixed to unit type.
			v.Ref = UnitType
		}
		return
	}

	t, ok := d.unwrap(symType)
	if !ok {
		msg := fmt.Sprintf("Cannot infer type of variable '%s'. Inferred type was '%s'", sym.DisplayName, symType.String())
		d.errIn(node, msg)
		return
	}

	// Also dereference type variable in symbol
	d.env.DeclTable[sym.Name] = t
}

func (d *typeVarDereferencer) VisitTopdown(node ast.Expr) ast.Visitor {
	switch n := node.(type) {
	case *ast.Let:
		d.derefSym(n, n.Symbol)
	case *ast.LetRec:
		// Note:
		// Need to dereference parameters at first because type of the function depends on type
		// of its parameters and parameters may be specified as '_'. '_' is unused. So its type
		// may not be determined and need to be fixed as unit type.
		for _, p := range n.Func.Params {
			d.derefSym(n, p.Ident)
		}
		d.derefSym(n, n.Func.Symbol)
	case *ast.LetTuple:
		for _, sym := range n.Symbols {
			d.derefSym(n, sym)
		}
	case *ast.Match:
		d.derefSym(n, n.SomeIdent)
	case *ast.VarRef:
		if inst, ok := d.insts[n]; ok {
			unwrapped, ok := d.unwrap(inst.To)
			if !ok {
				msg := fmt.Sprintf("Cannot instantiate declaration '%s' typed as type '%s'", n.Symbol.DisplayName, inst.From.String())
				d.errIn(n, msg)
				d.err = d.err.NotefAt(n.Pos(), "Tried to instantiate the generic type as '%s'", inst.To.String())
				return nil
			}
			inst.To = unwrapped
			for _, m := range inst.Mapping {
				t, ok := d.unwrap(m.Type)
				if !ok {
					msg := fmt.Sprintf("Cannot instantiate type variable in generic type '%s' at declaration '%s'", inst.From.String(), n.Symbol.DisplayName)
					d.errIn(n, msg)
					return nil
				}
				m.Type = t
			}
		}
	}
	return d
}

func (d *typeVarDereferencer) checkLess(op string, lhs ast.Expr) string {
	operand, ok := d.inferred[lhs]
	if !ok {
		panic("FATAL: Operand type of operator '" + op + "' not found at " + lhs.Pos().String())
	}
	// Note:
	// This type constraint may be useful for type inference. But current HM type inference algorithm cannot
	// handle a union type. In this context, the operand should be `int | float`
	switch operand.(type) {
	case *Unit, *Bool, *String, *Fun, *Tuple, *Array, *Option:
		return fmt.Sprintf("'%s' can't be compared with operator '%s'", operand.String(), op)
	default:
		return ""
	}
}

func (d *typeVarDereferencer) checkEq(op string, lhs ast.Expr) string {
	operand, ok := d.inferred[lhs]
	if !ok {
		panic("FATAL: Operand type of operator '" + op + "' not found at " + lhs.Pos().String())
	}
	// Note:
	// This type constraint may be useful for type inference. But current HM type inference algorithm cannot
	// handle a union type. In this context, the operand should be `() | bool | int | float | fun<R, TS...> | tuple<Args...>`
	if a, ok := operand.(*Array); ok {
		return fmt.Sprintf("Array type '%s' can't be compared with operator '%s'", a.String(), op)
	}
	return ""
}

func (d *typeVarDereferencer) miscCheck(node ast.Expr) {
	msg := ""
	switch n := node.(type) {
	case *ast.Less:
		msg = d.checkLess("<", n.Left)
	case *ast.LessEq:
		msg = d.checkLess("<=", n.Left)
	case *ast.Greater:
		msg = d.checkLess(">", n.Left)
	case *ast.GreaterEq:
		msg = d.checkLess(">=", n.Left)
	case *ast.Eq:
		msg = d.checkEq("=", n.Left)
	case *ast.NotEq:
		msg = d.checkEq("<>", n.Left)
	}
	if msg != "" {
		d.errIn(node, msg)
	}
}

func (d *typeVarDereferencer) VisitBottomup(node ast.Expr) {
	d.miscCheck(node)

	// Dereference all nodes' types
	t, ok := d.inferred[node]
	if !ok {
		return
	}

	unwrapped, ok := d.unwrap(t)
	if !ok {
		msg := fmt.Sprintf("Cannot infer type of expression. Type annotation is needed. Inferred type was '%s'", t.String())
		d.errIn(node, msg)
		return
	}

	d.inferred[node] = unwrapped
}

func (d *typeVarDereferencer) normalizePolyTypes() {
	polys := make(map[Type][]*Instantiation, len(d.schemes))
	for t := range d.schemes {
		polys[t] = make([]*Instantiation, 0, 3)
	}
RefLoop:
	for _, inst := range d.insts {
		insts := polys[inst.From]
		for _, i := range insts {
			if Equals(i.To, inst.To) {
				inst.To = i.To
				inst.Mapping = i.Mapping
				continue RefLoop
			}
		}
		polys[inst.From] = append(insts, inst)
	}
	d.env.PolyTypes = polys
}

func derefTypeVars(env *Env, root ast.Expr, inferred InferredTypes, ss schemes, insts map[*ast.VarRef]*Instantiation) *locerr.Error {
	deref := &typeVarDereferencer{nil, env, inferred, ss, insts}

	// Note:
	// Don't need to dereference types of external symbols because they must not contain any
	// type variables.
	ast.Visit(deref, root)

	// Note:
	// Cannot return v.err directly because `return v.err` returns typed nil (typed as *locerr.Error).
	if deref.err != nil {
		return deref.err
	}

	deref.normalizePolyTypes()

	return nil
}
