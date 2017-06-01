package typing

import (
	"strings"
	"testing"
)

func testTypeEquals(l, r Type) bool {
	switch l := l.(type) {
	case *Unit, *Int, *Float, *Bool, *String:
		return l == r
	case *Tuple:
		r, ok := r.(*Tuple)
		if !ok || len(l.Elems) != len(r.Elems) {
			return false
		}
		for i, e := range l.Elems {
			if !testTypeEquals(e, r.Elems[i]) {
				return false
			}
		}
		return true
	case *Array:
		r, ok := r.(*Array)
		if !ok {
			return false
		}
		return testTypeEquals(l.Elem, r.Elem)
	case *Fun:
		r, ok := r.(*Fun)
		if !ok || !testTypeEquals(l.Ret, r.Ret) || len(l.Params) != len(r.Params) {
			return false
		}
		for i, p := range l.Params {
			if !testTypeEquals(p, r.Params[i]) {
				return false
			}
		}
		return true
	case *Var:
		r, ok := r.(*Var)
		if !ok {
			return false
		}
		if l.Ref == nil || r.Ref == nil {
			return l.Ref == nil && r.Ref == nil
		}
		return testTypeEquals(l.Ref, r.Ref)
	case *Option:
		r, ok := r.(*Option)
		if !ok {
			return false
		}
		return testTypeEquals(l.Elem, r.Elem)
	default:
		panic("Unreachable")
	}
}

func TestUnwrapExternalSimpleTypes(t *testing.T) {
	v := &typeVarDereferencer{[]string{}, NewEnv()}
	for _, ty := range []Type{
		UnitType,
		IntType,
		FloatType,
		BoolType,
		&Tuple{
			[]Type{
				UnitType,
				IntType,
				&Tuple{
					[]Type{FloatType, IntType},
				},
				&Array{IntType},
				&Option{IntType},
				&Option{&Option{&Array{IntType}}},
			},
		},
		&Fun{IntType, []Type{FloatType, BoolType}},
	} {
		a := v.derefExternalSym("test", ty)
		if a != ty {
			t.Errorf("It must be %s but actually %s", ty.String(), a.String())
		}
		if len(v.errors) > 0 {
			t.Errorf("Unexpected error at %s: %s", ty.String(), strings.Join(v.errors, "\n"))
		}
	}
}

func TestUnwrapTypeVarsInExternals(t *testing.T) {
	v := &typeVarDereferencer{[]string{}, NewEnv()}
	for _, tc := range []struct {
		input    Type
		expected Type
	}{
		{&Var{UnitType}, UnitType},
		{&Var{&Var{IntType}}, IntType},
		{&Tuple{[]Type{&Var{FloatType}, &Var{IntType}}}, &Tuple{[]Type{FloatType, IntType}}},
		{&Array{&Var{&Tuple{[]Type{&Var{IntType}, UnitType}}}}, &Array{&Tuple{[]Type{IntType, UnitType}}}},
		{&Option{&Var{&Option{&Var{IntType}}}}, &Option{&Option{IntType}}},
	} {
		actual := v.derefExternalSym("test", tc.input)
		if !testTypeEquals(actual, tc.expected) {
			t.Errorf("Expected dereferenced type to be '%s' but actually '%s'", tc.expected.String(), actual.String())
		}
		if len(v.errors) > 0 {
			t.Errorf("Unexpected error at type %s: %s", tc.input.String(), strings.Join(v.errors, "\n"))
		}
	}
}

func TestRaiseErrorOnUnknownTypeInExternals(t *testing.T) {
	v := &typeVarDereferencer{[]string{}, NewEnv()}
	for _, ty := range []Type{
		&Var{},
		&Var{&Var{}},
		&Tuple{[]Type{IntType, &Var{}}},
		&Array{&Var{}},
		&Fun{IntType, []Type{&Var{&Var{}}}},
		&Fun{&Array{&Var{}}, []Type{}},
		&Option{&Var{}},
		&Fun{&Option{&Var{}}, []Type{}},
	} {
		v.derefExternalSym("test", ty)
		if len(v.errors) == 0 {
			t.Errorf("Error should be raised for dereferencing external's type %s", ty.String())
		}
	}
}

func TestFixReturnTypeOfExternalFunction(t *testing.T) {
	v := &typeVarDereferencer{[]string{}, NewEnv()}
	for _, ty := range []Type{
		&Fun{&Var{}, []Type{}},
		&Fun{&Var{&Var{}}, []Type{IntType}},
		&Var{&Fun{&Var{}, []Type{FloatType}}},
	} {
		derefed := v.derefExternalSym("test", ty)
		f, ok := derefed.(*Fun)
		if !ok {
			t.Errorf("It must be dereferenced as function but actually %s", derefed.String())
			continue
		}
		if _, ok := f.Ret.(*Unit); !ok {
			t.Errorf("Return type must be fixed to unit but actually %s", f.Ret.String())
		}
	}
}
