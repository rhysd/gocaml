package codegen

import (
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"llvm.org/llvm/bindings/go/llvm"
)

type OptLevel int

const (
	OptimizeAggressive OptLevel = iota
	OptimizeDefault
	OptimizeNone
)

type EmitOptions struct {
	Optimization OptLevel
	Triple       string
}

type Emitter struct {
	GCIL     *gcil.Program
	Env      *typing.Env
	Source   *token.Source
	Module   llvm.Module
	Disposed bool
}

func (emitter *Emitter) Dispose() {
	emitter.Module.Dispose()
	emitter.Disposed = true
}

func NewEmitter(prog *gcil.Program, env *typing.Env, src *token.Source, opts EmitOptions) (*Emitter, error) {
	builder, err := newModuleBuilder(env, src.Name, opts)
	if err != nil {
		return nil, err
	}
	if err = builder.build(prog); err != nil {
		return nil, err
	}

	return &Emitter{
		prog,
		env,
		src,
		builder.module,
		false,
	}, nil
}
