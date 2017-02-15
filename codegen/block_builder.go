package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/typing"
	"llvm.org/llvm/bindings/go/llvm"
)

type blockBuilder struct {
	*moduleBuilder
	registers map[string]llvm.Value
}

func newBlockBuilder(b *moduleBuilder) *blockBuilder {
	return &blockBuilder{b, map[string]llvm.Value{}}
}

func (b *blockBuilder) resolve(ident string) llvm.Value {
	if glob, ok := b.globalTable[ident]; ok {
		return b.builder.CreateLoad(glob, ident)
	}
	if reg, ok := b.registers[ident]; ok {
		return reg
	}
	panic("No value was found for identifier: " + ident)
}

func (b *blockBuilder) typeOf(ident string) typing.Type {
	if t, ok := b.env.Table[ident]; ok {
		for {
			v, ok := t.(*typing.Var)
			if !ok {
				return t
			}
			if v.Ref == nil {
				panic("Empty type variable while searching variable: " + ident)
			}
			t = v.Ref
		}
	}
	if t, ok := b.env.Externals[ident]; ok {
		for {
			v, ok := t.(*typing.Var)
			if !ok {
				return t
			}
			if v.Ref == nil {
				panic("Empty type variable while searching external variable: " + ident)
			}
			t = v.Ref
		}
	}
	panic("Type was not found for ident: " + ident)
}

