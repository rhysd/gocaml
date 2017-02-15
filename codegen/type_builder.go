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
	unit.StructSetBody([]llvm.Type{} /*packed*/, false)
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

func (b *typeBuilder) buildCapturesStruct(name string, closure []string) llvm.Type {
	if cached, ok := b.captures[name]; ok {
		return cached
	}
	fields := make([]llvm.Type, 0, len(closure))
	for _, capture := range closure {
		t, ok := b.env.Table[capture]
		if !ok {
			panic(fmt.Sprintf("Type of capture '%s' not found!", capture))
		}
		fields = append(fields, b.build(t))
	}
	ty := b.context.StructCreateNamed(fmt.Sprintf("%s.capture", name))
	ty.StructSetBody(fields /*packed*/, false)
	b.captures[name] = ty
	return ty
}

func (b *typeBuilder) buildExternalFun(from *typing.Fun) llvm.Type {
	ret := b.build(from.Ret)
	if ret == b.unitT {
		ret = b.voidT
	}
	params := make([]llvm.Type, 0, len(from.Params))
	for _, p := range from.Params {
		params = append(params, b.build(p))
	}
	return llvm.FunctionType(ret, params, false /*varargs*/)
}

// Note:
// Function type is basically closure type. Only when applying function directly
// or applying external function, callee should not be closure.
func (b *typeBuilder) buildFun(from *typing.Fun, known bool) llvm.Type {
	ret := b.build(from.Ret)
	l := len(from.Params)
	if !known {
		l++
	}
	params := make([]llvm.Type, 0, l)
	if !known {
		params = append(params, b.voidPtrT) // Closure
	}
	for _, p := range from.Params {
		params = append(params, b.build(p))
	}
	return llvm.FunctionType(ret, params /*varargs*/, false)
}

func (b *typeBuilder) build(from typing.Type) llvm.Type {
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
		return b.buildFun(ty, true)
	case *typing.Tuple:
		elems := make([]llvm.Type, 0, len(ty.Elems))
		for _, e := range ty.Elems {
			elems = append(elems, b.build(e))
		}
		return b.context.StructType(elems /*packed*/, false)
	case *typing.Array:
		array := b.context.StructCreateNamed("array")
		array.StructSetBody([]llvm.Type{
			llvm.PointerType(b.build(ty.Elem) /*address space*/, 0),
			// size
			b.intT,
		}, /*packed*/ false)
		return array
	case *typing.Var:
		if ty.Ref == nil {
			panic("Unknown type variable while building LLVM type value!")
		}
		return b.build(ty.Ref)
	}
	panic("unreachable")
}
