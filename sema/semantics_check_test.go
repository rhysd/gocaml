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
	s := locerr.NewDummySource(`external y: int = "c_y"; let x = 1 in x + y; ()`)
	ast, err := syntax.Parse(s)
	if err != nil {
		panic(ast.Root)
	}

	env, _, err := SemanticsCheck(ast)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := env.DeclTable["x"]; ok {
		t.Error("'x' was resolved as internal symbol:", env.DeclTable)
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
				t.Fatal(err)
			}

			_, _, err = SemanticsCheck(ast)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSemanticsCheckFail(t *testing.T) {
	cases := map[string]string{
		"alpha transform":         "let rec f a a = a in f 42 42; ()",
		"type mismatch":           "3.14 + 10",
		"invalid root expression": "42",
		"dereference failure":     "None",
	}
	for what, code := range cases {
		t.Run(what, func(t *testing.T) {
			s := locerr.NewDummySource(code)
			parsed, err := syntax.Parse(s)
			if err != nil {
				panic(err)
			}
			_, _, err = SemanticsCheck(parsed)
			if err == nil {
				t.Fatal("Semantics should fail with:", code)
			}
		})
	}
}
