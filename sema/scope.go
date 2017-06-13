package sema

import (
	"github.com/rhysd/gocaml/ast"
)

type scope struct {
	parent *scope
	vars   map[string]*ast.Symbol
}

func newScope(parent *scope) *scope {
	return &scope{
		parent,
		map[string]*ast.Symbol{},
	}
}

func (m *scope) mapSymbol(from string, to *ast.Symbol) {
	m.vars[from] = to
}

func (m *scope) resolve(name string) (*ast.Symbol, bool) {
	if mapped, ok := m.vars[name]; ok {
		return mapped, true
	}
	if m.parent == nil {
		return nil, false
	}
	return m.parent.resolve(name)
}
