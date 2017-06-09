// Package typing provides type inference and type check for GoCaml.
// This package only provides type operations. To know data structures of types, please see
// https://godoc.org/github.com/rhysd/gocaml/types
package typing

import (
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
)

// TypeCheck applies type inference, checks semantics of types and finally converts AST into MIR
// with inferred type information.
func TypeCheck(parsed *ast.AST) (*types.Env, *mir.Block, error) {
	// First, run unification on all nodes and dereference type variables
	inferer := NewInferer()
	if err := inferer.Infer(parsed); err != nil {
		return nil, nil, locerr.NoteAt(parsed.Root.Pos(), err, "Type inference failed")
	}

	// TODO:
	// Type dereference should be done with generating MIR

	// Second, convert AST into MIR
	block, err := ToMIR(parsed.Root, inferer.Env)
	if err != nil {
		return nil, nil, locerr.NoteAt(parsed.Root.Pos(), err, "AST to MIR conversion failed")
	}

	return inferer.Env, block, nil
}
