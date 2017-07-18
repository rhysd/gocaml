// Package types provides data structures for types in GoCaml.
package types

import (
	"fmt"
)

type VarMapping struct {
	ID   VarID
	Type Type
}

// Instantiation is the information of instantiation of a generic type.
type Instantiation struct {
	// From is a generic type variable instantiated.
	From Type
	// To is a type variable instantiated from generic type variable.
	To Type
	// Mapping from ID of generic type variable to actual instantiated type variable
	Mapping []*VarMapping
}

type External struct {
	Type  Type
	CName string
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
	DeclTable map[string]Type
	// External variable names which are referred but not defined.
	// External variables are exposed as external symbols in other object files.
	Externals map[string]*External
	// GoCaml uses let-polymorphic type inference. It means that instantiation occurs when new
	// symbol is introduced. So instantiation only occurs at variable reference.
	RefInsts map[string]*Instantiation
	// Mappings from generic type to instantiated types for each declarations.
	// e.g.
	//   'a -> 'a => {int -> int, bool -> bool, float -> float}
	//
	// Note: This is set in sema/deref.go
	PolyTypes map[Type][]*Instantiation
}

// NewEnv creates empty Env instance.
func NewEnv() *Env {
	return &Env{
		map[string]Type{},
		builtinPopulatedTable(),
		map[string]*Instantiation{},
		nil,
	}
}

// TODO: Dump environment as JSON

func (env *Env) Dump() {
	// Note: RefInsts is not displayed because it is filled by ToMIR conversion function and not
	// filled by the type analysis.
	env.DumpVariables()
	fmt.Println()
	env.DumpPolyTypes()
	fmt.Println()
	env.DumpExternals()
}

func (env *Env) DumpVariables() {
	fmt.Println("Variables:")
	for s, t := range env.DeclTable {
		fmt.Printf("  %s: %s\n", s, t.String())
	}
}

func (env *Env) DumpExternals() {
	fmt.Println("External Variables:")
	for s, e := range env.Externals {
		fmt.Printf("  %s: %s (=> %s)\n", s, e.Type.String(), e.CName)
	}
}

func (env *Env) DumpPolyTypes() {
	fmt.Println("PolyTypes:")
	for t, insts := range env.PolyTypes {
		fmt.Printf("  '%s' (%d instances) =>\n", t.String(), len(insts))
		for i, inst := range insts {
			fmt.Printf("    %d: %s\n", i, inst.To.String())
		}
	}
}

func (env *Env) DumpDebug() {
	fmt.Println("Variables:")
	for s, t := range env.DeclTable {
		fmt.Printf("  %s: %s\n", s, Debug(t))
	}
	fmt.Println("\nInstantiations:")
	for ref, inst := range env.RefInsts {
		fmt.Printf("  '%s'\n", ref)
		fmt.Printf("    From: %s\n", Debug(inst.From))
		fmt.Printf("    To:   %s\n", Debug(inst.To))
		for i, m := range inst.Mapping {
			fmt.Printf("      VAR%d: '%d => '%s'\n", i, m.ID, Debug(m.Type))
		}
	}
	fmt.Println()
	fmt.Println("PolyTypes:")
	for t, insts := range env.PolyTypes {
		fmt.Printf("  '%s' (%d instance(s)) =>\n", Debug(t), len(insts))
		for i, inst := range insts {
			fmt.Printf("    %d: %s\n", i, Debug(inst.To))
		}
	}
	fmt.Println()
	env.DumpExternals()
}
