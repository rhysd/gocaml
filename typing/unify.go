package typing

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/ast"
)

// Check cyclic dependency. When unifying t and u where t is type variable and
// u is a type which contains t, it results in infinite-length type.
// It should be reported as semantic error.
func occur(v *ast.TypeVar, rhs ast.Type) bool {
	switch t := rhs.(type) {
	case *ast.TupleType:
		for _, e := range t.Elems {
			if occur(v, e) {
				return true
			}
		}
	case *ast.ArrayType:
		return occur(v, t.Elem)
	case *ast.FunType:
		if occur(v, t.Ret) {
			return true
		}
		for _, p := range t.Params {
			if occur(v, p) {
				return true
			}
		}
	case *ast.TypeVar:
		if t.Ref != nil {
			return occur(v, t.Ref)
		}
	}
	return false
}

func unifyTuple(left, right *ast.TupleType) error {
	length := len(left.Elems)
	if length != len(right.Elems) {
		return fmt.Errorf("Number of elements of tuple does not match between '%s' and '%s'", left.String(), right.String())
	}

	for i := 0; i < length; i++ {
		l := left.Elems[i]
		r := right.Elems[i]
		if err := Unify(l, r); err != nil {
			return errors.Wrap(err, fmt.Sprintf("On unifying tuples '%s' and '%s'", left.String(), right.String()))
		}
	}

	return nil
}

func unifyFun(left, right *ast.FunType) error {
	if err := Unify(left.Ret, right.Ret); err != nil {
		return errors.Wrap(err, fmt.Sprintf("On unifying functions' return types of '%s' and '%s'", left.String(), right.String()))
	}

	length := len(left.Params)
	if length != len(right.Params) {
		return fmt.Errorf("Number of parameters of function does not match between '%s' and '%s'", left.String(), right.String())
	}

	for i := 0; i < length; i++ {
		l := left.Params[i]
		r := right.Params[i]
		if err := Unify(l, r); err != nil {
			return errors.Wrap(err, fmt.Sprintf("On unifying function parameters of function '%s' and '%s'", left.String(), right.String()))
		}
	}

	return nil
}

func unifyTypeVar(l *ast.TypeVar, right ast.Type) error {
	switch r := right.(type) {
	case *ast.TypeVar:
		if l.Id == r.Id {
			return nil
		}
	}

	if l.Ref != nil {
		return Unify(l.Ref, right)
	}

	if occur(l, right) {
		return fmt.Errorf("Cyclic dependency found in types. Type variable '%s' is contained in '%s'")
	}

	// Assign rhs type to type variable when lhs type variable is unknown
	l.Ref = right
	return nil
}

func Unify(left, right ast.Type) error {
	switch l := left.(type) {
	case *ast.UnitType:
		if _, ok := right.(*ast.UnitType); ok {
			return nil
		}
	case *ast.BoolType:
		if _, ok := right.(*ast.BoolType); ok {
			return nil
		}
	case *ast.IntType:
		if _, ok := right.(*ast.IntType); ok {
			return nil
		}
	case *ast.FloatType:
		if _, ok := right.(*ast.FloatType); ok {
			return nil
		}
	case *ast.TupleType:
		switch r := right.(type) {
		case *ast.TupleType:
			return unifyTuple(l, r)
		}
	case *ast.ArrayType:
		switch r := right.(type) {
		case *ast.ArrayType:
			return Unify(l.Elem, r.Elem)
		}
	case *ast.FunType:
		switch r := right.(type) {
		case *ast.FunType:
			return unifyFun(l, r)
		}
	case *ast.TypeVar:
		return unifyTypeVar(l, right)
	}

	switch v := right.(type) {
	case *ast.TypeVar:
		return unifyTypeVar(v, left)
	}

	return fmt.Errorf("Cannot unify types. Type mismatch between '%s' and '%s'", left.String(), right.String())
}
