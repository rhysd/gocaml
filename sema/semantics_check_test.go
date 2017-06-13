package sema

import (
	"github.com/rhysd/gocaml/syntax"
	"github.com/rhysd/locerr"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvedSymbols(t *testing.T) {
	s := locerr.NewDummySource("let x = 1 in x + y; ()")
	ast, err := syntax.Parse(s)
	if err != nil {
		panic(ast.Root)
	}

	env, _, err := SemanticsCheck(ast)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := env.Table["x"]; ok {
		t.Error("'x' was resolved as internal symbol:", env.Table)
	}
	if _, ok := env.Externals["y"]; !ok {
		t.Error("'y' was not resolved as external symbol:", env.Externals)
	}
}

func TestTypeCheckMinCamlTests(t *testing.T) {
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

		t.Run("from-mincaml:"+n, func(t *testing.T) {
			s, err := locerr.NewSourceFromFile(n)
			if err != nil {
				panic(err)
			}

			ast, err := syntax.Parse(s)
			if err != nil {
				panic(ast.Root)
			}

			_, _, err = SemanticsCheck(ast)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProgramRootTypeIsUnit(t *testing.T) {
	s := locerr.NewDummySource("42")
	ast, err := syntax.Parse(s)
	if err != nil {
		panic(ast.Root)
	}

	_, _, err = SemanticsCheck(ast)
	if err == nil {
		t.Fatalf("Type check must raise an error when root type of program is not ()")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Type of root expression of program must be unit") {
		t.Fatalf("Expected error for root type of program but actually '%s'", msg)
	}
}

func TestTypeCheckFail(t *testing.T) {
	s := locerr.NewDummySource("let x = 42 in x +. 3.14")
	ast, err := syntax.Parse(s)
	if err != nil {
		panic(ast.Root)
	}

	_, _, err = SemanticsCheck(ast)
	if err == nil {
		t.Fatalf("Type check must raise a type error")
	}
}
