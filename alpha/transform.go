// Package alpha provides alpha transformation for parsed GoCaml AST.
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
		if left.IsIgnored() {
			continue
		}
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
	if s.IsIgnored() {
		return
	}
	s.Name = t.newID(s.DisplayName)
	t.current.add(s.DisplayName, s)
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
		if s := duplicateSymbol(n.Func.ParamSymbols()); s != nil {
			t.setDuplicateError(n, s.DisplayName)
			return nil
		}
		t.nest()
		t.register(n.Func.Symbol)
		t.nest()
		for _, p := range n.Func.Params {
			t.register(p.Ident)
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
	case *ast.Match:
		ast.Visit(t, n.Target)
		t.nest()
		t.register(n.SomeIdent)
		ast.Visit(t, n.IfSome)
		t.pop()
		ast.Visit(t, n.IfNone)
		return nil
	case *ast.VarRef:
		if n.Symbol.DisplayName == "_" {
			// Note: Check '_'. Without this check, compiler will consdier it as
			// external variable wrongly.
			p := n.Pos()
			t.err = fmt.Errorf("Cannot refer '_' variable at (line:%d, column:%d) because creating '_' variable is not permitted", p.Line, p.Column)
			return nil
		}
		mapped, ok := t.current.resolve(n.Symbol.DisplayName)
		if !ok {
			// External symbol is ignored because name should be identical.
			return nil
		}
		n.Symbol = mapped
		return nil
	default:
		// Visit recursively
		return t
	}
}

// Transform adds identical names to all identifiers in AST nodes.
// If there are some duplicate names, it causes an error.
// External symbols are named the same as display names.
func Transform(root ast.Expr) error {
	v := newTransformer()
	ast.Visit(v, root)
	return v.err
}
