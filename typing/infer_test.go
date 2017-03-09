package typing

import (
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"strings"
	"testing"
)

func TestEdgeCases(t *testing.T) {
	testcases := []struct {
		what string
		code string
	}{
		{
			what: "param and function have the same name",
			code: "let rec f f = f + 1 in print_int (f 10)",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.what, func(t *testing.T) {
			s := token.NewDummySource(tc.code)
			l := lexer.NewLexer(s)
			go l.Lex()
			e, err := parser.Parse(l.Tokens)
			if err != nil {
				panic(err)
			}
			if err = alpha.Transform(e); err != nil {
				panic(err)
			}
			env := NewEnv()
			_, err = env.infer(e)
			if err != nil {
				t.Fatalf("Type check raised an error for code '%s': %s", tc.code, err.Error())
			}
		})
	}
}

func TestInvalidExpressions(t *testing.T) {
	testcases := []struct {
		what     string
		code     string
		expected string
	}{
		{
			what:     "+. with int",
			code:     "1 +. 2",
			expected: "Type mismatch between 'float' and 'int'",
		},
		{
			what:     "+ with float",
			code:     "1.0 + 2.0",
			expected: "Type mismatch between 'int' and 'float'",
		},
		{
			what:     "/ with float",
			code:     "1.0 / 2.0",
			expected: "Type mismatch between 'int' and 'float'",
		},
		{
			what:     "'not' with non-bool value",
			code:     "not 42",
			expected: "Type mismatch between 'bool' and 'int'",
		},
		{
			what:     "invalid equal compare",
			code:     "41 = true",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "invalid = compare",
			code:     "41 = 3.14",
			expected: "Type mismatch between 'int' and 'float'",
		},
		{
			what:     "invalid <> compare",
			code:     "41 <> 3.14",
			expected: "Type mismatch between 'int' and 'float'",
		},
		{
			what:     "invalid < compare",
			code:     "41 < true",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "invalid <= compare",
			code:     "41 <= true",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "invalid > compare",
			code:     "41 > true",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "invalid >= compare",
			code:     "41 >= true",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "&& must have boolean operands",
			code:     "42 && true",
			expected: "Type mismatch between 'bool' and 'int'",
		},
		{
			what:     "|| must have boolean operands",
			code:     "false || 42",
			expected: "Type mismatch between 'bool' and 'int'",
		},
		{
			what:     "&& is evaluated as bool",
			code:     "(true && false) + 3",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "/. with int",
			code:     "1 /. 2",
			expected: "Type mismatch between 'float' and 'int'",
		},
		{
			what:     "*. with int",
			code:     "1 *. 2",
			expected: "Type mismatch between 'float' and 'int'",
		},
		{
			what:     "unary - without number",
			code:     "-true",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "not a bool condition in if",
			code:     "if 42 then true else false",
			expected: "Type mismatch between 'bool' and 'int'",
		},
		{
			what:     "mismatch type between else and then",
			code:     "if true then 42 else 4.2",
			expected: "Type mismatch between 'int' and 'float'",
		},
		{
			what:     "mismatch type of variable",
			code:     "let x = true in x + 42",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "mismatch parameter type",
			code:     "let rec f a b = a < b in (f 1 1) = (f 1.0 1.0)",
			expected: "On unifying 1th parameter of function '(int, int) -> bool' and '(float, float) -> bool'",
		},
		{
			what:     "does not meet parameter type requirements",
			code:     "let rec f a b = a + b in f 1 1.0",
			expected: "On unifying 2th parameter of function '(int, int) -> int' and '(int, float) -> int'",
		},
		{
			what:     "wrong number of arguments",
			code:     "let rec f a b = a + b in f 1",
			expected: "Number of parameters of function does not match between '(int, int) -> int' and '(int) -> int'",
		},
		{
			what:     "type mismatch in return type",
			code:     "let rec f a b = a + b in 1.0 +. f 1 2",
			expected: "Type mismatch between 'int' and 'float'",
		},
		{
			what:     "wrong number of tuple assignment",
			code:     "let (x, y) = (1, 2, 3) in ()",
			expected: "Number of elements of tuple does not match",
		},
		{
			what:     "type mismatch for tuple elements",
			code:     "let (x, y) = (1, 2.0) in x + y",
			expected: "Type mismatch between 'float' and 'int'",
		},
		{
			what:     "index is not a number",
			code:     "let a = Array.create 3 1.0 in a.(true)",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "wrong array length type",
			code:     "let a = Array.create true 1.0 in ()",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "element type mismatch in array",
			code:     "let a = Array.create 3 1.0 in 1 + a.(0)",
			expected: "Type mismatch between 'float' and 'int'",
		},
		{
			what:     "index access to wrong value",
			code:     "true.(1)",
			expected: "} array' and 'bool'",
		},
		{
			what:     "set wrong type value to array",
			code:     "let a = Array.create 3 1.0 in a.(0) <- true",
			expected: "Type mismatch between 'bool' and 'float'",
		},
		{
			what:     "wrong index type in index access",
			code:     "let a = Array.create 3 1.0 in a.(true) <- 2.0",
			expected: "Type mismatch between 'int' and 'bool'",
		},
		{
			what:     "index assign to wrong value",
			code:     "false.(1) <- 10",
			expected: "Type mismatch between 'int array' and 'bool'",
		},
		{
			what:     "index assign is evaluated as unit",
			code:     "let a = Array.make 3 1.0 in 1.0 = a.(0) <- 2.0",
			expected: "Type mismatch between 'float' and '()'",
		},
		{
			what:     "Array.size with invalid argument",
			code:     "Array.size true",
			expected: "} array' and 'bool'",
		},
		{
			what:     "Array.size returns int type value",
			code:     "(Array.size (Array.create 3 true)) = 3.0",
			expected: "'int' and 'float'",
		},
		{
			what:     "occur check",
			code:     "let rec f x = f in f 4",
			expected: "Cyclic dependency found while unification with",
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
			if err = alpha.Transform(e); err != nil {
				panic(err)
			}
			env := NewEnv()
			_, err = env.infer(e)
			if err == nil {
				t.Fatalf("Type check did not raise an error for code '%s'", testcase.code)
			}
			if !strings.Contains(err.Error(), testcase.expected) {
				t.Fatalf("Expected error message '%s' to contain '%s'", err.Error(), testcase.expected)
			}
		})
	}
}
