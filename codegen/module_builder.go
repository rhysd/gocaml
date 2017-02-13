package codegen

import (
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/typing"
	"llvm.org/llvm/bindings/go/llvm"
	"strings"
)

type moduleBuilder struct {
	module     llvm.Module
	env        *typing.Env
	dataLayout string
	context    llvm.Context
	// TODO: Data layout, LLVM context, IRBuilder, ...
}

func lookupTargetInfo(triple Triple) (llvm.Target, error) {
	arch := triple.Arch()
	for target := llvm.FirstTarget(); target.C != nil; target = target.NextTarget() {
		if strings.HasPrefix(target.Name(), arch) {
			return target, nil
		}
	}
	return llvm.Target{}, errors.Errorf("No target information found for triple '%s'", triple)
}

func newModuleBuilder(env *typing.Env, name string, opts EmitOptions) (*moduleBuilder, error) {

	triple := NewTriple(opts.Triple)

	optLevel := llvm.CodeGenLevelDefault
	switch opts.Optimization {
	case OptimizeAggressive:
		optLevel = llvm.CodeGenLevelAggressive
	case OptimizeNone:
		optLevel = llvm.CodeGenLevelNone
	}

	target, err := lookupTargetInfo(triple)
	if err != nil {
		return nil, err
	}

	machine := target.CreateTargetMachine(
		string(triple),
		"", // CPU
		"", // Features
		optLevel,
		llvm.RelocDefault,     // static or dynamic-no-pic or default
		llvm.CodeModelDefault, // small, medium, large, kernel, JIT-default, default
	)

	targetData := machine.CreateTargetData()
	dataLayout := targetData.String()
	targetData.Dispose()
	machine.Dispose()

	module := llvm.NewModule(name)
	module.SetTarget(string(triple))
	module.SetDataLayout(dataLayout)

	return &moduleBuilder{module, env, dataLayout, llvm.GlobalContext()}, nil
}

func (builder *moduleBuilder) build(prog *gcil.Program) error {
	return nil // TODO
}
