// Package codegen provides code generation of GoCaml language.
//
// MIR compilation unit is compiled to an LLVM IR, an assembly, an object then finally linked to an executable.
// You can add many optimizations and debug information (DWARF).
package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/typing"
	"github.com/rhysd/locerr"
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
	// Equivalent to -O0
	OptimizeNone OptLevel = iota
	// Equivalent to -O1
	OptimizeLess
	// Equivalent to -O2
	OptimizeDefault
	// Equivalent to -O3
	OptimizeAggressive
)

// Options to customize emitter behavior
type EmitOptions struct {
	// Determines how many optimizations are added
	Optimization OptLevel
	// Target triple "{arch}-{vendor}-{sys}". Empty string means a default target on your machine.
	// https://clang.llvm.org/docs/CrossCompilation.html#target-triple
	Triple string
	// Additional linker flags used at linking generated object files
	LinkerFlags string
	// Generate debug information or not. If true, debug information will be added and you can debug the generated executable with debugger like an LLDB.
	DebugInfo bool
}

// Emitter object to emit LLVM IR, object file, assembly or executable.
type Emitter struct {
	EmitOptions
	MIR     *mir.Program
	Env      *typing.Env
	Source   *locerr.Source
	Module   llvm.Module
	Machine  llvm.TargetMachine
	Disposed bool
}

// Finalization for internal module and target machine.
// You need to call this with defer statement.
func (emitter *Emitter) Dispose() {
	if emitter.Disposed {
		return
	}
	emitter.Module.Dispose()
	emitter.Machine.Dispose()
	emitter.Disposed = true
}

// Passes optimizations on generated LLVM IR module following specified optimization level.
func (emitter *Emitter) RunOptimizationPasses() {
	if emitter.Optimization == OptimizeNone {
		return
	}
	level := int(emitter.Optimization)

	builder := llvm.NewPassManagerBuilder()
	defer builder.Dispose()
	builder.SetOptLevel(level)

	// Threshold magic numbers came from computeThresholdFromOptLevels() in llvm/lib/Analysis/InlineCost.cpp
	threshold := uint(225) // O2
	if emitter.Optimization == OptimizeAggressive {
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

// Returns LLVM IR as string.
func (emitter *Emitter) EmitLLVMIR() string {
	return emitter.Module.String()
}

// Returns assembly code as string.
func (emitter *Emitter) EmitAsm() (string, error) {
	buf, err := emitter.Machine.EmitToMemoryBuffer(emitter.Module, llvm.AssemblyFile)
	if err != nil {
		return "", err
	}
	asm := string(buf.Bytes())
	buf.Dispose()
	return asm, nil
}

// Returns object file contents as byte sequence.
func (emitter *Emitter) EmitObject() ([]byte, error) {
	buf, err := emitter.Machine.EmitToMemoryBuffer(emitter.Module, llvm.ObjectFile)
	if err != nil {
		return nil, err
	}
	obj := buf.Bytes()
	buf.Dispose()
	return obj, nil
}

// Create executable file with specified name. This is the final result of compilation!
func (emitter *Emitter) EmitExecutable(executable string) (err error) {
	objfile := fmt.Sprintf("%s.tmp.o", executable)
	obj, err := emitter.EmitObject()
	if err != nil {
		return
	}
	if err = ioutil.WriteFile(objfile, obj, 0666); err != nil {
		return
	}
	defer os.Remove(objfile)
	linker := newDefaultLinker(emitter.LinkerFlags)
	err = linker.link(executable, []string{objfile})
	// Linker link runtime and make an executable
	return
}

// Creates new emitter object.
func NewEmitter(prog *mir.Program, env *typing.Env, src *locerr.Source, opts EmitOptions) (*Emitter, error) {
	builder, err := newModuleBuilder(env, src, opts)
	if err != nil {
		return nil, err
	}

	if err = builder.build(prog); err != nil {
		return nil, err
	}
	defer builder.dispose()

	return &Emitter{
		opts,
		prog,
		env,
		src,
		builder.module,
		builder.machine,
		false,
	}, nil
}
