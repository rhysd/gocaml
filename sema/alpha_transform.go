package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
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

func isBuiltinTypeCtor(name string) bool {
	switch name {
	case "_", "array", "option", "unit", "int", "bool", "float", "string":
		return true
	default:
		return false
	}
}

type transformer struct {
	current   *scope
	typeScope *scope
	varId     uint
	tyId      uint
	err       error
	externals map[string]struct{}
}

func newTransformer() *transformer {
	return &transformer{
		current:   newScope(nil),
		typeScope: newScope(nil),
		varId:     0,
		tyId:      0,
		externals: nil,
	}
}

func (t *transformer) duplicateError(node ast.Expr, name string) {
	t.err = locerr.ErrorfIn(node.Pos(), node.End(), "Detected duplicate symbol '%s'", name)
}

func (t *transformer) newVarID(n string) string {
	t.varId++
	return fmt.Sprintf("%s$t%d", n, t.varId)
}

func (t *transformer) newTyID(n string) string {
	t.tyId++
	return fmt.Sprintf("%s.t%d", n, t.tyId)
}

func (t *transformer) register(s *ast.Symbol) {
	if s.IsIgnored() {
		return
	}
	s.Name = t.newVarID(s.DisplayName)
	t.current.mapSymbol(s.DisplayName, s)
}

func (t *transformer) nest() {
	t.current = newScope(t.current)
}

func (t *transformer) pop() {
	t.current = t.current.parent
}

func (t *transformer) VisitTopdown(node ast.Expr) ast.Visitor {
	switch n := node.(type) {
	case *ast.Let:
		// At first, transform value bound to the variable
		ast.Visit(t, n.Bound)
		if n.Type != nil {
			ast.Visit(t, n.Type)
		}
		t.nest()
		t.register(n.Symbol)
		ast.Visit(t, n.Body)
		t.pop()
		return nil
	case *ast.LetRec:
		if s := duplicateSymbol(n.Func.ParamSymbols()); s != nil {
			t.duplicateError(n, s.DisplayName)
			return nil
		}
		t.nest()
		t.register(n.Func.Symbol)
		t.nest()
		for _, p := range n.Func.Params {
			if p.Type != nil {
				ast.Visit(t, p.Type)
			}
			t.register(p.Ident)
		}
		if n.Func.RetType != nil {
			ast.Visit(t, n.Func.RetType)
		}
		ast.Visit(t, n.Func.Body)
		t.pop() // Pop parameters scope
		ast.Visit(t, n.Body)
		t.pop() // Pop function scope
		return nil
	case *ast.LetTuple:
		if n.Type != nil {
			ast.Visit(t, n.Type)
		}
		ast.Visit(t, n.Bound)
		if s := duplicateSymbol(n.Symbols); s != nil {
			t.duplicateError(n, s.DisplayName)
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
			t.err = locerr.ErrorIn(n.Pos(), n.End(), "Cannot refer '_' variable because creating '_' variable is not permitted")
			return nil
		}
		if mapped, ok := t.current.resolve(n.Symbol.DisplayName); ok {
			n.Symbol = mapped
			return nil
		}
		// Check external it's an external symbol
		if _, ok := t.externals[n.Symbol.Name]; !ok {
			t.err = locerr.ErrorfIn(n.Pos(), n.End(), "Undefined variable '%s'", n.Symbol.DisplayName)
		}
		return nil
	case *ast.CtorType:
		if isBuiltinTypeCtor(n.Ctor.DisplayName) {
			// '_' or other builtin types such as 'int' should not be alpha-transformed and handled as-is.
			return t
		}
		mapped, ok := t.typeScope.resolve(n.Ctor.DisplayName)
		if !ok {
			t.err = locerr.ErrorfIn(n.Pos(), n.End(), "Undefined type name '%s'", n.Ctor.DisplayName)
			return nil
		}
		n.Ctor = mapped
		return t
	default:
		// Visit recursively
		return t
	}
}

func (t *transformer) VisitBottomup(ast.Expr) {
	return
}

// AlphaTransform adds identical names to all identifiers in AST nodes.
// If there are some duplicate names, it causes an error.
// External symbols are named the same as display names.
func AlphaTransform(tree *ast.AST, env *types.Env) error {
	v := newTransformer()
	for _, decl := range tree.TypeDecls {
		ast.Visit(v, decl.Type)
		if v.err != nil {
			return v.err
		}

		i := decl.Ident
		if isBuiltinTypeCtor(i.DisplayName) {
			return locerr.ErrorfIn(decl.Pos(), decl.End(), "Cannot redefine built-in type '%s'", i.DisplayName)
		}

		// Note: Overwrite previous type mapping if already existing
		i.Name = v.newTyID(i.DisplayName)
		v.typeScope.mapSymbol(i.DisplayName, i)
	}

	exts := make(map[string]struct{}, len(tree.Externals)+len(env.Externals))
	cnames := make(map[string]struct{}, len(tree.Externals)+len(env.Externals))
	// Register built-in external symbols
	for n, e := range env.Externals {
		exts[n] = struct{}{}
		cnames[e.CName] = struct{}{}
	}
	// Register declared external symbols
	for _, e := range tree.Externals {
		if e.Ident.IsIgnored() {
			return locerr.ErrorIn(e.Pos(), e.End(), "Cannot define external symbol as '_'")
		}
		exts[e.Ident.Name] = struct{}{}
		if _, ok := cnames[e.C]; ok {
			return locerr.ErrorfIn(e.Pos(), e.End(), "Cannot redeclare existing C symbol '%s'", e.C)
		}
		cnames[e.C] = struct{}{}
	}
	v.externals = exts

	ast.Visit(v, tree.Root)
	return v.err
}
