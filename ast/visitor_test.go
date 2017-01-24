package ast

import (
	"github.com/rhysd/gocaml/token"
	"testing"
)

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
	tok := &token.Token{}
	e := &Let{
		LetToken: tok,
		Decl: Decl{
			Name: "test",
		},
		Bound: &Int{
			Token: tok,
			Value: 42,
		},
		Body: &Add{
			Left: &Var{
				Token: tok,
				Ident: "test",
			},
			Right: &Float{
				Token: tok,
				Value: 3.14,
			},
		},
	}
	v := &testNumAllNodes{0}
	Visit(v, e)
	if v.total != 5 {
		t.Fatalf("5 is expected as total nodes but actually %d", v.total)
	}
}

func TestVisitorCancelVisit(t *testing.T) {
	tok := &token.Token{}
	e := &Let{
		LetToken: tok,
		Decl: Decl{
			Name: "test",
		},
		Bound: &Int{
			Token: tok,
			Value: 42,
		},
		Body: &Add{
			Left: &Var{
				Token: tok,
				Ident: "test",
			},
			Right: &Float{
				Token: tok,
				Value: 3.14,
			},
		},
	}
	v := &testNumRootChildren{0, false}
	Visit(v, e)
	if v.numChildren != 3 {
		t.Fatalf("3 is expected as number of root children but actually %d", v.numChildren)
	}
}
