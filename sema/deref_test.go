package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/syntax"
	"github.com/rhysd/gocaml/token"
	. "github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
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

func TestDerefFailure(t *testing.T) {
	s := locerr.NewDummySource("")
	pos := locerr.Pos{0, 0, 0, s}
	tok := &token.Token{token.ILLEGAL, pos, pos, s}
	env := NewEnv()
	env.Table["hello"] = &Var{}
	v := &typeVarDereferencer{nil, env, map[ast.Expr]Type{}}
	root := &ast.Let{
		tok,
		ast.NewSymbol("hello"),
		&ast.Int{tok, 0},
		&ast.Int{tok, 0},
		nil,
	}
	ast.Visit(v, root)
	if v.err == nil {
		t.Fatal("Unknown symbol 'hello' must cause an error")
	}
	msg := v.err.Error()
	if !strings.Contains(msg, "Cannot infer type of variable 'hello'") {
		t.Fatal("Unexpected error message:", msg)
	}
}

func TestUnwrapEmptyTypeVar(t *testing.T) {
	e := &Var{}
	for _, ty := range []Type{
		e,
		&Var{e},
		&Var{&Var{e}},
		&Tuple{[]Type{e}},
		&Fun{e, []Type{}},
		&Fun{IntType, []Type{e}},
		&Option{e},
		&Array{e},
	} {
		_, ok := unwrap(ty)
		if ok {
			t.Error("Unwrapping type variable must cause an error:", ty.String())
		}
	}
}

func TestMiscCheckError(t *testing.T) {
	cases := []struct {
		what     string
		code     string
		expected string
	}{
		{
			what:     "unit is invalid for operator '<'",
			code:     "() < ()",
			expected: "'unit' can't be compared with operator '<'",
		},
		{
			what:     "tuple is invalid for operator '<'",
			code:     "let t = (1, 2) in t < t",
			expected: "'int * int' can't be compared with operator '<'",
		},
		{
			what:     "option is invalid for operator '<'",
			code:     "let a = Some 3 in a < None",
			expected: "'int option' can't be compared with operator '<'",
		},
		{
			what:     "array is invalid for operator '='",
			code:     "let a = Array.make  3 3 in a = a",
			expected: "'int array' can't be compared with operator '='",
		},
	}

	for _, tc := range cases {
		t.Run(tc.what, func(t *testing.T) {
			s := locerr.NewDummySource(fmt.Sprintf("%s; ()", tc.code))
			parsed, err := syntax.Parse(s)
			if err != nil {
				t.Fatal(err)
			}

			env := NewEnv()
			if err := AlphaTransform(parsed, env); err != nil {
				t.Fatal(err)
			}

			inf := NewInferer(env)

			// inf.Infer() invokes type dereferences
			err = inf.Infer(parsed)

			if err == nil {
				t.Fatalf("Expected code '%s' to cause an error '%s' but actually there is no error", tc.code, tc.expected)
			}
			if !strings.Contains(err.Error(), tc.expected) {
				t.Fatalf("Error message '%s' does not contain '%s'", err.Error(), tc.expected)
			}
		})
	}
}
