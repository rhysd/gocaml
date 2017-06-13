package sema

import (
	"github.com/rhysd/gocaml/ast"
	"testing"
)

func TestFindSymbol(t *testing.T) {
	s := newScope(nil)
	foo := ast.NewSymbol("foo")
	foo.Name = "foo1"
	s.mapSymbol("foo", foo)
	sym, ok := s.resolve("foo")
	if !ok {
		t.Errorf("symbol for current scope not found")
	}
	if sym.Name != "foo1" {
		t.Errorf("expected foo1 but actually %s", sym.Name)
	}
}

func TestFindNestedSymbol(t *testing.T) {
	s := newScope(nil)
	foo := ast.NewSymbol("foo")
	foo.Name = "foo1"
	s.mapSymbol("foo", foo)
	s.mapSymbol("bar", ast.NewSymbol("bar"))

	s = newScope(s)
	foo2 := ast.NewSymbol("foo")
	foo2.Name = "foo2"
	s.mapSymbol("foo", foo2)

	sym, ok := s.resolve("foo")
	if !ok {
		t.Errorf("symbol for current scope not found")
	}
	if sym.Name != "foo2" {
		t.Errorf("expected foo2 but actually %s", sym.Name)
	}

	sym, ok = s.resolve("bar")
	if !ok {
		t.Errorf("symbol for current scope not found")
	}
	if sym.Name != "bar" {
		t.Errorf("expected bar but actually %s", sym.Name)
	}

	if sym, ok = s.resolve("piyo"); ok {
		t.Errorf("symbol piyo should not be found but actually %v was found", sym)
	}
}
