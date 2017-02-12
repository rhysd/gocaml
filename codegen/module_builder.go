package codegen

import (
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/typing"
	"llvm.org/llvm/bindings/go/llvm"
)

type moduleBuilder struct {
	module llvm.Module
	env    *typing.Env
	// TODO: Data layout, LLVM context, IRBuilder, ...
}

func newModuleBuilder(env *typing.Env, name string) *moduleBuilder {
	return &moduleBuilder{llvm.NewModule(name), env}
}

func (builder moduleBuilder) build(prog *gcil.Program) error {
	return nil // TODO
}
