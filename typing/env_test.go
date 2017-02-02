package typing

import (
	"fmt"
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvedSymbols(t *testing.T) {
	s := token.NewDummySource("let x = 1 in x + y; ()")
	l := lexer.NewLexer(s)
	go l.Lex()
	root, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(root)
	}

	env := NewEnv()
	if err := env.ApplyTypeAnalysis(root); err != nil {
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
			s, err := token.NewSourceFromFile(n)
			if err != nil {
				panic(err)
			}

			l := lexer.NewLexer(s)
			go l.Lex()

			root, err := parser.Parse(l.Tokens)
			if err != nil {
				panic(root)
			}

			if err = alpha.Transform(root); err != nil {
				panic(err)
			}

			env := NewEnv()
			if err := env.ApplyTypeAnalysis(root); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProgramRootTypeIsUnit(t *testing.T) {
	s := token.NewDummySource("42")
	l := lexer.NewLexer(s)
	go l.Lex()
	root, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(root)
	}

	env := NewEnv()
	err = env.ApplyTypeAnalysis(root)
	if err == nil {
		t.Fatalf("Type check must raise an error when root type of program is not ()")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Type of root expression of program must be unit") {
		t.Fatalf("Expected error for root type of program but actually '%s'", msg)
	}
}

func TestTypeCheckFail(t *testing.T) {
	s := token.NewDummySource("let x = 42 in x +. 3.14")
	l := lexer.NewLexer(s)
	go l.Lex()
	root, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(root)
	}

	env := NewEnv()
	err = env.ApplyTypeAnalysis(root)
	if err == nil {
		t.Fatalf("Type check must raise a type error")
	}
}
