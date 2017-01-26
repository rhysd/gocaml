package typing

import (
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"testing"
)

func TestInvalidExpressions(t *testing.T) {
	testcases := []struct {
		what string
		code string
	}{
		{
			what: "+. with int",
			code: "1 +. 2",
		},
		{
			what: "+ with float",
			code: "1.0 + 2.0",
		},
		{
			what: "'not' with non-bool value",
			code: "not 42",
		},
		{
			what: "invalid equal compare",
			code: "41 = true",
		},
		{
			what: "invalid less compare",
			code: "41 < true",
		},
		{
			what: "invalid less compare",
			code: "41 < true",
		},
		{
			what: "/. with int",
			code: "1 /. 2",
		},
		{
			what: "*. with int",
			code: "1 *. 2",
		},
		{
			what: "unary - without number",
			code: "-true",
		},
		{
			what: "not a bool condition in if",
			code: "if 42 then true else false",
		},
		{
			what: "mismatch type between else and then",
			code: "if true then 42 else 4.2",
		},
		{
			what: "mismatch type of variable",
			code: "let x = true in x + 42",
		},
		{
			what: "mismatch type of variable",
			code: "let x = true in x + 42",
		},
		{
			what: "mismatch parameter type",
			code: "let rec f a b = a < b in (f 1 1) = (f 1.0 1.0)",
		},
		{
			what: "does not meet parameter type requirements",
			code: "let rec f a b = a + b in f 1 1.0",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.what, func(t *testing.T) {
			s := token.NewDummySource(testcase.code)
			l := lexer.NewLexer(s)
			go l.Lex()
			e, err := parser.Parse(l.Tokens)
			if err != nil {
				panic(err)
			}
			env := NewEnv()
			if _, err := env.infer(e); err == nil {
				t.Fatalf("Type check did not raise an error for code '%s'", testcase.code)
			}
		})
	}
}

// TODO: Test external variables detection
