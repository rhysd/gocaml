package gcil

import (
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

func TestEliminatingRef(t *testing.T) {
	cases := []struct {
		what     string
		code     string
		expected []string
	}{
		{
			"binary operator",
			"let a = 1 in let b = 2 in a + b",
			[]string{
				"= binary + a$t1 b$t2",
			},
		},
		{
			"unary operator",
			"let a = 1 in -a",
			[]string{
				"= unary - a$t1",
			},
		},
		{
			"if expression",
			"let a = 1 in if a < 0 then a else a + 1",
			[]string{
				"= binary < a$t1",
				"= ref a$t1", // This ref cannot be eliminated
				"= binary + a$t1",
			},
		},
		{
			"function",
			"let x = 1 in let rec f a = f x in (f 42) + 0",
			[]string{
				"= app f$t2 x$t1",
				"= app f$t2 $k6",
			},
		},
		{
			"tuple",
			"let x = 1 in let t = (x, 2, 3) in let (a, b, c) = t in a",
			[]string{
				"= tuple x$t1,$k3,$k4",
				"a$t3 = tplload 0 t$t2",
				"b$t4 = tplload 1 t$t2",
				"c$t5 = tplload 2 t$t2",
				"= ref a$t3", // This ref cannot be eliminated
			},
		},
		{
			"array",
			"let x = 1 in let arr = Array.create x x in arr.(x); arr.(x) <- x",
			[]string{
				"= array x$t1 x$t1",
				"= arrload x$t1 arr$t2",
				"= arrstore x$t1 arr$t2 x$t1",
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
			ElimRefs(ir, env)
			var buf bytes.Buffer
			ir.Println(&buf, env)
			actual := buf.String()
			for _, expected := range tc.expected {
				if !strings.Contains(actual, expected) {
					t.Errorf("Expected to contain '%s' in '%s'", expected, actual)
				}
			}
		})
	}
}