func (b *blockBuilder) buildVal(ident string, val gcil.Val) llvm.Value {
	switch val := val.(type) {
	case *gcil.Unit:
		return llvm.ConstStruct([]llvm.Value{}, false /*packed*/)
	case *gcil.Bool:
		c := uint64(1)
		if !val.Const {
			c = 0
		}
		return llvm.ConstInt(b.typeBuilder.boolT, c, false /*sign extend*/)
	case *gcil.Int:
		return llvm.ConstInt(b.typeBuilder.intT, uint64(val.Const), true /*sign extend*/)
	case *gcil.Float:
		return llvm.ConstFloat(b.typeBuilder.floatT, val.Const)
	case *gcil.Unary:
		child := b.resolve(val.Child)
		switch val.Op {
		case gcil.NEG:
			return b.builder.CreateNeg(child, "neg")
		case gcil.FNEG:
			return b.builder.CreateFNeg(child, "fneg")
		case gcil.NOT:
			return b.builder.CreateNot(child, "not")
		default:
			panic("unreachable")
		}
	case *gcil.Binary:
		lhs := b.resolve(val.Lhs)
		rhs := b.resolve(val.Rhs)
		switch val.Op {
		case gcil.ADD:
			return b.builder.CreateAdd(lhs, rhs, "add")
		case gcil.SUB:
			return b.builder.CreateSub(lhs, rhs, "sub")
		case gcil.FADD:
			return b.builder.CreateFAdd(lhs, rhs, "fadd")
		case gcil.FSUB:
			return b.builder.CreateFSub(lhs, rhs, "fsub")
		case gcil.FMUL:
			return b.builder.CreateFMul(lhs, rhs, "fmul")
		case gcil.FDIV:
			return b.builder.CreateFDiv(lhs, rhs, "fdiv")
		case gcil.LESS:
			lty := b.typeOf(val.Lhs)
			switch lty.(type) {
			case *typing.Int:
				return b.builder.CreateICmp(llvm.IntSLT /*Signed Less Than*/, lhs, rhs, "less")
			case *typing.Float:
				return b.builder.CreateFCmp(llvm.FloatOLT /*Ordered and Less Than*/, lhs, rhs, "less")
			default:
				panic("Invalid type for '<' operator: " + lty.String())
			}
		case gcil.EQ:
			lty := b.typeOf(val.Lhs)
			switch lty.(type) {
			case *typing.Unit:
				// `() = ()` is always true.
				return llvm.ConstInt(b.typeBuilder.boolT, 1, false /*sign extend*/)
			case *typing.Bool, *typing.Int:
				return b.builder.CreateICmp(llvm.IntEQ, lhs, rhs, "eql")
			case *typing.Float:
				return b.builder.CreateFCmp(llvm.FloatOEQ, lhs, rhs, "eql")
			case *typing.Tuple:
				panic("not implemented yet: comparing tuples")
			case *typing.Array:
				panic("not implemented yet: comparing arrays")
			default:
				panic("unreachable")
			}
		default:
			panic("unreachable")
		}
	case *gcil.Ref:
		reg, ok := b.registers[val.Ident]
		if !ok {
			panic("Value not found for ref: " + val.Ident)
		}
		return reg
	case *gcil.If:
		parent := b.builder.GetInsertBlock().Parent()
		thenBlock := llvm.AddBasicBlock(parent, "if.then")
		elseBlock := llvm.AddBasicBlock(parent, "if.else")
		endBlock := llvm.AddBasicBlock(parent, "if.end")

		ty := b.typeBuilder.build(b.typeOf(ident))
		cond := b.resolve(val.Cond)
		b.builder.CreateCondBr(cond, thenBlock, elseBlock)

		b.builder.SetInsertPointAtEnd(thenBlock)
		thenVal := b.build(val.Then)
		b.builder.CreateBr(endBlock)

		b.builder.SetInsertPointAtEnd(elseBlock)
		elseVal := b.build(val.Else)
		b.builder.CreateBr(endBlock)

		b.builder.SetInsertPointAtEnd(endBlock)
		phi := b.builder.CreatePHI(ty, "if.merge")
		phi.AddIncoming([]llvm.Value{thenVal, elseVal}, []llvm.BasicBlock{thenBlock, elseBlock})
		return phi
	case *gcil.Fun:
		panic("unreachable because IR was closure-transformed")
	case *gcil.App:
		argsLen := len(val.Args)
		if val.Kind == gcil.CLOSURE_CALL {
			argsLen++
		}
		argVals := make([]llvm.Value, 0, argsLen)

		if val.Kind == gcil.CLOSURE_CALL {
			// Add pointer to closure captures
			argVals = append(argVals, b.resolve(val.Callee))
		}
		for _, a := range val.Args {
			argVals = append(argVals, b.resolve(a))
		}

		table := b.funcTable
		if val.Kind == gcil.EXTERNAL_CALL {
			table = b.globalTable
		}
		funVal, ok := table[val.Callee]
		if !ok {
			panic("Value for function is not found in table: " + val.Callee)
		}

		// Note:
		// Call inst cannot have a name when the return type is void.
		return b.builder.CreateCall(funVal, argVals, "")
	case *gcil.Tuple:
		ty := b.typeBuilder.build(b.typeOf(ident))
		ptr := b.builder.CreateAlloca(ty, ident)
		for i, e := range val.Elems {
			v := b.resolve(e)
			p := b.builder.CreateStructGEP(ptr, i, fmt.Sprintf("%s.%d", ident, i))
			b.builder.CreateStore(v, p)
		}
		return ptr
	case *gcil.Array:
		t, ok := b.typeOf(ident).(*typing.Array)
		if !ok {
			panic("Type of array literal is not array")
		}

		ty := b.typeBuilder.build(t)
		elemTy := b.typeBuilder.build(t.Elem)
		ptr := b.builder.CreateAlloca(ty, ident)

		sizeVal := b.resolve(val.Size)

		// XXX:
		// Arrays are allocated on stack. So returning array value from function
		// now breaks the array value.
		arrVal := b.builder.CreateArrayAlloca(elemTy, sizeVal, "array.ptr")

		arrPtr := b.builder.CreateStructGEP(ptr, 0, "")
		b.builder.CreateStore(arrVal, arrPtr)

		// XXX:
		// Need to store rhs value deeply (consider tuple or array as element of array)
		// We need to implement array creation as a function in runtime.
		//
		// elemVal := b.resolve(val.Elem)

		sizePtr := b.builder.CreateStructGEP(ptr, 1, "")
		b.builder.CreateStore(sizeVal, sizePtr)

		return ptr
	case *gcil.TplLoad:
		from := b.resolve(val.From)
		p := b.builder.CreateStructGEP(from, val.Index, "")
		return b.builder.CreateLoad(p, "tplload")
	case *gcil.ArrLoad:
		fromVal := b.resolve(val.From)
		idxVal := b.resolve(val.Index)
		arrPtr := b.builder.CreateStructGEP(fromVal, 0, "")
		elemPtr := b.builder.CreateInBoundsGEP(arrPtr, []llvm.Value{idxVal}, "")
		return b.builder.CreateLoad(elemPtr, "arrload")
	case *gcil.ArrStore:
		// XXX:
		// Need to store rhs value deeply (consider tuple or array as element of array)
		toVal := b.resolve(val.To)
		idxVal := b.resolve(val.Index)
		rhsVal := b.resolve(val.Rhs)
		arrPtr := b.builder.CreateStructGEP(toVal, 0, "")
		elemPtr := b.builder.CreateInBoundsGEP(arrPtr, []llvm.Value{idxVal}, "")
		return b.builder.CreateStore(rhsVal, elemPtr)
	case *gcil.XRef:
		x, ok := b.globalTable[val.Ident]
		if !ok {
			panic("Value for external value not found: " + val.Ident)
		}
		return b.builder.CreateLoad(x, val.Ident)
	case *gcil.MakeCls:
		freevarVals := make([]llvm.Value, 0, len(val.Vars))
		for _, v := range val.Vars {
			freevarVals = append(freevarVals, b.resolve(v))
		}

		closure, ok := b.closures[val.Fun]
		if !ok {
			panic("closure for function not found: " + val.Fun)
		}
		closureTy := b.typeBuilder.buildCapturesStruct(val.Fun, closure)
		alloca := b.builder.CreateAlloca(closureTy, fmt.Sprintf("closure.%s", val.Fun))
		ptr := b.builder.CreateBitCast(alloca, b.typeBuilder.voidPtrT, "closure.ptr")
		return ptr
	case *gcil.NOP:
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (b *blockBuilder) buildInsn(insn *gcil.Insn) llvm.Value {
	v := b.buildVal(insn.Ident, insn.Val)
	b.registers[insn.Ident] = v
	return v
}

func (b *blockBuilder) build(block *gcil.Block) llvm.Value {
	i := block.Top.Next
	for {
		v := b.buildInsn(i)
		i = i.Next
		if i.Next == nil {
			return v
		}
	}
}
