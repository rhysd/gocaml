package codegen

import (
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"llvm.org/llvm/bindings/go/llvm"
)

type Emitter struct {
	GCIL   *gcil.Program
	Env    *typing.Env
	Source *token.Source
	Module llvm.Module
}

func NewEmitter(prog *gcil.Program, env *typing.Env, src *token.Source) (*Emitter, error) {
	builder := newModuleBuilder(env, src.Name)
	if err := builder.build(prog); err != nil {
		return nil, err
	}

	return &Emitter{
		prog,
		env,
		src,
		builder.module,
	}, nil
}
