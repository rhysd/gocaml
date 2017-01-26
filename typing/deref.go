package typing

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"strings"
)

type typeVarDereferencer struct {
	errors []string
	env    *Env
}

func (d *typeVarDereferencer) unwrapTypeVar(variable *ast.TypeVar) (ast.Type, bool) {
	if variable.Ref != nil {
		r, ok := d.unwrap(variable.Ref)
		if !ok {
			return nil, false
		}
		return r, true
	}

	// XXX: Different behavior from MinCaml.
	//
	// In MinCaml, unknown type value will be fallbacked into Int.
	// But GoCaml decided to fallback unit type.
	//
	//   1. When type variable is empty (e.g. not $1(unknown list), but $1(unknown))
	//   2. When the type variable appears in return type of external function symbol.
	//
	// For example, `print_int 42; ()` causes a type error such as 'type of $tmp1 is unknown'
	// This is because it will be transformed to `let $tmp1 = print_int 42 in ()` and return
	// type of external function `print_int` is unknown.
	// To avoid kinds of this error, GoCaml decided to assign `()` to the return type.
	// Then $tmp can be inferred as `()`. $tmp1 is always unused variable. So it doesn't
	// cause any problem, I believe.
	//
	// (Test case: testdata/basic/external_func_unknown_ret_type.ml)
	for _, t := range d.env.Externals {
		if v, ok := t.(*ast.TypeVar); ok && v.Ref != nil {
			if f, ok := v.Ref.(*ast.FunType); ok && f.Ret == variable {
				f.Ret = ast.UnitTypeVal
				return ast.UnitTypeVal, true
			}
		}
	}

	return nil, false
}

func (d *typeVarDereferencer) unwrap(target ast.Type) (ast.Type, bool) {
	switch t := target.(type) {
	case *ast.FunType:
		r, ok := d.unwrap(t.Ret)
		if !ok {
			return nil, false
		}
		t.Ret = r
		for i, param := range t.Params {
			p, ok := d.unwrap(param)
			if !ok {
				return nil, false
			}
			t.Params[i] = p
		}
	case *ast.TupleType:
		for i, elem := range t.Elems {
			e, ok := d.unwrap(elem)
			if !ok {
				return nil, false
			}
			t.Elems[i] = e
		}
	case *ast.ArrayType:
		e, ok := d.unwrap(t.Elem)
		if !ok {
			return nil, false
		}
		t.Elem = e
	case *ast.TypeVar:
		return d.unwrapTypeVar(t)
	}
	return target, true
}

func (d *typeVarDereferencer) derefSym(node ast.Expr, sym *ast.Symbol) {
	t, ok := d.unwrap(sym.Type)
	if !ok {
		pos := node.Pos()
		d.errors = append(d.errors, fmt.Sprintf("Cannot infer type of variable '%s' in node %s (line:%d, column:%d). Inferred type was '%s'", sym.Name, node.Name(), pos.Line, pos.Column, sym.Type.String()))
		return
	}
	sym.Type = t
}

func (d *typeVarDereferencer) Visit(node ast.Expr) ast.Visitor {
	switch n := node.(type) {
	case *ast.Let:
		d.derefSym(n, n.Symbol)
	case *ast.LetRec:
		d.derefSym(n, n.Func.Symbol)
		for _, sym := range n.Func.Params {
			d.derefSym(n, sym)
		}
	case *ast.LetTuple:
		for _, sym := range n.Symbols {
			d.derefSym(n, sym)
		}
	}
	return d
}

func (env *Env) DerefTypeVars(root ast.Expr) error {
	v := &typeVarDereferencer{[]string{}, env}
	ast.Visit(v, root)
	if len(v.errors) > 0 {
		return fmt.Errorf("Error while type inference (dereferencing type vars)\n%s", strings.Join(v.errors, "\n"))
	}
	return nil
}
