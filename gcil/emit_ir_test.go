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
