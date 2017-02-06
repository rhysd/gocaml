package closure

import (
	"fmt"
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"testing"
)

func TestFreeVarsAnalysis(t *testing.T) {
	cases := []struct {
		what     string
		code     string
		expected map[string]freeVars
	}{
		{
			"no function",
			"let x = 42 in if true then print_int(x) else x + 1",
			map[string]freeVars{},
		},
		{
			"no free var function",
			"let rec f a = a in f 42",
			map[string]freeVars{
				"f$t1": {
					[]string{},
					false,
				},
			},
		},
		{
			"no free var nested functions",
			"let rec f a = a in let rec g a = a in f (g 42)",
			map[string]freeVars{
				"f$t1": {
					[]string{},
					false,
				},
				"g$t3": {
					[]string{},
					false,
				},
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
			ir := gcil.EmitIR(root, env)
			gcil.ElimRefs(ir, env)
			analysis := newFreeVarAnalysis(env)
			analysis.analyzeBlock(ir)
			fvs := analysis.funcs
			if len(fvs) != len(tc.expected) {
				t.Errorf("Number of analyzed function mismatch: %d v.s. %d", len(fvs), len(tc.expected))
			}
			for fun, fv := range fvs {
				expected, ok := tc.expected[fun]
				if !ok {
					t.Fatalf("Expected function '%s' is not found in analyzed ones %v", fun, fv)
				}
				if fv.appearInRef != expected.appearInRef {
					t.Errorf("appaerInRef mismatch at '%s' (actual: %v, expected %v)", fun, fv.appearInRef, expected.appearInRef)
				}
				if len(fv.names) != len(expected.names) {
					t.Fatalf("Number of free variables mismatch at %s: %d v.s. %d (actual=%v)", fun, len(fv.names), len(expected.names), fv.names)
				}
				for i, name := range fv.names {
					if name != expected.names[i] {
						t.Errorf("Name of %dth free variable mismatch at %s: actual: '%s' but expected '%s'", i, fun, name, expected.names[i])
					}
				}
			}
		})
	}
}
