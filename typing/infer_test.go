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
			what: "! with non-bool value",
			code: "!42",
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
