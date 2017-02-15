package codegen

import (
	"path/filepath"
	"strings"

	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"llvm.org/llvm/bindings/go/llvm"
)

func init() {
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
}

type OptLevel int

const (
	OptimizeNone OptLevel = iota
	OptimizeLess
	OptimizeDefault
	OptimizeAggressive
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
	if emitter.Disposed {
		return
	}
	emitter.Module.Dispose()
	emitter.Disposed = true
}

func (emitter *Emitter) baseName() string {
	if !emitter.Source.Exists {
		return "out"
	}
	b := filepath.Base(emitter.Source.Name)
	return strings.TrimSuffix(b, filepath.Ext(b))
}

func (emitter *Emitter) EmitLLVMIR() string {
	return emitter.Module.String()
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
