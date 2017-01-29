package alpha

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
)

// Alpha transform.
// Transform all idenfiers to unique ones.
//
// Example:
//   Before: let x = 1 in let x = 2 in x + x
//   After:  let x$1 = 1 in let x$2 = 2 in x$2 + x$2

func duplicateSymbol(symbols []*ast.Symbol) *ast.Symbol {
	len := len(symbols)
	for i, left := range symbols {
		for _, right := range symbols[i+1 : len] {
			if left.DisplayName == right.DisplayName {
				return left
			}
		}
	}
	return nil
}

type transformer struct {
	current *mapping
	count   uint
	err     error
}

func newTransformer() *transformer {
	return &transformer{
		current: newMapping(nil),
		count:   0,
		err:     nil,
	}
}

func (t *transformer) setDuplicateError(node ast.Expr, name string) {
	pos := node.Pos()
	t.err = fmt.Errorf("Detected duplicate symbol '%s' in node '%s' at (line:%d, column:%d)", name, node.Name(), pos.Line, pos.Column)
}

func (t *transformer) newID(n string) string {
	t.count++
	return fmt.Sprintf("%s$t%d", n, t.count)
}

func (t *transformer) register(s *ast.Symbol) {
	s.Name = t.newID(s.DisplayName)
	t.current.vars[s.DisplayName] = s
}

func (t *transformer) nest() {
	t.current = newMapping(t.current)
}

func (t *transformer) pop() {
	t.current = t.current.parent
}

func (t *transformer) Visit(node ast.Expr) ast.Visitor {
	switch n := node.(type) {
	case *ast.Let:
		// At first, transform value bound to the variable
		ast.Visit(t, n.Bound)
		t.nest()
		t.register(n.Symbol)
		ast.Visit(t, n.Body)
		t.pop()
		return nil
	case *ast.LetRec:
		if s := duplicateSymbol(n.Func.Params); s != nil {
			t.setDuplicateError(n, s.DisplayName)
			return nil
		}
		t.nest()
		t.register(n.Func.Symbol)
		t.nest()
		for _, p := range n.Func.Params {
			t.register(p)
		}
		ast.Visit(t, n.Func.Body)
		t.pop() // Pop parameters scope
		ast.Visit(t, n.Body)
		t.pop() // Pop function scope
		return nil
	case *ast.LetTuple:
		ast.Visit(t, n.Bound)
		if s := duplicateSymbol(n.Symbols); s != nil {
			t.setDuplicateError(n, s.DisplayName)
			return nil
		}
		t.nest()
		for _, e := range n.Symbols {
			t.register(e)
		}
		ast.Visit(t, n.Body)
		t.pop()
		return nil
	case *ast.VarRef:
		mapped, ok := t.current.resolve(n.Symbol.DisplayName)
		if !ok {
			// External symbol is ignored because name should be identical.
			return nil
		}
		n.Symbol = mapped
		return nil
	}

	// Visit recursively
	return t
}

func Transform(root ast.Expr) error {
	v := newTransformer()
	ast.Visit(v, root)
	return v.err
}
