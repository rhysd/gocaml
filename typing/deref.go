package typing

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/loc"
)

func unwrapVar(variable *Var) (Type, bool) {
	if variable.Ref != nil {
		r, ok := unwrap(variable.Ref)
		if ok {
			return r, true
		}
	}

	return nil, false
}

func unwrapFun(fun *Fun) (Type, bool) {
	r, ok := unwrap(fun.Ret)
	if !ok {
		return nil, false
	}
	fun.Ret = r
	for i, param := range fun.Params {
		p, ok := unwrap(param)
		if !ok {
			return nil, false
		}
		fun.Params[i] = p
	}
	return fun, true
}

func unwrap(target Type) (Type, bool) {
	switch t := target.(type) {
	case *Fun:
		return unwrapFun(t)
	case *Tuple:
		for i, elem := range t.Elems {
			e, ok := unwrap(elem)
			if !ok {
				return nil, false
			}
			t.Elems[i] = e
		}
	case *Array:
		e, ok := unwrap(t.Elem)
		if !ok {
			return nil, false
		}
		t.Elem = e
	case *Option:
		e, ok := unwrap(t.Elem)
		if !ok {
			return nil, false
		}
		t.Elem = e
	case *Var:
		return unwrapVar(t)
	}
	return target, true
}

type typeVarDereferencer struct {
	err *loc.Error
	env *Env
}

func (d *typeVarDereferencer) derefSym(node ast.Expr, sym *ast.Symbol) {
	symType, ok := d.env.Table[sym.Name]

	if sym.IsIgnored() {
		// Parser expands `foo; bar` to `let $unused = foo in bar`. In this situation, type of the
		// variable will never be determined because it's unused.
		// So skipping it in order to avoid unknown type error for the unused variable.
		if v, ok := symType.(*Var); ok {
			// $unused variables are never be used. So its type may not be determined. In the case,
			// it's type should be fixed to unit type.
			v.Ref = UnitType
		}
		return
	}

	if !ok {
		panic(fmt.Sprintf("FATAL: Cannot dereference unknown symbol '%s'", sym.Name))
		return
	}

	t, ok := unwrap(symType)
	if !ok {
		msg := fmt.Sprintf("Cannot infer type of variable '%s'. Inferred type was '%s'", sym.DisplayName, symType.String())
		if d.err == nil {
			d.err = loc.ErrorAt(node.Pos(), msg)
		} else {
			d.err = d.err.NoteAt(node.Pos(), msg)
		}
		return
	}

	// Also dereference type variable in symbol
	d.env.Table[sym.Name] = t
}

// XXX: Different behavior from MinCaml.
//
// In MinCaml, unknown type value will be fallbacked into Int.
// But GoCaml decided to fallback unit type.
//
//   1. When type variable is empty
//   2. When the type variable appears in return type of external function symbol.
//
// For example, `print 42; ()` causes a type error such as 'type of $tmp1 is unknown'.
// This is because it will be transformed to `let $tmp1 = print 42 in ()` and return
// type of external function `print` is unknown.
// To avoid kinds of this error, GoCaml decided to assign `()` to the return type.
// Then $tmp can be inferred as `()`. $tmp1 is always unused variable. So it doesn't
// cause any problem, I believe.
//
// (Test case: testdata/basic/external_func_unknown_ret_type.ml)
func (d *typeVarDereferencer) fixExternalFuncRet(ret Type) Type {
	for {
		v, ok := ret.(*Var)
		if !ok {
			return ret
		}
		if v.Ref == nil {
			return UnitType
		}
		ret = v.Ref
	}
}

func (d *typeVarDereferencer) externalSymError(n string, t Type) {
	msg := fmt.Sprintf("Cannot infer type of external symbol '%s'. Note: Inferred as '%s'", n, t.String())
	if d.err == nil {
		d.err = loc.NewError(msg)
		return
	}
	d.err = d.err.Note(msg)
}

func (d *typeVarDereferencer) derefExternalSym(name string, symType Type) Type {
	switch ty := symType.(type) {
	case *Var:
		// Unwrap type variables: $($($(t))) -> t
		if ty.Ref == nil {
			d.externalSymError(name, symType)
			return symType
		}
		return d.derefExternalSym(name, ty.Ref)
	case *Fun:
		ty.Ret = d.fixExternalFuncRet(ty.Ret)
		t, ok := unwrapFun(ty)
		if !ok {
			d.externalSymError(name, symType)
			return ty
		}
		return t
	default:
		t, ok := unwrap(symType)
		if !ok {
			d.externalSymError(name, symType)
			return symType
		}
		return t
	}
}

func (d *typeVarDereferencer) Visit(node ast.Expr) ast.Visitor {
	switch n := node.(type) {
	case *ast.Let:
		d.derefSym(n, n.Symbol)
	case *ast.LetRec:
		// Note:
		// Need to dereference parameters at first because type of the function depends on type
		// of its parameters and parameters may be specified as '_'.
		// '_' is unused. So its type may not be detemined and need to be fixed as unit type.
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
	}
	return d
}

func derefTypeVars(env *Env, root ast.Expr) error {
	v := &typeVarDereferencer{nil, env}
	for n, t := range env.Externals {
		env.Externals[n] = v.derefExternalSym(n, t)
	}
	ast.Visit(v, root)

	if v.err != nil {
		return v.err
	}

	for n, t := range env.NoneTypes {
		deref, ok := unwrap(t.Elem)
		if !ok {
			p := n.Pos()
			panic(fmt.Sprintf("FATAL: Dereferencing type of 'None' value must not fail. Value at (line:%d, col:%d) was wrongly typed as %s", p.Line, p.Column, t.String()))
		}
		t.Elem = deref
	}

	return nil
}
