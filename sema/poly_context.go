package sema

import (
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/types"
)

type polyContext struct {
	env      *types.Env
	assign   map[types.VarID]types.Type
	inferred InferredTypes
}

func newPolyContext(env *types.Env, inferred InferredTypes) *polyContext {
	return &polyContext{env, map[types.VarID]types.Type{}, inferred}
}

func (poly *polyContext) typeOf(node ast.Expr) types.Type {
	t, ok := poly.inferred[node]
	if !ok {
		panic("FATAL: Type was not inferred for node '" + node.Name() + "' at " + node.Pos().String())
	}
	return t
}

// TODO
