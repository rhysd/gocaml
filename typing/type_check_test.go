package typing

import (
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvedSymbols(t *testing.T) {
	s := locerr.NewDummySource("let x = 1 in x + y; ()")
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	env, err := TypeCheck(ast)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := env.Table["x"]; !ok {
		t.Error("'x' was not resolved as internal symbol:", env.Table)
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

			l := lexer.NewLexer(s)
			go l.Lex()

			ast, err := parser.Parse(l.Tokens)
			if err != nil {
				panic(ast.Root)
			}

			if err = alpha.Transform(ast.Root); err != nil {
				panic(err)
			}

			_, err = TypeCheck(ast)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProgramRootTypeIsUnit(t *testing.T) {
	s := locerr.NewDummySource("42")
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	_, err = TypeCheck(ast)
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
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	_, err = TypeCheck(ast)
	if err == nil {
		t.Fatalf("Type check must raise a type error")
	}
}

func TestDerefNoneTypes(t *testing.T) {
	s := locerr.NewDummySource("let rec f x = () in f (Some 42); f None; let a = None in f a")
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	env, err := TypeCheck(ast)
	if err != nil {
		t.Fatal(err)
	}

	if len(env.TypeHints) != 2 {
		t.Fatal("None type values were not detected")
	}

	for _, h := range env.TypeHints {
		v, ok := h.(*types.Option).Elem.(*types.Var)
		if ok {
			t.Errorf("Element type of 'None' value was not dereferenced: %s", v.String())
		}
	}
}

func TestDerefEmptyArray(t *testing.T) {
	s := locerr.NewDummySource("let a = [| |] in println_int a.(0)")
	l := lexer.NewLexer(s)
	go l.Lex()
	ast, err := parser.Parse(l.Tokens)
	if err != nil {
		panic(ast.Root)
	}

	env, err := TypeCheck(ast)
	if err != nil {
		t.Fatal(err)
	}

	if len(env.TypeHints) != 1 {
		t.Fatal("Empty array value was not detected")
	}

	for _, h := range env.TypeHints {
		v, ok := h.(*types.Array).Elem.(*types.Var)
		if ok {
			t.Errorf("Element type of 'None' value was not dereferenced: %s", v.String())
		}
	}
}
