package gcil

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"strings"
	"testing"
)

func TestEmitInsn(t *testing.T) {
	cases := []struct {
		what     string
		code     string
		expected []string
	}{
		{
			"int",
			"42",
			[]string{"int 42 ; type=int"},
		},
		{
			"unit",
			"()",
			[]string{"unit ; type=()"},
		},
		{
			"float",
			"3.14",
			[]string{"float 3.140000 ; type=float"},
		},
		{
			"boolean",
			"false",
			[]string{"bool false ; type=bool"},
		},
		{
			"unary relational op",
			"not true",
			[]string{
				"bool true ; type=bool",
				"unary not $k1 ; type=bool",
			},
		},
		{
			"unary arithmetic op",
			"-42",
			[]string{
				"int 42 ; type=int",
				"unary - $k1 ; type=int",
			},
		},
		{
			"binary int op",
			"1 + 2",
			[]string{
				"int 1 ; type=int",
				"int 2 ; type=int",
				"binary + $k1 $k2 ; type=int",
			},
		},
		{
			"binary float op",
			"3.14 *. 2.0",
			[]string{
				"float 3.140000 ; type=float",
				"float 2.000000 ; type=float",
				"binary *. $k1 $k2 ; type=float",
			},
		},
		{
			"binary relational op",
			"1 < 2",
			[]string{
				"int 1 ; type=int",
				"int 2 ; type=int",
				"binary < $k1 $k2 ; type=bool",
			},
		},
		{
			"if expression",
			"if 1 < 2 then 3 else 4",
			[]string{
				"int 1 ; type=int",
				"int 2 ; type=int",
				"binary < $k1 $k2 ; type=bool",
				"if $k3",
				"BEGIN: then",
				"int 3 ; type=int",
				"END: then",
				"BEGIN: else",
				"int 4 ; type=int",
				"END: else",
			},
		},
		{
			"let expression and variable reference",
			"let a = 1 in let b = a in b",
			[]string{
				"int 1 ; type=int",
				"ref a$t1 ; type=int",
				"ref b$t2 ; type=int",
			},
		},
		{
			"function and its application",
			"let rec f a = a + 1 in f 3",
			[]string{
				"fun a$t2",
				"BEGIN: body (f$t1)",
				"ref a$t2 ; type=int",
				"int 1 ; type=int",
				"binary + $k1 $k2 ; type=int",
				"END: body (f$t1)",
				"; type=int -> int",
				"ref f$t1 ; type=int -> int",
				"int 3 ; type=int",
				"app $k4 $k5 ; type=int",
			},
		},
		{
			"tuple literal",
			"(1, 2, 3)",
			[]string{
				"int 1 ; type=int",
				"int 2 ; type=int",
				"int 3 ; type=int",
				"tuple $k1,$k2,$k3 ; type=(int, int, int)",
			},
		},
		{
			"let tuple substitution",
			"let (a, b) = (1, 2) in a + b",
			[]string{
				"int 1 ; type=int",
				"int 2 ; type=int",
				"tuple $k1,$k2 ; type=(int, int)",
				"tplload 0 $k3 ; type=int",
				"tplload 1 $k3 ; type=int",
				"ref a$t1 ; type=int",
				"ref b$t2 ; type=int",
				"binary + $k4 $k5 ; type=int",
			},
		},
		{
			"array creation",
			"Array.create 3 true",
			[]string{
				"int 3 ; type=int",
				"bool true ; type=bool",
				"array $k1 $k2 ; type=bool array",
			},
		},
		{
			"access to array",
			"let a = Array.create 3 true in a.(1)",
			[]string{
				"int 3 ; type=int",
				"bool true ; type=bool",
				"array $k1 $k2 ; type=bool array",
				"ref a$t1 ; type=bool array",
				"int 1 ; type=int",
				"arrload $k5 $k4 ; type=bool",
			},
		},
		{
			"modify element of array",
			"let a = Array.create 3 true in a.(1) <- false",
			[]string{
				"int 3 ; type=int",
				"bool true ; type=bool",
				"array $k1 $k2 ; type=bool array",
				"ref a$t1 ; type=bool array",
				"int 1 ; type=int",
				"bool false ; type=bool",
				"arrstore $k5 $k4 $k6 ; type=bool",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.what, func(t *testing.T) {
			s := token.NewDummySource(fmt.Sprintf("%s; ()", tc.code))
			l := lexer.NewLexer(s)
			go l.Lex()
			root, err := parser.Parse(l.Tokens)
			if err != nil {
				panic(err)
			}
			if err = alpha.Transform(root); err != nil {
				panic(err)
			}
			env := typing.NewEnv()
			if err := env.ApplyTypeAnalysis(root); err != nil {
				panic(err)
			}
			ir := EmitIR(root, env)
			var buf bytes.Buffer
			ir.Println(&buf)
			r := bufio.NewReader(&buf)
			line, _, err := r.ReadLine()
			if err != nil {
				panic(err)
			}
			if string(line) != "BEGIN: program" {
				t.Fatalf("First line must begin with 'BEGIN: program' because it's root block")
			}
			for i, expected := range tc.expected {
				line, _, err = r.ReadLine()
				if err != nil {
					panic(err)
				}
				actual := string(line)
				if !strings.HasSuffix(actual, expected) {
					t.Errorf("Expected to end with '%s' for line %d of output of code '%s'. But actually output was '%s'", expected, i, tc.code, actual)
				}
			}
		})
	}
}