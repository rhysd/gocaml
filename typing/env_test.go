package typing

import (
	"bytes"
	"fmt"
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/loc"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvedSymbols(t *testing.T) {
	s := loc.NewDummySource("let x = 1 in x + y; ()")
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	env, err := TypeInferernce(ast)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := env.Table["x"]; !ok {
		t.Errorf("'x' was not resolved as internal symbol: %v", env.Table)
	}
	if _, ok := env.Externals["y"]; !ok {
		t.Errorf("'y' was not resolved as external symbol: %v", env.Externals)
	}
}

func TestTypeCheckOK(t *testing.T) {
	testdir := filepath.FromSlash("../testdata/from-mincaml/")
	files, err := ioutil.ReadDir(testdir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		n := filepath.Join(testdir, f.Name())
		if !strings.HasSuffix(n, ".ml") {
			continue
		}

		t.Run(fmt.Sprintf("Infer types successfully: %s", n), func(t *testing.T) {
			s, err := loc.NewSourceFromFile(n)
			if err != nil {
				panic(err)
			}

			l := lexer.NewLexer(s)
			go l.Lex()

			ast, err := parser.Parse(l.Tokens)
			if err != nil {
				panic(ast.Root)
			}

			if err = alpha.Transform(ast.Root); err != nil {
				panic(err)
			}

			_, err = TypeInferernce(ast)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProgramRootTypeIsUnit(t *testing.T) {
	s := loc.NewDummySource("42")
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	_, err = TypeInferernce(ast)
	if err == nil {
		t.Fatalf("Type check must raise an error when root type of program is not ()")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Type of root expression of program must be unit") {
		t.Fatalf("Expected error for root type of program but actually '%s'", msg)
	}
}

func TestTypeCheckFail(t *testing.T) {
	s := loc.NewDummySource("let x = 42 in x +. 3.14")
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	_, err = TypeInferernce(ast)
	if err == nil {
		t.Fatalf("Type check must raise a type error")
	}
}

func TestDumpResult(t *testing.T) {
	s := loc.NewDummySource("let x = 42 in x + y; ()")
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	env, err := TypeInferernce(ast)
	if err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	env.Dump()

	ch := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		ch <- buf.String()
	}()
	w.Close()
	os.Stdout = old

	out := <-ch
	if !strings.HasPrefix(out, "Variables:\n") {
		t.Errorf("Output does not contain internal symbols table: %s", out)
	}
	if !strings.Contains(out, "External Variables:\n") {
		t.Errorf("Output does not contain external symbols table: %s", out)
	}
}

func TestDerefNoneTypes(t *testing.T) {
	s := loc.NewDummySource("let rec f x = () in f (Some 42); f None; let a = None in f a")
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	env, err := TypeInferernce(ast)
	if err != nil {
		t.Fatal(err)
	}

	if len(env.NoneTypes) != 2 {
		t.Fatal("None type values were not detected")
	}

	for _, o := range env.NoneTypes {
		v, ok := o.Elem.(*Var)
		if ok {
			t.Errorf("Element type of 'None' value was not dereferenced: %s", v.String())
		}
	}
}
