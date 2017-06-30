// Package types provides data structures for types in GoCaml.
package types

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
)

// Instantiation is the information of instantiation of a generic type.
type Instantiation struct {
	// From is a generic type variable instantiated.
	From Type
	// To is a type variable instantiated from generic type variable.
	To Type
	// Mapping from ID of generic type variable to actual instantiated type variable
	Mapping map[VarID]Type
}

// Result of type analysis.
type Env struct {
	// Types for declarations. This is referred by type variables to resolve
	// type variables' actual types
	//
	// XXX:
	// Currently nested identifiers don't work. Example:
	//   let
	//     x = 42
	//   in
	//     let x = true in print_bool (x);
	//     print_int (x)
	// We need alpha transform before type inference in order to ensure
	// all symbol names are unique.
	Table map[string]Type
	// External variable names which are referred but not defined.
	// External variables are exposed as external symbols in other object files.
	Externals map[string]Type
	// GoCaml uses let-polymorphic type inference. It means that instantiation occurs when new
	// symbol is introduced. So instantiation only occurs at variable reference.
	Instantiations map[*ast.VarRef]*Instantiation
}

// NewEnv creates empty Env instance.
func NewEnv() *Env {
	return &Env{
		map[string]Type{},
		builtinPopulatedTable(),
		map[*ast.VarRef]*Instantiation{},
	}
}

// TODO: Dump environment as JSON

func (env *Env) Dump() {
	env.DumpVariables()
	fmt.Println()
	env.DumpInstantiations()
	fmt.Println()
	env.DumpExternals()
}

func (env *Env) DumpVariables() {
	fmt.Println("Variables:")
	for s, t := range env.Table {
		fmt.Printf("  %s: %s\n", s, t.String())
	}
}

func (env *Env) DumpExternals() {
	fmt.Println("External Variables:")
	for s, t := range env.Externals {
		fmt.Printf("  %s: %s\n", s, t.String())
	}
}

func (env *Env) DumpInstantiations() {
	fmt.Println("Instantiations:")
	for ref, inst := range env.Instantiations {
		fmt.Printf("  '%s' at %s\n", ref.Symbol.DisplayName, ref.Pos().String())
		fmt.Printf("    From: %s\n", inst.From.String())
		fmt.Printf("    To:   %s\n", inst.To.String())
	}
}

func (env *Env) DumpDebug() {
	fmt.Println("Variables:")
	for s, t := range env.Table {
		fmt.Printf("  %s: %s\n", s, Debug(t))
	}
	fmt.Println("\nInstantiations:")
	for ref, inst := range env.Instantiations {
		fmt.Printf("  '%s' at %s\n", ref.Symbol.Name, ref.Pos().String())
		fmt.Printf("    From: %s\n", Debug(inst.From))
		fmt.Printf("    To:   %s\n", Debug(inst.To))
		for id, free := range inst.Mapping {
			fmt.Printf("      VAR %d: => '%s'\n", id, free.String())
		}
	}
	fmt.Println()
	env.DumpExternals()
}
