package alpha

import (
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
	"testing"
)

func TestFlatScope(t *testing.T) {
	ref := &ast.VarRef{
		nil,
		ast.NewSymbol("test"),
	}
	root := &ast.Let{
		nil,
		ast.NewSymbol("test"),
		&ast.Int{nil, 42},
		ref,
		nil,
	}
	if err := Transform(root); err != nil {
		t.Fatal(err)
	}
	if ref.Symbol.Name != "test$t1" {
		t.Fatalf("VarRef's symbol was not resolved: %s", ref.Symbol.Name)
	}
	if root.Symbol != ref.Symbol {
		t.Fatalf("VarRef's symbol should be resolved to declaration's symbol")
	}
}

func TestNested(t *testing.T) {
	ref := &ast.VarRef{
		nil,
		ast.NewSymbol("test"),
	}
	child := &ast.Let{
		nil,
		ast.NewSymbol("test"),
		&ast.Int{nil, 42},
		ref,
		nil,
	}
	root := &ast.Let{
		nil,
		ast.NewSymbol("test"),
		&ast.Int{nil, 42},
		child,
		nil,
	}

	if err := Transform(root); err != nil {
		t.Fatal(err)
	}

	if child.Symbol.Name != "test$t2" {
		t.Fatalf("Symbol in let expression was not transformed: %s", child.Symbol.Name)
	}
	if ref.Symbol.Name != "test$t2" {
		t.Fatalf("VarRef's symbol was not resolved: %s", ref.Symbol.Name)
	}
	if child.Symbol != ref.Symbol {
		t.Fatalf("VarRef's symbol should be resolved to declaration's symbol")
	}
}

func TestMatch(t *testing.T) {
	tok := &token.Token{}
	someRef := &ast.VarRef{
		tok,
		ast.NewSymbol("a"),
	}
	noneRef := &ast.VarRef{
		tok,
		ast.NewSymbol("a"),
	}
	match := &ast.Match{
		tok,
		&ast.Int{tok, 42},
		someRef,
		noneRef,
		ast.NewSymbol("a"),
		token.Position{},
	}
	root := &ast.Let{
		tok, ast.NewSymbol("a"),
		&ast.Int{tok, 42},
		match,
		nil,
	}

	if err := Transform(root); err != nil {
		t.Fatal(err)
	}

	if match.SomeIdent.Name != "a$t2" {
		t.Fatalf("Symbol in match expression is not transformed correctly. Expected a$t1 but actually %s", match.SomeIdent.Name)
	}
	if someRef.Symbol.Name != "a$t2" {
		t.Errorf("Symbol in some arm must refer a$t1 but %s", someRef.Symbol.Name)
	}
	if noneRef.Symbol.Name != "a$t1" {
		t.Errorf("Symbol in none arm must refer a$t1 but %s", noneRef.Symbol.Name)
	}
}

func TestLetTuple(t *testing.T) {
	ref := &ast.VarRef{
		nil,
		ast.NewSymbol("b"),
	}
	root := &ast.LetTuple{
		nil,
		[]*ast.Symbol{
			ast.NewSymbol("a"),
			ast.NewSymbol("b"),
			ast.NewSymbol("c"),
		},
		&ast.Int{nil, 42},
		ref,
		nil,
	}

	if err := Transform(root); err != nil {
		t.Fatal(err)
	}

	expects := []string{"a$t1", "b$t2", "c$t3"}
	for i, s := range root.Symbols {
		if s.Name != expects[i] {
			t.Errorf("Variables in LetTuple was not transformed as %s: %s", expects[i], s.Name)
		}
	}
	if ref.Symbol.Name != "b$t2" {
		t.Fatalf("VarRef's symbol was not resolved: %s", ref.Symbol.Name)
	}
	if root.Symbols[1] != ref.Symbol {
		t.Fatalf("VarRef's symbol should be resolved to declaration's symbol")
	}
}

func TestLetTupleHasDuplicateName(t *testing.T) {
	root := &ast.LetTuple{
		&token.Token{},
		[]*ast.Symbol{
			ast.NewSymbol("a"),
			ast.NewSymbol("b"),
			ast.NewSymbol("b"),
		},
		&ast.Int{nil, 42},
		&ast.Int{nil, 42},
		nil,
	}

	if err := Transform(root); err == nil {
		t.Fatalf("LetTuple contains duplicate symbols but error did not occur")
	}
}

