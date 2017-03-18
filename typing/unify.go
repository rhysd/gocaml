package typing

import (
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/common"
)

type UnificationError struct {
}

// Check cyclic dependency. When unifying t and u where t is type variable and
// u is a type which contains t, it results in infinite-length type.
// It should be reported as semantic error.
func occur(v *Var, rhs Type) bool {
	switch t := rhs.(type) {
	case *Tuple:
		for _, e := range t.Elems {
			if occur(v, e) {
				return true
			}
		}
	case *Array:
		return occur(v, t.Elem)
	case *Fun:
		if occur(v, t.Ret) {
			return true
		}
		for _, p := range t.Params {
			if occur(v, p) {
				return true
			}
		}
	case *Var:
		if v == t {
			return true
		}
		if t.Ref != nil {
			return occur(v, t.Ref)
		}
	}
	return false
}

func unifyTuple(left, right *Tuple) error {
	length := len(left.Elems)
	if length != len(right.Elems) {
		return errors.Errorf("Number of elements of tuple does not match: %d v.s. %d (between '%s' and '%s')", length, len(right.Elems), left.String(), right.String())
	}

	for i := 0; i < length; i++ {
		l := left.Elems[i]
		r := right.Elems[i]
		if err := Unify(l, r); err != nil {
			return errors.Wrapf(err, "On unifying tuples' %s elements of '%s' and '%s'\n", common.Ordinal(i+1), left.String(), right.String())
		}
	}

	return nil
}

func unifyFun(left, right *Fun) error {
	if err := Unify(left.Ret, right.Ret); err != nil {
		return errors.Wrapf(err, "On unifying functions' return types of '%s' and '%s'\n", left.String(), right.String())
	}

	if len(left.Params) != len(right.Params) {
		return errors.Errorf("Number of parameters of function does not match: %d v.s. %d (between '%s' and '%s')", len(left.Params), len(right.Params), left.String(), right.String())
	}

	for i, l := range left.Params {
		r := right.Params[i]
		if err := Unify(l, r); err != nil {
			return errors.Wrapf(err, "On unifying %s parameter of function '%s' and '%s'\n", common.Ordinal(i+1), left.String(), right.String())
		}
	}

	return nil
}

func unifyVar(l *Var, right Type) error {
	if l == right {
		return nil
	}

	if l.Ref != nil {
		return Unify(l.Ref, right)
	}

	if occur(l, right) {
		return errors.Errorf("Cannot resolve uninstantiated type variable. Cyclic dependency found while unification with '%s'", right.String())
	}

	// Assign rhs type to type variable when lhs type variable is unknown
	l.Ref = right
	return nil
}

func Unify(left, right Type) error {
	switch l := left.(type) {
	case *Unit, *Bool, *Int, *Float, *String:
		// Types for Unit, Bool, Int, Float and String are singleton instance.
		// So comparing directly is OK.
		if l == right {
			return nil
		}
	case *Tuple:
		if r, ok := right.(*Tuple); ok {
			return unifyTuple(l, r)
		}
	case *Array:
		if r, ok := right.(*Array); ok {
			return Unify(l.Elem, r.Elem)
		}
	case *Fun:
		if r, ok := right.(*Fun); ok {
			return unifyFun(l, r)
		}
	case *Var:
		return unifyVar(l, right)
	}

	if v, ok := right.(*Var); ok {
		return unifyVar(v, left)
	}

	return errors.Errorf("Cannot unify types. Type mismatch between '%s' and '%s'", left.String(), right.String())
}
