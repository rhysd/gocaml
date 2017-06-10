package sema

import (
	"github.com/rhysd/gocaml/ast"
	"testing"
)

func TestFindSymbol(t *testing.T) {
	m := newMapping(nil)
	foo := ast.NewSymbol("foo")
	foo.Name = "foo1"
	m.add("foo", foo)
	s, ok := m.resolve("foo")
	if !ok {
		t.Errorf("symbol for current scope not found")
	}
	if s.Name != "foo1" {
		t.Errorf("expected foo1 but actually %s", s.Name)
	}
}

func TestFindNestedSymbol(t *testing.T) {
	m := newMapping(nil)
	foo := ast.NewSymbol("foo")
	foo.Name = "foo1"
	m.add("foo", foo)
	m.add("bar", ast.NewSymbol("bar"))

	m = newMapping(m)
	foo2 := ast.NewSymbol("foo")
	foo2.Name = "foo2"
	m.add("foo", foo2)

	s, ok := m.resolve("foo")
	if !ok {
		t.Errorf("symbol for current scope not found")
	}
	if s.Name != "foo2" {
		t.Errorf("expected foo2 but actually %s", s.Name)
	}

	s, ok = m.resolve("bar")
	if !ok {
		t.Errorf("symbol for current scope not found")
	}
	if s.Name != "bar" {
		t.Errorf("expected bar but actually %s", s.Name)
	}

	if s, ok = m.resolve("piyo"); ok {
		t.Errorf("symbol piyo should not be found but actually %v was found", s)
	}
}
