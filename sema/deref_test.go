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

func varT(t Type) *Var {
	return NewVar(t, 0)
}

func TestDerefFailure(t *testing.T) {
	s := locerr.NewDummySource("")
	pos := locerr.Pos{0, 0, 0, s}
	tok := &token.Token{token.ILLEGAL, pos, pos, s}
	env := NewEnv()
	env.DeclTable["hello"] = varT(nil)
	v := &typeVarDereferencer{
		nil,
		env,
		map[ast.Expr]Type{},
		schemes{},
	}
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
	e := varT(nil)
	for _, ty := range []Type{
		e,
		varT(e),
		varT(varT(e)),
		&Tuple{[]Type{e}},
		&Fun{e, []Type{}},
		&Fun{IntType, []Type{e}},
		&Option{e},
		&Array{e},
	} {
		v := &typeVarDereferencer{
			nil,
			NewEnv(),
			map[ast.Expr]Type{},
			schemes{},
		}
		_, ok := v.unwrap(ty)
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
