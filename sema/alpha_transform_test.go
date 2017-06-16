package sema

import (
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/locerr"
	"strings"
	"testing"
)

func TestFlatScope(t *testing.T) {
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
	ref := &ast.VarRef{
		tok,
		ast.NewSymbol("test"),
	}
	root := &ast.Let{
		tok,
		ast.NewSymbol("test"),
		&ast.Int{nil, 42},
		ref,
		nil,
	}
	if err := AlphaTransform(&ast.AST{Root: root}); err != nil {
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
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
	ref := &ast.VarRef{
		tok,
		ast.NewSymbol("test"),
	}
	child := &ast.Let{
		tok,
		ast.NewSymbol("test"),
		&ast.Int{nil, 42},
		ref,
		nil,
	}
	root := &ast.Let{
		tok,
		ast.NewSymbol("test"),
		&ast.Int{nil, 42},
		child,
		nil,
	}

	if err := AlphaTransform(&ast.AST{Root: root}); err != nil {
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
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
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
		locerr.Pos{},
	}
	root := &ast.Let{
		tok, ast.NewSymbol("a"),
		&ast.Int{tok, 42},
		match,
		nil,
	}

	if err := AlphaTransform(&ast.AST{Root: root}); err != nil {
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

	if err := AlphaTransform(&ast.AST{Root: root}); err != nil {
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
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
	root := &ast.LetTuple{
		tok,
		[]*ast.Symbol{
			ast.NewSymbol("a"),
			ast.NewSymbol("b"),
			ast.NewSymbol("b"),
		},
		&ast.Int{tok, 42},
		&ast.Int{tok, 42},
		nil,
	}

	if err := AlphaTransform(&ast.AST{Root: root}); err == nil {
		t.Fatalf("LetTuple contains duplicate symbols but error did not occur")
	}
}

func TestLetRec(t *testing.T) {
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
	ref := &ast.VarRef{
		tok,
		ast.NewSymbol("f"),
	}
	ref2 := &ast.VarRef{
		tok,
		ast.NewSymbol("b"),
	}
	root := &ast.LetRec{
		tok,
		&ast.FuncDef{
			ast.NewSymbol("f"),
			[]ast.Param{
				{ast.NewSymbol("a"), nil},
				{ast.NewSymbol("b"), nil},
				{ast.NewSymbol("c"), nil},
			},
			ref2,
			nil,
		},
		ref,
	}

	if err := AlphaTransform(&ast.AST{Root: root}); err != nil {
		t.Fatal(err)
	}

	expects := []string{"a$t2", "b$t3", "c$t4"}
	for i, p := range root.Func.Params {
		if p.Ident.Name != expects[i] {
			t.Errorf("Parameter should be transformed to %s but actually %s", expects[i], p.Ident.Name)
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
	if root.Func.Params[1].Ident != ref2.Symbol {
		t.Fatalf("Ref symbol should be resolved to parameter symbol")
	}
}

func TestRecursiveFunc(t *testing.T) {
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
	ref := &ast.VarRef{
		tok,
		ast.NewSymbol("f"),
	}
	root := &ast.LetRec{
		tok,
		&ast.FuncDef{
			ast.NewSymbol("f"),
			[]ast.Param{
				{ast.NewSymbol("a"), nil},
				{ast.NewSymbol("b"), nil},
				{ast.NewSymbol("c"), nil},
			},
			ref,
			nil,
		},
		&ast.Int{tok, 42},
	}

	if err := AlphaTransform(&ast.AST{Root: root}); err != nil {
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
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
	ref := &ast.VarRef{
		tok,
		ast.NewSymbol("f"),
	}
	ref2 := &ast.VarRef{
		tok,
		ast.NewSymbol("f"),
	}
	root := &ast.LetRec{
		tok,
		&ast.FuncDef{
			ast.NewSymbol("f"),
			[]ast.Param{
				{ast.NewSymbol("f"), nil},
			},
			ref,
			nil,
		},
		ref2,
	}

	if err := AlphaTransform(&ast.AST{Root: root}); err != nil {
		t.Fatal(err)
	}

	if ref.Symbol.Name != "f$t2" {
		t.Fatalf("Ref should be resolved to parameter but actually %s", ref.Symbol.Name)
	}
	if root.Func.Params[0].Ident != ref.Symbol {
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
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
	root := &ast.LetRec{
		tok,
		&ast.FuncDef{
			ast.NewSymbol("f"),
			[]ast.Param{
				{ast.NewSymbol("a"), nil},
				{ast.NewSymbol("b"), nil},
				{ast.NewSymbol("b"), nil},
			},
			&ast.Int{tok, 42},
			nil,
		},
		&ast.Int{tok, 42},
	}

	if err := AlphaTransform(&ast.AST{Root: root}); err == nil {
		t.Fatal("Duplicate in parameters must raise an error")
	}
}

func TestExternalSymbol(t *testing.T) {
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
	ref := &ast.VarRef{
		tok,
		ast.NewSymbol("x"),
	}

	if err := AlphaTransform(&ast.AST{Root: ref}); err != nil {
		t.Fatal(err)
	}

	if ref.Symbol.Name != ref.Symbol.DisplayName {
		t.Fatalf("External symbol's name should not be changed but actually %s was changed to %s", ref.Symbol.DisplayName, ref.Symbol.Name)
	}
}

func TestUnderscoreName(t *testing.T) {
	tok := &token.Token{
		Start: locerr.Pos{},
		End:   locerr.Pos{},
	}
	ref := &ast.VarRef{
		tok,
		ast.NewSymbol("_"),
	}
	err := AlphaTransform(&ast.AST{Root: ref})
	if err == nil {
		t.Fatal("Error was expected")
	}
	if !strings.Contains(err.Error(), "Cannot refer '_' variable") {
		t.Fatal("Unexpected error for '_' variable reference:", err)
	}
}

func TestInvalidTypeAlias(t *testing.T) {
	pos := locerr.Pos{}
	tok := &token.Token{
		Start: pos,
		End:   pos,
		File:  locerr.NewDummySource(""),
	}
	prim := func(name string) ast.Expr {
		return &ast.CtorType{
			nil,
			tok,
			nil,
			ast.NewSymbol(name),
		}
	}

	cases := []struct {
		what  string
		types []*ast.TypeDecl
		root  ast.Expr
		err   string
	}{
		{
			what: "cannot define '_'",
			types: []*ast.TypeDecl{
				{tok, ast.NewSymbol("_"), prim("int")},
			},
			root: &ast.Unit{tok, tok},
			err:  "Cannot redefine built-in type '_'",
		},
		{
			what: "cannot define primitive type",
			types: []*ast.TypeDecl{
				{tok, ast.NewSymbol("float"), prim("int")},
			},
			root: &ast.Unit{tok, tok},
			err:  "Cannot redefine built-in type 'float'",
		},
		{
			what: "undefined type name in type decls",
			types: []*ast.TypeDecl{
				{tok, ast.NewSymbol("foo"), prim("bar")},
			},
			root: &ast.Unit{tok, tok},
			err:  "Undefined type name 'bar'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.what, func(t *testing.T) {
			tree := &ast.AST{tc.root, tc.types}
			err := AlphaTransform(tree)
			if err == nil {
				t.Fatal("Error did not occur. Expected:", tc.err)
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.err) {
				t.Fatalf("Unexpected error message '%s'. '%s' should be contained", msg, tc.err)
			}
		})
	}
}

func TestTypeAlias(t *testing.T) {
	pos := locerr.Pos{}
	tok := &token.Token{
		Start: pos,
		End:   pos,
		File:  locerr.NewDummySource(""),
	}
	prim := func(sym *ast.Symbol) *ast.CtorType {
		return &ast.CtorType{
			nil,
			tok,
			nil,
			sym,
		}
	}

	foo := ast.NewSymbol("foo")
	bar := ast.NewSymbol("bar")
	primitive := ast.NewSymbol("int")
	anyTy := prim(ast.NewSymbol("_"))
	ty1 := prim(ast.NewSymbol("foo"))

	root := &ast.Let{
		tok,
		ast.NewSymbol("foo"),
		&ast.Unit{},
		&ast.Let{
			tok,
			ast.NewSymbol("bar"),
			&ast.Unit{},
			&ast.Unit{},
			anyTy,
		},
		ty1,
	}

	ty2 := prim(ast.NewSymbol("foo"))
	decls := []*ast.TypeDecl{
		{tok, foo, prim(primitive)},
		{tok, bar, ty2},
	}

	tree := &ast.AST{root, decls}

	if err := AlphaTransform(tree); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(foo.Name, "foo.t") {
		t.Fatal("Unexpected symbol name: ", foo.Name)
	}
	if ty1.Ctor != foo {
		t.Fatal("Failed to refer 'foo' at let expr")
	}
	if ty2.Ctor != foo {
		t.Fatal("Failed to refer 'foo' at type decl")
	}
	if !strings.HasPrefix(bar.Name, "bar.t") {
		t.Fatal("Unexpected symbol name: ", bar.Name)
	}
	if primitive.Name != "int" {
		t.Fatal("Primitive type should not be transformed:", primitive.Name)
	}
	if anyTy.Ctor.Name != "_" {
		t.Fatal("'_' should not be transformed:", anyTy.Ctor.Name)
	}
}
