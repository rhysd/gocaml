package sema

import (
	"github.com/rhysd/gocaml/common"
	. "github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
)

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
	case *Option:
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
		if t.Ref != nil {
			return occur(v, t.Ref)
		}
		if t.IsGeneric() {
			panic("FATAL: Generic type variable must not appear in occur check")
		}
		if v == t {
			return true
		}
		if t.Level > v.Level {
			// Adjust levels
			t.Level = v.Level
		}
	}
	return false
}

func unifyTuple(left, right *Tuple) *locerr.Error {
	length := len(left.Elems)
	if length != len(right.Elems) {
		return locerr.Errorf("Number of elements of tuple does not match: %d vs %d (between '%s' and '%s')", length, len(right.Elems), left.String(), right.String())
	}

	for i := 0; i < length; i++ {
		l := left.Elems[i]
		r := right.Elems[i]
		if err := Unify(l, r); err != nil {
			return locerr.Notef(err, "On unifying tuples' %s elements of '%s' and '%s'", common.Ordinal(i+1), left.String(), right.String())
		}
	}

	return nil
}

func unifyFun(left, right *Fun) *locerr.Error {
	if err := Unify(left.Ret, right.Ret); err != nil {
		return locerr.Notef(err, "On unifying functions' return types of '%s' and '%s'", left.String(), right.String())
	}

	if len(left.Params) != len(right.Params) {
		return locerr.Errorf("Number of parameters of function does not match: %d vs %d (between '%s' and '%s')", len(left.Params), len(right.Params), left.String(), right.String())
	}

	for i, l := range left.Params {
		r := right.Params[i]
		if err := Unify(l, r); err != nil {
			return locerr.Notef(err, "On unifying %s parameter of function '%s' and '%s'", common.Ordinal(i+1), left.String(), right.String())
		}
	}

	return nil
}

func assignVar(v *Var, t Type) *locerr.Error {
	// When rv.Ref == nil
	if occur(v, t) {
		return locerr.Errorf("Cannot resolve free type variable. Cyclic dependency found for free type variable '%s' while unification with '%s'", v.String(), t.String())
	}

	// Note:
	// 'v' may be generic type variable because of external symbols.
	// e.g.
	//   let _ = x in x + x
	// The `x` is an external symbol and typed as ?. And it is bound to `_` in `let` expression.
	// The `_` is typed as 'a so the type of `x` will be 'a.
	// In `x + x`, type of `x` is unified although its type is generic.

	v.Ref = t
	return nil
}

func Unify(left, right Type) *locerr.Error {
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
	case *Option:
		if r, ok := right.(*Option); ok {
			return Unify(l.Elem, r.Elem)
		}
	case *Fun:
		if r, ok := right.(*Fun); ok {
			return unifyFun(l, r)
		}
	}

	lv, lok := left.(*Var)
	rv, rok := right.(*Var)

	// Order of below 'if' statements is important! (#15)

	if (lok && rok) && (lv == rv) {
		return nil
	}
	if lok && lv.Ref != nil {
		return Unify(lv.Ref, right)
	}
	if rok && rv.Ref != nil {
		return Unify(left, rv.Ref)
	}
	if lok {
		// When lv.Ref == nil
		return assignVar(lv, right)
	}
	if rok {
		// When rv.Ref == nil
		return assignVar(rv, left)
	}

	return locerr.Errorf("Cannot unify types. Type mismatch between '%s' and '%s'", left.String(), right.String())
}
