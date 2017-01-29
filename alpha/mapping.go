package alpha

import (
	"github.com/rhysd/gocaml/ast"
)

type mapping struct {
	parent *mapping
	vars   map[string]*ast.Symbol
}

func newMapping(parent *mapping) *mapping {
	return &mapping{
		parent,
		map[string]*ast.Symbol{},
	}
}

func (m *mapping) add(from string, to *ast.Symbol) {
	m.vars[from] = to
}

func (m *mapping) resolve(name string) (*ast.Symbol, bool) {
	if mapped, ok := m.vars[name]; ok {
		return mapped, true
	}
	if m.parent == nil {
		return nil, false
	}
	return m.parent.resolve(name)
}
