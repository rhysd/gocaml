package typing

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/ast"
)

// Note:
// MinCaml doesn't have scope. When you want to add scope, you need to implement nested table and each scope
// should have its table. When some variable reference appears in code, try to find its definition from
// the nested tables.

type Env struct {
	// Types for declarations. This is referred by type variables to resolve
	// type variables' actual types
	Table map[string]ast.Type
	// External variable names which are referred but not defined.
	// External variables are exposed as external symbols in other object files.
	Externals map[string]ast.Type
}

func NewEnv() *Env {
	return &Env{
		map[string]ast.Type{},
		map[string]ast.Type{},
	}
}

func (env *Env) ApplyTypeAnalysis(root ast.Expr) error {
	t, err := env.infer(root)
	if err != nil {
		return err
	}

	if err := Unify(ast.UnitTypeVal, t); err != nil {
		return errors.Wrap(err, "Type of root expression of program must be unit\n")
	}

	// While dereferencing type variables in table, we can detect type variables
	// which does not have exact type and raise an error for that.
	// External variables must be well-typed also.
	if err := DerefTypeVars(root); err != nil {
		return err
	}

	return nil
}

func (env *Env) Dump() {
	fmt.Println("Variables:")
	for n, t := range env.Table {
		fmt.Printf("  %s: %s\n", n, t.String())
	}
	fmt.Println()
	env.DumpExternals()
}

func (env *Env) DumpExternals() {
	fmt.Println("External Variables:")
	for n, t := range env.Externals {
		fmt.Printf("  %s: %s\n", n, t.String())
	}
}
