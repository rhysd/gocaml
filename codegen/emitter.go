package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"io/ioutil"
	"llvm.org/llvm/bindings/go/llvm"
	"os"
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
	LinkerFlags  string
}

type Emitter struct {
	GCIL     *gcil.Program
	Env      *typing.Env
	Source   *token.Source
	Module   llvm.Module
	Machine  llvm.TargetMachine
	Options  EmitOptions
	Disposed bool
}

func (emitter *Emitter) Dispose() {
	if emitter.Disposed {
		return
	}
	emitter.Module.Dispose()
	emitter.Machine.Dispose()
	emitter.Disposed = true
}

func (emitter *Emitter) RunOptimizationPasses() {
	if emitter.Options.Optimization == OptimizeNone {
		return
	}
	level := int(emitter.Options.Optimization)

	builder := llvm.NewPassManagerBuilder()
	defer builder.Dispose()
	builder.SetOptLevel(level)

	// Threshold magic numbers came from computeThresholdFromOptLevels() in llvm/lib/Analysis/InlineCost.cpp
	threshold := uint(225) // O2
	if emitter.Options.Optimization == OptimizeAggressive {
		// -O1 is the same inline level as -O2
		threshold = 275
	}
	builder.UseInlinerWithThreshold(threshold)

	funcPasses := llvm.NewFunctionPassManagerForModule(emitter.Module)
	defer funcPasses.Dispose()
	builder.PopulateFunc(funcPasses)
	for fun := emitter.Module.FirstFunction(); fun.C != nil; fun = llvm.NextFunction(fun) {
		if fun.IsDeclaration() {
			continue
		}
		funcPasses.InitializeFunc()
		funcPasses.RunFunc(fun)
		funcPasses.FinalizeFunc()
	}

	modPasses := llvm.NewPassManager()
	defer modPasses.Dispose()
	builder.Populate(modPasses)
	modPasses.Run(emitter.Module)
}

func (emitter *Emitter) EmitLLVMIR() string {
	return emitter.Module.String()
}

func (emitter *Emitter) EmitAsm() (string, error) {
	buf, err := emitter.Machine.EmitToMemoryBuffer(emitter.Module, llvm.AssemblyFile)
	if err != nil {
		return "", err
	}
	asm := string(buf.Bytes())
	buf.Dispose()
	return asm, nil
}

func (emitter *Emitter) EmitObject() ([]byte, error) {
	buf, err := emitter.Machine.EmitToMemoryBuffer(emitter.Module, llvm.ObjectFile)
	if err != nil {
		return nil, err
	}
	obj := buf.Bytes()
	buf.Dispose()
	return obj, nil
}

func (emitter *Emitter) EmitExecutable(executable string) (err error) {
	objfile := fmt.Sprintf("_%s.o", executable)
	obj, err := emitter.EmitObject()
	if err != nil {
		return
	}
	if err = ioutil.WriteFile(objfile, obj, 0666); err != nil {
		return
	}
	defer os.Remove(objfile)
	linker := newDefaultLinker(emitter.Options.LinkerFlags)
	err = linker.link(executable, []string{objfile})
	// Linker link runtime and make an executable
	return
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
		builder.machine,
		opts,
		false,
	}, nil
}
