package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
	"llvm.org/llvm/bindings/go/llvm"
)

type moduleBuilder struct {
	module      llvm.Module
	env         *types.Env
	machine     llvm.TargetMachine
	targetData  llvm.TargetData
	context     llvm.Context
	builder     llvm.Builder
	typeBuilder *typeBuilder
	debug       *debugInfoBuilder
	attributes  map[string]llvm.Attribute
	globalTable map[string]llvm.Value
	funcTable   map[string]llvm.Value
	closures    mir.Closures
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
		"alwaysinline",
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

func newModuleBuilder(env *types.Env, file *locerr.Source, opts EmitOptions) (*moduleBuilder, error) {
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

	module := ctx.NewModule(file.Path)
	module.SetTarget(triple)
	module.SetDataLayout(dataLayout)

	typeBuilder := newTypeBuilder(ctx, targetData.IntPtrType(), env)

	var debug *debugInfoBuilder
	if opts.DebugInfo {
		debug, err = newDebugInfoBuilder(module, file, typeBuilder, targetData, opts.Optimization != OptimizeNone)
		if err != nil {
			return nil, err
		}
	}

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
		typeBuilder,
		debug,
		createAttributeTable(ctx),
		nil,
		nil,
		nil,
	}, nil
}

func (b *moduleBuilder) dispose() {
	b.targetData.Dispose()
	if b.debug != nil {
		b.debug.dispose()
	}
}

// Wrap as a closure for the external symbol function.
// This is necessary when the external symbol function is used as a variable.
// In GoCaml, all function variable falls back into closure value.
// External symbol function should also be closure in the case.
func (b *moduleBuilder) buildExternalClosureWrapper(funName string, ty *types.Fun, cName string) llvm.Value {
	name := funName + "$closure"
	if f, ok := b.funcTable[name]; ok {
		return f
	}

	if b.debug != nil {
		b.debug.clearLocation(b.builder)
	}

	// Build declaration of closure wrapper
	tyVal := b.typeBuilder.buildExternalClosure(ty)
	val := llvm.AddFunction(b.module, name, tyVal)
	val.SetLinkage(llvm.PrivateLinkage)
	val.AddFunctionAttr(b.attributes["alwaysinline"])
	val.AddFunctionAttr(b.attributes["nounwind"])
	val.AddFunctionAttr(b.attributes["ssp"])
	val.AddFunctionAttr(b.attributes["uwtable"])
	val.AddFunctionAttr(b.attributes["disable-tail-calls"])
	b.funcTable[name] = val

	extFunVal, ok := b.globalTable[cName]
	if !ok {
		panic("No external symbol for closure wrapper not found: " + cName)
	}

	// Build definition of closure wrapper
	saved := b.builder.GetInsertBlock()
	body := b.context.AddBasicBlock(val, "entry")
	b.builder.SetInsertPointAtEnd(body)
	lenArgs := len(ty.Params)
	args := make([]llvm.Value, 0, lenArgs)
	for i := 0; i < lenArgs; i++ {
		args = append(args, val.Param(i+1))
	}
	ret := b.builder.CreateCall(extFunVal, args, "")
	if ty.Ret == types.UnitType {
		// When the external function returns void
		ret = llvm.ConstNamedStruct(b.typeBuilder.unitT, []llvm.Value{})
	}
	b.builder.CreateRet(ret)
	b.builder.SetInsertPointAtEnd(saved)

	return val
}

func (b *moduleBuilder) buildExternalDecl(ext *types.External) {
	switch ty := ext.Type.(type) {
	case *types.Var:
		panic("unreachable")
	case *types.Fun:
		// Make a declaration for the external symbol function
		tyVal := b.typeBuilder.buildExternalFun(ty)
		val := llvm.AddFunction(b.module, ext.CName, tyVal)
		val.SetLinkage(llvm.ExternalLinkage)
		val.AddFunctionAttr(b.attributes["disable-tail-calls"])
		b.globalTable[ext.CName] = val
	default:
		t := b.typeBuilder.fromMIR(ty)
		v := llvm.AddGlobal(b.module, t, ext.CName)
		v.SetLinkage(llvm.ExternalLinkage)
		b.globalTable[ext.CName] = v
	}
}

func (b *moduleBuilder) buildFuncDecl(insn mir.FunInsn) {
	name := insn.Name
	_, isClosure := b.closures[name]
	found, ok := b.env.DeclTable[name]
	if !ok {
		panic(fmt.Sprintf("Type not found for function '%s'", name))
	}

	ty, ok := found.(*types.Fun)
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

	// Currently GoCaml does not have modules. So all functions are private.
	v.SetLinkage(llvm.PrivateLinkage)

	v.AddFunctionAttr(b.attributes["inlinehint"])
	v.AddFunctionAttr(b.attributes["nounwind"])
	v.AddFunctionAttr(b.attributes["ssp"])
	v.AddFunctionAttr(b.attributes["uwtable"])
	v.AddFunctionAttr(b.attributes["disable-tail-calls"])

	b.funcTable[name] = v
}

