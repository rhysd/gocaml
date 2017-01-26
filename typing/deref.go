package typing

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"strings"
)

func stripTypeVar(target ast.Type) (ast.Type, bool) {
	switch t := target.(type) {
	case *ast.FunType:
		r, ok := stripTypeVar(t.Ret)
		if !ok {
			return nil, false
		}
		t.Ret = r
		for i, param := range t.Params {
			p, ok := stripTypeVar(param)
			if !ok {
				return nil, false
			}
			t.Params[i] = p
		}
	case *ast.TupleType:
		for i, elem := range t.Elems {
			e, ok := stripTypeVar(elem)
			if !ok {
				return nil, false
			}
			t.Elems[i] = e
		}
	case *ast.ArrayType:
		e, ok := stripTypeVar(t.Elem)
		if !ok {
			return nil, false
		}
		t.Elem = e
	case *ast.TypeVar:
		if t.Ref == nil {
			return nil, false
		}
		// Dereference type variable
		r, ok := stripTypeVar(t.Ref)
		if !ok {
			return nil, false
		}
		return r, true
	}
	return target, true
}

type typeVarDereferencer struct {
	errors []string
}

func (d *typeVarDereferencer) derefSym(node ast.Expr, sym *ast.Symbol) {
	t, ok := stripTypeVar(sym.Type)
	if !ok {
		pos := node.Pos()
		d.errors = append(d.errors, fmt.Sprintf("Cannot infer type of variable '%s' in node %s (line:%d, column:%d). Infered type was '%s'", sym.Name, node.Name(), pos.Line, pos.Column, sym.Type.String()))
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

func DerefTypeVars(root ast.Expr) error {
	v := &typeVarDereferencer{
		errors: []string{},
	}
	ast.Visit(v, root)
	if len(v.errors) > 0 {
		return fmt.Errorf("Error while type inference (dereferencing type vars)\n%s", strings.Join(v.errors, "\n"))
	}
	return nil
}
