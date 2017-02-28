package codegen

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/typing"
	"llvm.org/llvm/bindings/go/llvm"
)

type moduleBuilder struct {
	module      llvm.Module
	env         *typing.Env
	machine     llvm.TargetMachine
	targetData  llvm.TargetData
	context     llvm.Context
	builder     llvm.Builder
	typeBuilder *typeBuilder
	attributes  map[string]llvm.Attribute
	globalTable map[string]llvm.Value
	funcTable   map[string]llvm.Value
	closures    gcil.Closures
}

func createAttributeTable(ctx llvm.Context) map[string]llvm.Attribute {
	attrs := map[string]llvm.Attribute{}

	// Enum attributes
	for _, attr := range []string{
		"nounwind",
		"noreturn",
		"inlinehint",
		"ssp",
		"uwtable",
	} {
		kind := llvm.AttributeKindID(attr)
		attrs[attr] = ctx.CreateEnumAttribute(kind, 0)
	}

	// String attributes
	for _, attr := range []struct {
		kind  string
		value string
	}{
		{"disable-tail-calls", "false"},
	} {
		attrs[attr.kind] = ctx.CreateStringAttribute(attr.kind, attr.value)
	}

	return attrs
}

func newModuleBuilder(env *typing.Env, name string, opts EmitOptions) (*moduleBuilder, error) {
	triple := opts.Triple
	if triple == "" {
		triple = llvm.DefaultTargetTriple()
	}

	optLevel := llvm.CodeGenLevelDefault
	switch opts.Optimization {
	case OptimizeNone:
		optLevel = llvm.CodeGenLevelNone
	case OptimizeLess:
		optLevel = llvm.CodeGenLevelLess
	case OptimizeAggressive:
		optLevel = llvm.CodeGenLevelAggressive
	}

	target, err := llvm.GetTargetFromTriple(triple)
	if err != nil {
		return nil, err
	}

	machine := target.CreateTargetMachine(
		triple,
		"", // CPU
		"", // Features
		optLevel,
		llvm.RelocDefault,     // static or dynamic-no-pic or default
		llvm.CodeModelDefault, // small, medium, large, kernel, JIT-default, default
	)

	targetData := machine.CreateTargetData()
	dataLayout := targetData.String()

	// XXX: Should make a new instance
	ctx := llvm.GlobalContext()

	module := ctx.NewModule(name)
	module.SetTarget(string(triple))
	module.SetDataLayout(dataLayout)

	// Note:
	// We create registers table for each blocks because closure transform
	// breaks alpha-transformed identifiers. But all identifiers are identical
	// in block.
	return &moduleBuilder{
		module,
		env,
		machine,
		targetData,
		ctx,
		ctx.NewBuilder(),
		newTypeBuilder(ctx, targetData.IntPtrType(), env),
		createAttributeTable(ctx),
		nil,
		nil,
		nil,
	}, nil
}

func (b *moduleBuilder) Dispose() {
	b.targetData.Dispose()
}

func (b *moduleBuilder) declareExternalDecl(name string, from typing.Type) llvm.Value {
	switch ty := from.(type) {
	case *typing.Var:
		panic("unreachable") // because type variables are dereferenced at type analysis
	case *typing.Fun:
		t := b.typeBuilder.buildExternalFun(ty)
		v := llvm.AddFunction(b.module, name, t)
		v.SetLinkage(llvm.ExternalLinkage)
		v.AddFunctionAttr(b.attributes["disable-tail-calls"])
		return v
	default:
		t := b.typeBuilder.convertGCIL(from)
		v := llvm.AddGlobal(b.module, t, name)
		v.SetLinkage(llvm.ExternalLinkage)
		return v
	}
}

func (b *moduleBuilder) declareFun(insn gcil.FunInsn) llvm.Value {
	name := insn.Name
	_, isClosure := b.closures[name]
	found, ok := b.env.Table[name]
	if !ok {
		panic(fmt.Sprintf("Type not found for function '%s'", name))
	}

	ty, ok := found.(*typing.Fun)
	if !ok {
		panic(fmt.Sprintf("Type of function '%s' is not a function type: %s", name, found.String()))
	}

	t := b.typeBuilder.buildFun(ty, !isClosure)
	v := llvm.AddFunction(b.module, name, t)

	index := 0
	if isClosure {
		v.Param(index).SetName("closure")
		index++
	}

	for _, p := range insn.Val.Params {
		v.Param(index).SetName(p)
		index++
	}

	v.AddFunctionAttr(b.attributes["inlinehint"])
	v.AddFunctionAttr(b.attributes["nounwind"])
	v.AddFunctionAttr(b.attributes["ssp"])
	v.AddFunctionAttr(b.attributes["uwtable"])
	v.AddFunctionAttr(b.attributes["disable-tail-calls"])

	return v
}