func TestLetRec(t *testing.T) {
	ref := &ast.VarRef{
		&token.Token{},
		ast.NewSymbol("f"),
	}
	ref2 := &ast.VarRef{
		&token.Token{},
		ast.NewSymbol("b"),
	}
	root := &ast.LetRec{
		&token.Token{},
		&ast.FuncDef{
			ast.NewSymbol("f"),
			[]*ast.Symbol{
				ast.NewSymbol("a"),
				ast.NewSymbol("b"),
				ast.NewSymbol("c"),
			},
			ref2,
		},
		ref,
	}

	if err := Transform(root); err != nil {
		t.Fatal(err)
	}

	expects := []string{"a$t2", "b$t3", "c$t4"}
	for i, s := range root.Func.Params {
		if s.Name != expects[i] {
			t.Errorf("Parameter should be transformed to %s but actually %s", expects[i], s.Name)
		}
	}
	if root.Func.Symbol.Name != "f$t1" {
		t.Errorf("Function name was not transformed: %s", root.Func.Symbol.Name)
	}
	if ref.Symbol.Name != "f$t1" {
		t.Fatalf("Ref should be resolved to function but actually %s", ref.Symbol.Name)
	}
	if root.Func.Symbol != ref.Symbol {
		t.Fatalf("Ref symbol should be resolved to function symbol")
	}
	if ref2.Symbol.Name != "b$t3" {
		t.Fatalf("Ref should be resolved to transformed parameter for 'b' but actually '%s'", ref2.Symbol.Name)
	}
	if root.Func.Params[1] != ref2.Symbol {
		t.Fatalf("Ref symbol should be resolved to parameter symbol")
	}
}

func TestRecursiveFunc(t *testing.T) {
	ref := &ast.VarRef{
		&token.Token{},
		ast.NewSymbol("f"),
	}
	root := &ast.LetRec{
		&token.Token{},
		&ast.FuncDef{
			ast.NewSymbol("f"),
			[]*ast.Symbol{
				ast.NewSymbol("a"),
				ast.NewSymbol("b"),
				ast.NewSymbol("c"),
			},
			ref,
		},
		&ast.Int{&token.Token{}, 42},
	}

	if err := Transform(root); err != nil {
		t.Fatal(err)
	}

	if ref.Symbol.Name != "f$t1" {
		t.Fatalf("Ref should be resolved to recursive function but actually %s", ref.Symbol.Name)
	}
	if root.Func.Symbol != ref.Symbol {
		t.Fatalf("Ref symbol should be resolved to function symbol")
	}
}

func TestFuncAndParamHaveSameName(t *testing.T) {
	ref := &ast.VarRef{
		&token.Token{},
		ast.NewSymbol("f"),
	}
	ref2 := &ast.VarRef{
		&token.Token{},
		ast.NewSymbol("f"),
	}
	root := &ast.LetRec{
		&token.Token{},
		&ast.FuncDef{
			ast.NewSymbol("f"),
			[]*ast.Symbol{
				ast.NewSymbol("f"),
			},
			ref,
		},
		ref2,
	}

	if err := Transform(root); err != nil {
		t.Fatal(err)
	}

	if ref.Symbol.Name != "f$t2" {
		t.Fatalf("Ref should be resolved to parameter but actually %s", ref.Symbol.Name)
	}
	if root.Func.Params[0] != ref.Symbol {
		t.Fatalf("Ref symbol should be resolved to parameter symbol")
	}

	if ref2.Symbol.Name != "f$t1" {
		t.Fatalf("Ref should be resolved to function but actually %s", ref2.Symbol.Name)
	}
	if root.Func.Symbol != ref2.Symbol {
		t.Fatalf("Ref symbol should be resolved to function symbol")
	}
}

func TestParamDuplicate(t *testing.T) {
	root := &ast.LetRec{
		&token.Token{},
		&ast.FuncDef{
			ast.NewSymbol("f"),
			[]*ast.Symbol{
				ast.NewSymbol("a"),
				ast.NewSymbol("b"),
				ast.NewSymbol("b"),
			},
			&ast.Int{&token.Token{}, 42},
		},
		&ast.Int{&token.Token{}, 42},
	}

	if err := Transform(root); err == nil {
		t.Fatal("Duplicate in parameters must raise an error")
	}
}

func TestExternalSymbol(t *testing.T) {
	ref := &ast.VarRef{
		&token.Token{},
		ast.NewSymbol("x"),
	}

	if err := Transform(ref); err != nil {
		t.Fatal(err)
	}

	if ref.Symbol.Name != ref.Symbol.DisplayName {
		t.Fatalf("External symbol's name should not be changed but actually %s was changed to %s", ref.Symbol.DisplayName, ref.Symbol.Name)
	}
}
