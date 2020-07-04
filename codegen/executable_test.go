package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/closure"
	"github.com/rhysd/gocaml/sema"
	"github.com/rhysd/gocaml/syntax"
	"github.com/rhysd/locerr"
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
			defer func() {
				err := recover()
				if err != nil {
					t.Fatal(err)
				}
			}()

			s, err := locerr.NewSourceFromFile(input)
			if err != nil {
				t.Fatal(err)
			}

			ast, err := syntax.Parse(s)
			if err != nil {
				t.Fatal(err)
			}

			env, ir, err := sema.SemanticsCheck(ast)
			if err != nil {
				t.Fatal(err)
			}
			prog := closure.Transform(ir)

			opts := EmitOptions{OptimizeDefault, "", "", true}
			emitter, err := NewEmitter(prog, env, s, opts)
			if err != nil {
				t.Fatal(err)
			}
			defer emitter.Dispose()
			emitter.RunOptimizationPasses()
			outfile, err := filepath.Abs(fmt.Sprintf("test.%s.a.out", base))
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
			want := ""
			if len(bytes) > 0 {
				want = string(bytes[:len(bytes)-1]) // Trim EOL (newline at the end of file)
			}

			if got != want {
				t.Fatalf("Unexpected output from executable:\n\nGot: '%s'\nWant: '%s'", got, want)
			}
		})
	}
}

func BenchmarkExecutableCreation(b *testing.B) {
	inputs, err := filepath.Glob("testdata/*.ml")
	if err != nil {
		panic(err)
	}
	if len(inputs) == 0 {
		panic("No test found")
	}
	sources := make(map[string]*locerr.Source, len(inputs))
	for _, input := range inputs {
		source, err := locerr.NewSourceFromFile(input)
		if err != nil {
			b.Fatal(err)
		}
		base := filepath.Base(input)
		sources[base] = source
	}

	makeEmitter := func(source *locerr.Source) *Emitter {
		ast, err := syntax.Parse(source)
		if err != nil {
			b.Fatal(err)
		}

		env, ir, err := sema.SemanticsCheck(ast)
		if err != nil {
			b.Fatal(err)
		}
		prog := closure.Transform(ir)

		opts := EmitOptions{OptimizeDefault, "", "", true}
		emitter, err := NewEmitter(prog, env, source, opts)
		if err != nil {
			b.Fatal(err)
		}
		return emitter
	}

	b.Run("emit executable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for base, source := range sources {
				emitter := makeEmitter(source)
				defer emitter.Dispose()
				emitter.RunOptimizationPasses()
				outfile, err := filepath.Abs(fmt.Sprintf("test.%s.a.out", base))
				if err != nil {
					panic(err)
				}
				if err := emitter.EmitExecutable(outfile); err != nil {
					b.Fatal(err)
				}
				defer os.Remove(outfile)
			}
		}
	})
	b.Run("build LLVM IR", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, source := range sources {
				e := makeEmitter(source)
				defer e.Dispose()
			}
		}
	})
}

func TestExamples(t *testing.T) {
	examples, err := filepath.Glob("../examples/*.ml")
	if err != nil {
		panic(err)
	}
	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			s, err := locerr.NewSourceFromFile(example)
			if err != nil {
				t.Fatal(err)
			}

			ast, err := syntax.Parse(s)
			if err != nil {
				t.Fatal(err)
			}

			env, ir, err := sema.SemanticsCheck(ast)
			if err != nil {
				t.Fatal(err)
			}
			prog := closure.Transform(ir)

			opts := EmitOptions{OptimizeDefault, "", "", true}
			emitter, err := NewEmitter(prog, env, s, opts)
			if err != nil {
				t.Fatal(err)
			}
			defer emitter.Dispose()
			emitter.RunOptimizationPasses()
		})
	}
}