func (b *moduleBuilder) buildFunBody(insn gcil.FunInsn) llvm.Value {
	name := insn.Name
	fun := insn.Val
	llvmFun, ok := b.funcTable[name]
	if !ok {
		panic("Unknown function on building IR: " + name)
	}
	body := b.context.AddBasicBlock(llvmFun, "entry")
	b.builder.SetInsertPointAtEnd(body)

	blockBuilder := newBlockBuilder(b)

	// Extract captured variables
	closure, isClosure := b.closures[name]

	for i, p := range fun.Params {
		if isClosure {
			// First parameter is a pointer to captures
			i++
		}
		blockBuilder.registers[p] = llvmFun.Param(i)
	}

	// Expose captures of closure
	if isClosure {
		if len(closure) > 0 {
			capturesTy := llvm.PointerType(b.typeBuilder.buildClosureCaptures(name, closure), 0 /*address space*/)
			closureVal := b.builder.CreateBitCast(llvmFun.Param(0), capturesTy, fmt.Sprintf("%s.capture", name))
			for i, n := range closure {
				ptr := b.builder.CreateStructGEP(closureVal, i, "")
				exposed := b.builder.CreateLoad(ptr, fmt.Sprintf("%s.capture.%s", name, n))
				blockBuilder.registers[n] = exposed
			}
		}
		if fun.IsRecursive {
			// When the closure itself is used in its body, it needs to prepare the closure object
			// for the recursive use.
			//
			// Note:
			// We cannot use a closure object {funptr, capturesptr} instead of capturesptr for the first parameter of
			// closure function because it will produce infinite recursive type in parameter of function type.
			// (please see the comment in 69970b6b16d2e6765d63a16647ccea2b379433c8)
			itselfTy := b.context.StructType([]llvm.Type{llvmFun.Type(), b.typeBuilder.voidPtrT}, false /*packed*/)
			itselfVal := llvm.Undef(itselfTy)
			itselfVal = b.builder.CreateInsertValue(itselfVal, llvmFun, 0, "")
			itselfVal = b.builder.CreateInsertValue(itselfVal, llvmFun.Param(0), 1, "")
			blockBuilder.registers[name] = itselfVal
		}
	}

	lastVal := blockBuilder.buildBlock(fun.Body)
	return b.builder.CreateRet(lastVal)
}

func (b *moduleBuilder) buildMain(entry *gcil.Block) {
	int32T := b.context.Int32Type()
	t := llvm.FunctionType(int32T, []llvm.Type{}, false /*varargs*/)
	funVal := llvm.AddFunction(b.module, "main", t)
	funVal.AddFunctionAttr(b.attributes["ssp"])
	funVal.AddFunctionAttr(b.attributes["uwtable"])
	funVal.AddFunctionAttr(b.attributes["disable-tail-calls"])

	body := b.context.AddBasicBlock(funVal, "entry")
	b.builder.SetInsertPointAtEnd(body)

	initGcFun, ok := b.globalTable["GC_init"]
	if !ok {
		panic("'GC_init' not found. Function prototypes for libgc were not emitted")
	}
	b.builder.CreateCall(initGcFun, []llvm.Value{}, "")

	builder := newBlockBuilder(b)
	builder.buildBlock(entry)

	b.builder.CreateRet(llvm.ConstInt(int32T, 0, true))
}

func (b *moduleBuilder) buildLibgcFuncDecls() {
	for _, fun := range []struct {
		name string
		ret  llvm.Type
		args []llvm.Type
	}{
		{
			"GC_malloc",
			b.typeBuilder.voidPtrT,
			[]llvm.Type{b.typeBuilder.sizeT},
		},
		{
			"GC_init",
			b.typeBuilder.voidT,
			[]llvm.Type{},
		},
	} {
		t := llvm.FunctionType(fun.ret, fun.args, false /*vaargs*/)
		v := llvm.AddFunction(b.module, fun.name, t)
		v.SetLinkage(llvm.ExternalLinkage)
		v.AddFunctionAttr(b.attributes["nounwind"])
		b.globalTable[fun.name] = v
	}
}

func (b *moduleBuilder) build(prog *gcil.Program) error {
	// Note:
	// Currently global variables are external symbols only.
	b.globalTable = make(map[string]llvm.Value, len(b.env.Externals)+2 /* 2 = libgc functions */)

	b.buildLibgcFuncDecls()
	for name, ty := range b.env.Externals {
		b.globalTable[name] = b.declareExternalDecl(name, ty)
	}

	b.closures = prog.Closures
	b.funcTable = make(map[string]llvm.Value, len(prog.Toplevel))
	for name, fun := range prog.Toplevel {
		b.funcTable[name] = b.declareFun(fun)
	}

	for _, fun := range prog.Toplevel {
		b.buildFunBody(fun)
	}

	b.buildMain(prog.Entry)

	if err := llvm.VerifyModule(b.module, llvm.ReturnStatusAction); err != nil {
		return errors.Wrapf(err, "Error while emitting IR:\n\n%s\n", b.module.String())
	}

	return nil
}
