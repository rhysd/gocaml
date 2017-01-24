package ast

import (
	"github.com/rhysd/gocaml/token"
	"testing"
)

var testTree = &Let{
	LetToken: &token.Token{},
	Decl: Decl{
		Name: "test",
	},
	Bound: &Int{
		Token: &token.Token{},
		Value: 42,
	},
	Body: &Add{
		Left: &Var{
			Token: &token.Token{},
			Ident: "test",
		},
		Right: &Float{
			Token: &token.Token{},
			Value: 3.14,
		},
	},
}

type testNumAllNodes struct {
	total int
}

func (v *testNumAllNodes) Visit(e Expr) Visitor {
	v.total += 1
	return v
}

type testNumRootChildren struct {
	numChildren int
	rootVisited bool
}

func (v *testNumRootChildren) Visit(e Expr) Visitor {
	v.numChildren += 1
	if v.rootVisited {
		return nil
	}
	v.rootVisited = true
	return v
}

func TestVisitorVisit(t *testing.T) {
	v := &testNumAllNodes{0}
	Visit(v, testTree)
	if v.total != 5 {
		t.Fatalf("5 is expected as total nodes but actually %d", v.total)
	}
}

func TestVisitorCancelVisit(t *testing.T) {
	v := &testNumRootChildren{0, false}
	Visit(v, testTree)
	if v.numChildren != 3 {
		t.Fatalf("3 is expected as number of root children but actually %d", v.numChildren)
	}
}

func TestVisitorFind(t *testing.T) {
	found := Find(testTree, func(e Expr) bool {
		return e.Name() == "Var"
	})
	if !found {
		t.Errorf("'Var' node was not found for test AST")
	}
	found = Find(testTree, func(e Expr) bool {
		_, ok := e.(*Not)
		return ok
	})
	if found {
		t.Errorf("'Not' node is not in test AST but was found by ast.Find()")
	}
}

func TestVisitorVisitChildren(t *testing.T) {
	names := []string{"Int", "Add"}
	index := 0
	VisitChildren(testTree, func(child Expr) {
		if names[index] != child.Name() {
			t.Errorf("Expected child name %s but actually %s", names[index], child.Name())
		}
		index += 1
	})
}
