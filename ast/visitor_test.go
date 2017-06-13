package ast

import (
	"github.com/rhysd/gocaml/token"
	"testing"
)

var testTree = &Let{
	LetToken: &token.Token{},
	Symbol:   NewSymbol("test"),
	Bound: &Int{
		Token: &token.Token{},
		Value: 42,
	},
	Body: &Add{
		Left: &VarRef{
			Token:  &token.Token{},
			Symbol: NewSymbol("test"),
		},
		Right: &Float{
			Token: &token.Token{},
			Value: 3.14,
		},
	},
}

type testNumAllNodes struct {
	tdTotal int
	buTotal int
}

func (v *testNumAllNodes) VisitTopdown(e Expr) Visitor {
	v.tdTotal++
	return v
}

func (v *testNumAllNodes) VisitBottomup(e Expr) {
	v.buTotal++
}

type testNumRootChildren struct {
	numChildren int
	rootVisited bool
}

func (v *testNumRootChildren) VisitTopdown(e Expr) Visitor {
	v.numChildren++
	if v.rootVisited {
		return nil
	}
	v.rootVisited = true
	return v
}

func (v *testNumRootChildren) VisitBottomup(Expr) {
}

func TestVisitorVisit(t *testing.T) {
	v := &testNumAllNodes{0, 0}
	Visit(v, testTree)
	if v.tdTotal != 5 {
		t.Fatalf("5 is expected as total nodes but actually %d", v.tdTotal)
	}
}

func TestVisitorCancelVisit(t *testing.T) {
	v := &testNumRootChildren{0, false}
	Visit(v, testTree)
	if v.numChildren != 3 {
		t.Fatalf("3 is expected as number of root children but actually %d", v.numChildren)
	}
}