func (b *moduleBuilder) buildFunBody(insn mir.FunInsn) {
	name := insn.Name
	fun := insn.Val
	funVal, ok := b.funcTable[name]
	if !ok {
		panic("Unknown function on building IR: " + name)
	}

	allocaBlock := b.context.AddBasicBlock(funVal, "entry")
	start := b.context.AddBasicBlock(funVal, "start")
	b.builder.SetInsertPointAtEnd(start)
	blockBuilder := newBlockBuilder(b, allocaBlock)

	// Extract captured variables
	closure, isClosure := b.closures[name]

	for i, p := range fun.Params {
		if isClosure {
			// First parameter is a pointer to captures
			i++
		}
		blockBuilder.registers[p] = funVal.Param(i)
	}

	if b.debug != nil {
		ty, ok := b.env.DeclTable[name].(*types.Fun)
		if !ok {
			panic("Type for function definition not found: " + name)
		}
		b.debug.setFuncInfo(funVal, ty, insn.Pos.Line, isClosure)
	}

	// Expose captures of closure
	if isClosure {
		if len(closure) > 0 {
			capturesTy := llvm.PointerType(b.typeBuilder.buildClosureCaptures(name, closure), 0 /*address space*/)
			closureVal := b.builder.CreateBitCast(funVal.Param(0), capturesTy, fmt.Sprintf("%s.capture", name))
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
			itselfTy := b.context.StructType([]llvm.Type{funVal.Type(), b.typeBuilder.voidPtrT}, false /*packed*/)
			itselfVal := llvm.Undef(itselfTy)
			itselfVal = b.builder.CreateInsertValue(itselfVal, funVal, 0, "")
			itselfVal = b.builder.CreateInsertValue(itselfVal, funVal.Param(0), 1, "")
			blockBuilder.registers[name] = itselfVal
		}
	}

	lastVal := blockBuilder.buildBlock(fun.Body)
	b.builder.CreateRet(lastVal)
	if b.debug != nil {
		b.debug.clearLocation(b.builder)
	}

	if allocaBlock.FirstInstruction().C == nil {
		// When no alloca instruction was used in the function body
		allocaBlock.EraseFromParent()
	} else {
		// Insert allocation block before starting to execute function body
		//
		// *Entry* -> [Alloca] -> [Start] -> ... -> *End*
		//
		b.builder.SetInsertPointAtEnd(allocaBlock)
		b.builder.CreateBr(start)
	}
}

func (b *moduleBuilder) buildMain(entry *mir.Block) {
	int32T := b.context.Int32Type()
	t := llvm.FunctionType(int32T, []llvm.Type{}, false /*varargs*/)
	funVal := llvm.AddFunction(b.module, "__gocaml_main", t)
	funVal.AddFunctionAttr(b.attributes["inlinehint"])
	funVal.AddFunctionAttr(b.attributes["nounwind"])
	funVal.AddFunctionAttr(b.attributes["ssp"])
	funVal.AddFunctionAttr(b.attributes["uwtable"])
	funVal.AddFunctionAttr(b.attributes["disable-tail-calls"])

	if b.debug != nil {
		pos := entry.Top.Next.Pos
		b.debug.setMainFuncInfo(funVal, pos.Line)
	}

	allocaBlock := b.context.AddBasicBlock(funVal, "entry")
	start := b.context.AddBasicBlock(funVal, "start")
	b.builder.SetInsertPointAtEnd(start)
	builder := newBlockBuilder(b, allocaBlock)
	builder.buildBlock(entry)

	b.builder.CreateRet(llvm.ConstInt(int32T, 0, true))
	if b.debug != nil {
		b.debug.clearLocation(b.builder)
	}

	if allocaBlock.FirstInstruction().C == nil {
		// When no alloca instruction was used in the function body
		allocaBlock.EraseFromParent()
	} else {
		// Insert allocation block before starting to execute function body
		//
		// *Entry* -> [Alloca] -> [Start] -> ... -> *End*
		//
		b.builder.SetInsertPointAtEnd(allocaBlock)
		b.builder.CreateBr(start)
	}
}

func (b *moduleBuilder) buildLibgcFuncDecls() {
	t := llvm.FunctionType(b.typeBuilder.voidPtrT, []llvm.Type{b.typeBuilder.sizeT}, false /*vaargs*/)
	v := llvm.AddFunction(b.module, "GC_malloc", t)
	v.SetLinkage(llvm.ExternalLinkage)
	v.AddFunctionAttr(b.attributes["nounwind"])
	b.globalTable["GC_malloc"] = v
}

func (b *moduleBuilder) build(prog *mir.Program) error {
	// Note:
	// Currently global variables are external symbols only.
	b.globalTable = make(map[string]llvm.Value, len(b.env.Externals)+1 /* 1 = libgc functions */)
	// Note:
	// Closures for external functions are also defined.
	b.funcTable = make(map[string]llvm.Value, len(prog.Toplevel)+len(b.env.Externals))

	b.buildLibgcFuncDecls()
	for _, ext := range b.env.Externals {
		b.buildExternalDecl(ext)
	}

	b.closures = prog.Closures
	for _, fun := range prog.Toplevel {
		b.buildFuncDecl(fun)
	}

	for _, fun := range prog.Toplevel {
		b.buildFunBody(fun)
	}

	b.buildMain(prog.Entry)
	if b.debug != nil {
		b.debug.finalize()
	}

	if err := llvm.VerifyModule(b.module, llvm.ReturnStatusAction); err != nil {
		return locerr.Notef(err, "Error while emitting IR:\n\n%s\n", b.module.String())
	}

	return nil
}
