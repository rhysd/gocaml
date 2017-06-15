// Package sema provides resolving symbols, type inference and type check for GoCaml.
// Semantic check finally converts given AST into MIR (Mid-level IR).
// This package only provides type operations. To know data structures of types, please see
// https://godoc.org/github.com/rhysd/gocaml/types
package sema

import (
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
)

func Analyze(parsed *ast.AST) (*types.Env, InferredTypes, error) {
	// First, resolve all symbols by alpha transform
	if err := AlphaTransform(parsed); err != nil {
		return nil, nil, locerr.NoteAt(parsed.Root.Pos(), err, "Alpha transform failed")
	}

	// Second, run unification on all nodes and dereference type variables
	inferer := NewInferer()
	if err := inferer.Infer(parsed); err != nil {
		return nil, nil, locerr.NoteAt(parsed.Root.Pos(), err, "Type inference failed")
	}

	return inferer.Env, inferer.inferred, nil
}

// SemanticsCheck applies type inference, checks semantics of types and finally converts AST into MIR
// with inferred type information.
func SemanticsCheck(parsed *ast.AST) (*types.Env, *mir.Block, error) {
	env, inferred, err := Analyze(parsed)
	if err != nil {
		return nil, nil, err
	}

	// Third, convert AST into MIR
	block := ToMIR(parsed.Root, env, inferred)

	return env, block, nil
}
