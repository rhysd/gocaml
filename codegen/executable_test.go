package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/closure"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutable(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*.ml")
	if err != nil {
		panic(err)
	}
	outputs, err := filepath.Glob("testdata/*.out")
	if err != nil {
		panic(err)
	}
	if len(inputs) == 0 {
		panic("No test found")
	}
	for _, input := range inputs {
		base := filepath.Base(input)
		expect := ""
		outputFile := strings.TrimSuffix(input, filepath.Ext(input)) + ".out"
		for _, e := range outputs {
			if e == outputFile {
				expect = e
				break
			}
		}
		if expect == "" {
			panic(fmt.Sprintf("Expected output file '%s' was not found for code '%s'", outputFile, input))
		}
		t.Run(base, func(t *testing.T) {
			s, err := token.NewSourceFromFile(input)
			if err != nil {
				t.Fatal(err)
			}

			l := lexer.NewLexer(s)
			go l.Lex()

			root, err := parser.Parse(l.Tokens)
			if err != nil {
				t.Fatal(err)
			}

			if err = alpha.Transform(root); err != nil {
				t.Fatal(err)
			}

			env := typing.NewEnv()
			if err := env.ApplyTypeAnalysis(root); err != nil {
				t.Fatal(err)
			}

			ir, err := gcil.FromAST(root, env)
			if err != nil {
				t.Fatal(err)
			}
			gcil.ElimRefs(ir, env)
			prog := closure.Transform(ir)

			opts := EmitOptions{OptimizeDefault, "", "", true}
			emitter, err := NewEmitter(prog, env, s, opts)
			if err != nil {
				t.Fatal(err)
			}
			emitter.RunOptimizationPasses()
			outfile, err := filepath.Abs("__test_a.out")
			if err != nil {
				panic(err)
			}
			if err := emitter.EmitExecutable(outfile); err != nil {
				t.Fatal(err)
			}
			defer os.Remove(outfile)

			bytes, err := exec.Command(outfile).Output()
			if err != nil {
				t.Fatal(err)
			}
			got := string(bytes)
			bytes, err = ioutil.ReadFile(expect)
			if err != nil {
				panic(err)
			}
			want := string(bytes[:len(bytes)-1]) // Trim EOL (newline at the end of file)

			if got != want {
				t.Fatalf("Unexpected output from executable:\n\nGot: '%s'\nWant: '%s'", got, want)
			}
		})
	}
}
