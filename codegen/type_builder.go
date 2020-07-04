package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/types"
	"llvm.org/llvm/bindings/go/llvm"
)

type typeBuilder struct {
	context   llvm.Context
	env       *types.Env
	unitT     llvm.Type
	intT      llvm.Type
	floatT    llvm.Type
	boolT     llvm.Type
	stringT   llvm.Type
	voidT     llvm.Type
	voidPtrT  llvm.Type
	sizeT     llvm.Type
	optIntT   llvm.Type
	optBoolT  llvm.Type
	optFloatT llvm.Type
	captures  map[string]llvm.Type
}

func newTypeBuilder(ctx llvm.Context, intPtrTy llvm.Type, env *types.Env) *typeBuilder {
	integer := ctx.Int64Type()
	unit := ctx.StructCreateNamed("gocaml.unit")
	unit.StructSetBody([]llvm.Type{}, false /*packed*/)
	str := ctx.StructCreateNamed("gocaml.string")
	str.StructSetBody([]llvm.Type{
		llvm.PointerType(ctx.Int8Type(), 0 /*address space*/),
		integer,
	}, false /*packed*/)

	return &typeBuilder{
		ctx,
		env,
		unit,
		integer,
		ctx.DoubleType(),
		ctx.Int1Type(),
		str,
		ctx.VoidType(),
		llvm.PointerType(ctx.Int8Type(), 0 /*address space*/),
		intPtrTy,
		ctx.IntType(65), // 64bit int + 1bit flag
		ctx.IntType(2),  // 1bit int + 1bit flag
		ctx.IntType(65), // 64bit float + 1bit flag
		map[string]llvm.Type{},
	}
}

func (b *typeBuilder) buildClosureCaptures(name string, closure []string) llvm.Type {
	if cached, ok := b.captures[name]; ok {
		return cached
	}

	fields := make([]llvm.Type, 0, len(closure))
	for _, capture := range closure {
		t, ok := b.env.DeclTable[capture]
		if !ok {
			panic(fmt.Sprintf("Type of capture '%s' not found!", capture))
		}
		fields = append(fields, b.fromMIR(t))
	}

	captures := b.context.StructType(fields, false /*packed*/)
	b.captures[name] = captures
	return captures
}

func (b *typeBuilder) buildExternalFun(from *types.Fun) llvm.Type {
	ret := b.fromMIR(from.Ret)
	if ret == b.unitT {
		// If return type of external function is unit, use void instead of unit
		// because external function (usually written in C) does not have unit type.
		// Instead, it has void for the purpose.
		ret = b.voidT
	}
	params := make([]llvm.Type, 0, len(from.Params))
	for _, p := range from.Params {
		params = append(params, b.fromMIR(p))
	}
	return llvm.FunctionType(ret, params, false /*varargs*/)
}

func (b *typeBuilder) buildExternalClosure(from *types.Fun) llvm.Type {
	ret := b.fromMIR(from.Ret)
	params := make([]llvm.Type, 0, len(from.Params)+1)
	params = append(params, b.voidPtrT)
	for _, p := range from.Params {
		params = append(params, b.fromMIR(p))
	}
	return llvm.FunctionType(ret, params, false /*varargs*/)
}

// Note:
// Function type is basically closure type. Only when applying function directly
// or applying external function, callee should not be closure.
func (b *typeBuilder) buildFun(from *types.Fun, known bool) llvm.Type {
	ret := b.fromMIR(from.Ret)
	l := len(from.Params)
	if !known {
		l++
	}
	params := make([]llvm.Type, 0, l)
	if !known {
		params = append(params, b.voidPtrT) // Closure
	}
	for _, p := range from.Params {
		params = append(params, b.fromMIR(p))
	}
	return llvm.FunctionType(ret, params, false /*varargs*/)
}

// Creates closure type for the specified function ignoring capture fields
// This function is used for retrieving function pointer from i8* closure value.
func (b *typeBuilder) buildClosure(ty *types.Fun) llvm.Type {
	funPtr := llvm.PointerType(b.buildFun(ty, false), 0 /*address space*/)
	return b.context.StructType([]llvm.Type{funPtr, b.voidPtrT}, false /*packed*/)
}

func (b *typeBuilder) buildOption(ty *types.Option) llvm.Type {
	switch elem := ty.Elem.(type) {
	case *types.Int:
		return b.optIntT
	case *types.Bool:
		return b.optBoolT
	case *types.Float:
		return b.optFloatT
	case *types.String, *types.Fun, *types.Tuple, *types.Array:
		// Represents 'None' value with NULL pointer
		return b.fromMIR(elem)
	case *types.Option:
		elems := []llvm.Type{
			b.boolT,
			b.buildOption(elem),
		}
		return b.context.StructType(elems, false /*packed*/)
	case *types.Unit:
		elems := []llvm.Type{
			b.boolT,
			b.unitT,
		}
		return b.context.StructType(elems, false /*packed*/)
	default:
		panic("unreachable: " + ty.String())
	}
}

func (b *typeBuilder) fromMIR(from types.Type) llvm.Type {
	switch ty := from.(type) {
	case *types.Unit:
		return b.unitT
	case *types.Bool:
		return b.boolT
	case *types.Int:
		return b.intT
	case *types.Float:
		return b.floatT
	case *types.String:
		return b.stringT
	case *types.Fun:
		// Function type which occurs in normal expression's type is always closure because
		// function type variable is always closure. Normal function pointer never occurs in value context.
		// It must be a callee of direct function call (optimized by known function optimization).
		return b.buildClosure(ty)
	case *types.Tuple:
		elems := make([]llvm.Type, 0, len(ty.Elems))
		for _, e := range ty.Elems {
			elems = append(elems, b.fromMIR(e))
		}
		return llvm.PointerType(b.context.StructType(elems, false /*packed*/), 0 /*address space*/)
	case *types.Array:
		return b.context.StructType([]llvm.Type{
			llvm.PointerType(b.fromMIR(ty.Elem), 0 /*address space*/),
			// size
			b.intT,
		}, false /*packed*/)
	case *types.Option:
		return b.buildOption(ty)
	case *types.Var:
		panic("unreachable")
	default:
		panic("unreachable")
	}
}
