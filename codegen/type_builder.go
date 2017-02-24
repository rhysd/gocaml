package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/typing"
	"llvm.org/llvm/bindings/go/llvm"
)

type typeBuilder struct {
	context  llvm.Context
	env      *typing.Env
	unitT    llvm.Type
	intT     llvm.Type
	floatT   llvm.Type
	boolT    llvm.Type
	voidT    llvm.Type
	voidPtrT llvm.Type
	captures map[string]llvm.Type
}

func newTypeBuilder(ctx llvm.Context, env *typing.Env) *typeBuilder {
	unit := ctx.StructCreateNamed("gocaml.unit")
	unit.StructSetBody([]llvm.Type{}, false /*packed*/)
	return &typeBuilder{
		ctx,
		env,
		unit,
		ctx.Int64Type(),
		ctx.DoubleType(),
		ctx.Int1Type(),
		ctx.VoidType(),
		llvm.PointerType(ctx.Int8Type(), 0),
		map[string]llvm.Type{},
	}
}

func (b *typeBuilder) buildClosureCaptures(name string, closure []string) llvm.Type {
	if cached, ok := b.captures[name]; ok {
		return cached
	}

	fields := make([]llvm.Type, 0, len(closure))
	for _, capture := range closure {
		t, ok := b.env.Table[capture]
		if !ok {
			panic(fmt.Sprintf("Type of capture '%s' not found!", capture))
		}
		fields = append(fields, b.convertGCIL(t))
	}

	captures := b.context.StructType(fields, false /*packed*/)
	b.captures[name] = captures
	return captures
}

func (b *typeBuilder) buildExternalFun(from *typing.Fun) llvm.Type {
	ret := b.convertGCIL(from.Ret)
	if ret == b.unitT {
		ret = b.voidT
	}
	params := make([]llvm.Type, 0, len(from.Params))
	for _, p := range from.Params {
		params = append(params, b.convertGCIL(p))
	}
	return llvm.FunctionType(ret, params, false /*varargs*/)
}

// Note:
// Function type is basically closure type. Only when applying function directly
// or applying external function, callee should not be closure.
func (b *typeBuilder) buildFun(from *typing.Fun, known bool) llvm.Type {
	ret := b.convertGCIL(from.Ret)
	l := len(from.Params)
	if !known {
		l++
	}
	params := make([]llvm.Type, 0, l)
	if !known {
		params = append(params, b.voidPtrT) // Closure
	}
	for _, p := range from.Params {
		params = append(params, b.convertGCIL(p))
	}
	return llvm.FunctionType(ret, params /*varargs*/, false)
}

// Creates closure type for the specified function ignoring capture fields
// This function is used for retrieving function pointer from i8* closure value.
func (b *typeBuilder) buildClosure(ty *typing.Fun) llvm.Type {
	funPtr := llvm.PointerType(b.buildFun(ty, false), 0 /*address space*/)
	return b.context.StructType([]llvm.Type{funPtr, b.voidPtrT}, false /*packed*/)
}

func (b *typeBuilder) convertGCIL(from typing.Type) llvm.Type {
	switch ty := from.(type) {
	case *typing.Unit:
		return b.unitT
	case *typing.Bool:
		return b.boolT
	case *typing.Int:
		return b.intT
	case *typing.Float:
		return b.floatT
	case *typing.Fun:
		// Function type which occurs in normal expression's type is always closure because
		// function type variable is always closure. Normal function pointer never occurs in value context.
		// It must be a callee of direct function call (optimized by known function optimization).
		return b.buildClosure(ty)
	case *typing.Tuple:
		elems := make([]llvm.Type, 0, len(ty.Elems))
		for _, e := range ty.Elems {
			elems = append(elems, b.convertGCIL(e))
		}
		return llvm.PointerType(b.context.StructType(elems, false /*packed*/), 0 /*address space*/)
	case *typing.Array:
		arrStruct := b.context.StructType([]llvm.Type{
			llvm.PointerType(b.convertGCIL(ty.Elem), 0 /*address space*/),
			// size
			b.intT,
		}, false /*packed*/)
		return llvm.PointerType(arrStruct, 0 /*address space*/)
	case *typing.Var:
		panic("unreachable")
	default:
		panic("unreachable")
	}
}
