package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
	"llvm.org/llvm/bindings/go/llvm"
)

func getOpCmpPredicate(op mir.OperatorKind) (llvm.IntPredicate, llvm.FloatPredicate, string) {
	switch op {
	case mir.LT:
		// SLT = Signed Less Than, OLT = Ordered and Less Than
		return llvm.IntSLT, llvm.FloatOLT, "less"
	case mir.LTE:
		return llvm.IntSLE, llvm.FloatOLE, "lesseq"
	case mir.GT:
		return llvm.IntSGT, llvm.FloatOGT, "greater"
	case mir.GTE:
		return llvm.IntSGE, llvm.FloatOGE, "greatereq"
	case mir.EQ:
		return llvm.IntEQ, llvm.FloatOEQ, "eql"
	case mir.NEQ:
		return llvm.IntNE, llvm.FloatONE, "neq"
	default:
		panic("unreachable")
	}
}

type blockBuilder struct {
	*moduleBuilder
	registers   map[string]llvm.Value
	unitVal     llvm.Value
	allocaBlock llvm.BasicBlock
}

func newBlockBuilder(b *moduleBuilder, allocaBlock llvm.BasicBlock) *blockBuilder {
	unit := llvm.Undef(b.typeBuilder.unitT)
	return &blockBuilder{b, map[string]llvm.Value{}, unit, allocaBlock}
}

func (b *blockBuilder) resolve(ident string) llvm.Value {
	// Note:
	// No need to check b.globalTable because there is no global variable in GoCaml.
	// Functions and external symbols are treated as global variable. But they are directly referred
	// in builder. So we don't need to check global variables generally here.
	if reg, ok := b.registers[ident]; ok {
		return reg
	}
	panic("No value was found for identifier: " + ident)
}

func (b *blockBuilder) typeOf(ident string) types.Type {
	if t, ok := b.env.DeclTable[ident]; ok {
		return t
	}
	// Note:
	// b.env.Table() now contains types for all identifiers of external symbols.
	// So we don't need to check b.env.Externals to know type of identifier.
	panic("Type was not found for ident: " + ident)
}

func (b *blockBuilder) buildMallocRaw(ty llvm.Type, sizeVal llvm.Value, name string) llvm.Value {
	mallocVal, ok := b.globalTable["GC_malloc"]
	if !ok {
		panic("'GC_malloc' not found. Function protoypes for libgc were not emitted")
	}
	allocated := b.builder.CreateCall(mallocVal, []llvm.Value{sizeVal}, "")
	ptrTy := llvm.PointerType(ty, 0 /*address space*/)
	return b.builder.CreateBitCast(allocated, ptrTy, name)
}

func (b *blockBuilder) buildMalloc(ty llvm.Type, name string) llvm.Value {
	size := b.targetData.TypeAllocSize(ty)
	sizeVal := llvm.ConstInt(b.typeBuilder.sizeT, size, false /*sign extend*/)
	return b.buildMallocRaw(ty, sizeVal, name)
}

func (b *blockBuilder) buildArrayMalloc(ty llvm.Type, numElems llvm.Value, name string) llvm.Value {
	size := b.targetData.TypeAllocSize(ty)
	tySizeVal := llvm.ConstInt(b.typeBuilder.sizeT, size, false /*sign extend*/)
	sizeVal := b.builder.CreateMul(tySizeVal, b.builder.CreateTrunc(numElems, b.typeBuilder.sizeT, ""), "")
	return b.buildMallocRaw(ty, sizeVal, name)
}

func (b *blockBuilder) buildAlloca(t llvm.Type, name string) llvm.Value {
	saved := b.builder.GetInsertBlock()
	b.builder.SetInsertPointAtEnd(b.allocaBlock)
	alloca := b.builder.CreateAlloca(t, name)

	// XXX:
	// This function assumes that the previous insertion point was at the end of the block.
	// If it pointed the middle of block, this function would fail to restore insertion point.
	// This is because there is no LLVM-C API corresponding to IRBuilderBase::GetInsertPoint.
	b.builder.SetInsertPointAtEnd(saved)

	return alloca
}

func (b *blockBuilder) buildEq(ty types.Type, bin *mir.Binary, lhs, rhs llvm.Value) llvm.Value {
	icmp, fcmp, name := getOpCmpPredicate(bin.Op)

	switch ty := ty.(type) {
	case *types.Unit:
		// `() = ()` is always true and `() <> ()` is never true.
		i := uint64(1)
		if bin.Op == mir.NEQ {
			i = 0
		}
		return llvm.ConstInt(b.typeBuilder.boolT, i, false /*sign extend*/)
	case *types.Bool, *types.Int:
		return b.builder.CreateICmp(icmp, lhs, rhs, name)
	case *types.Float:
		return b.builder.CreateFCmp(fcmp, lhs, rhs, name)
	case *types.String:
		eqlFun, ok := b.globalTable["__str_equal"]
		if !ok {
			panic("__str_equal() not found")
		}
		cmp := b.builder.CreateCall(eqlFun, []llvm.Value{lhs, rhs}, "")
		i := uint64(1)
		if bin.Op == mir.NEQ {
			i = 0
		}
		return b.builder.CreateICmp(llvm.IntEQ, cmp, llvm.ConstInt(b.typeBuilder.boolT, i, false /*signed*/), "eql.str")
	case *types.Tuple:
		cmp := llvm.Value{}
		for i, elemTy := range ty.Elems {
			l := b.builder.CreateLoad(b.builder.CreateStructGEP(lhs, i, "tpl.left"), "")
			r := b.builder.CreateLoad(b.builder.CreateStructGEP(rhs, i, "tpl.right"), "")
			elemCmp := b.buildEq(elemTy, bin, l, r)
			if cmp.C == nil {
				cmp = elemCmp
			} else {
				cmp = b.builder.CreateAnd(cmp, elemCmp, "")
			}
		}
		cmp.SetName(name + ".tpl")
		return cmp
	case *types.Fun:
		// Note:
		// The function instance must be a closure because all functions which is used
		// as variable are treated as closure in closure-transform.
		lfun := b.builder.CreateExtractValue(lhs, 0, "")
		rfun := b.builder.CreateExtractValue(rhs, 0, "")
		return b.builder.CreateICmp(icmp, lfun, rfun, name+".fun")
	case *types.Option:
		return b.buildEqOption(ty, bin, lhs, rhs)
	case *types.Array:
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (b *blockBuilder) buildLess(val *mir.Binary, lhs, rhs llvm.Value) llvm.Value {
	lty := b.typeOf(val.LHS)
	ipred, fpred, name := getOpCmpPredicate(val.Op)
	switch lty.(type) {
	case *types.Int:
		return b.builder.CreateICmp(ipred, lhs, rhs, name)
	case *types.Float:
		return b.builder.CreateFCmp(fpred, lhs, rhs, name)
	default:
		panic(fmt.Sprintf("Invalid type for '%s' operator: %s", name, lty.String()))
	}
}

func (b *blockBuilder) buildEqOption(ty *types.Option, bin *mir.Binary, lhs, rhs llvm.Value) llvm.Value {
	tyVal := b.typeBuilder.buildOption(ty)
	lhsIsSome := b.buildIsSome(lhs, tyVal, ty)
	rhsIsSome := b.buildIsSome(rhs, tyVal, ty)

	parent := b.builder.GetInsertBlock().Parent()
	bothSomeBlk := llvm.AddBasicBlock(parent, "eq.opt.both")
	elseBlk := llvm.AddBasicBlock(parent, "eq.opt.else")
	endBlk := llvm.AddBasicBlock(parent, "eq.opt.end")

	cond := b.builder.CreateAnd(lhsIsSome, rhsIsSome, "")
	b.builder.CreateCondBr(cond, bothSomeBlk, elseBlk)

	// When both values are Some(v), compare contained values
	b.builder.SetInsertPointAtEnd(bothSomeBlk)
	lhsDeref := b.buildDerefSome(lhs, ty)
	rhsDeref := b.buildDerefSome(rhs, ty)
	bothEqVal := b.buildEq(ty.Elem, bin, lhsDeref, rhsDeref)
	b.builder.CreateBr(endBlk)
	bothLastBlk := b.builder.GetInsertBlock()

	// Otherwise, see issome(lhs) || issome(rhs)
	elseBlk.MoveAfter(bothLastBlk)
	b.builder.SetInsertPointAtEnd(elseBlk)
	// Either lhs or rhs is Some(v).
	elseEqVal := b.builder.CreateOr(lhsIsSome, rhsIsSome, "")
	if bin.Op == mir.EQ {
		elseEqVal = b.builder.CreateNot(elseEqVal, "")
	}
	b.builder.CreateBr(endBlk)

	b.builder.SetInsertPointAtEnd(endBlk)
	phi := b.builder.CreatePHI(b.typeBuilder.boolT, "eq.opt.merge")
	phi.AddIncoming([]llvm.Value{bothEqVal, elseEqVal}, []llvm.BasicBlock{bothLastBlk, elseBlk})
	return phi
}

func (b *blockBuilder) buildIsSome(optVal llvm.Value, tyVal llvm.Type, ty *types.Option) llvm.Value {
	switch ty.Elem.(type) {
	case *types.Int, *types.Bool, *types.Float:
		one := llvm.ConstInt(tyVal, 1, false /*signed*/)
		// Extract flag value
		flag := b.builder.CreateAnd(optVal, one, "")
		// flag == 1 means that it contains a value
		return b.builder.CreateICmp(llvm.IntEQ, flag, one, "issome")
	case *types.String, *types.Fun, *types.Array:
		ptr := b.builder.CreateExtractValue(optVal, 0, "")
		return b.builder.CreateNot(b.builder.CreateIsNull(ptr, ""), "issome")
	case *types.Tuple:
		return b.builder.CreateNot(b.builder.CreateIsNull(optVal, ""), "issome")
	case *types.Option, *types.Unit:
		flag := b.builder.CreateExtractValue(optVal, 0, "")
		return b.builder.CreateICmp(
			llvm.IntEQ,
			flag,
			llvm.ConstInt(b.typeBuilder.boolT, 1, false /*signed*/),
			"issome",
		)
	default:
		panic("unreachable")
	}
}

func (b *blockBuilder) buildDerefSome(optVal llvm.Value, ty *types.Option) llvm.Value {
	switch ty.Elem.(type) {
	case *types.Int:
		// shift 1 bit to squash a flag
		one := llvm.ConstInt(llvm.IntType(65), 1, false /*signed*/)
		v := b.builder.CreateLShr(optVal, one, "")
		// Truncate to the same size bits
		return b.builder.CreateTrunc(v, b.typeBuilder.intT, "derefsome")
	case *types.Float:
		// shift 1 bit to squash a flag
		one := llvm.ConstInt(llvm.IntType(65), 1, false /*signed*/)
		v := b.builder.CreateLShr(optVal, one, "")
		// Truncate to the same size bits
		v = b.builder.CreateTrunc(v, llvm.IntType(64), "")
		return b.builder.CreateBitCast(v, b.typeBuilder.fromMIR(ty.Elem), "derefsome")
	case *types.Bool:
		// shift 1 bit to squash a flag
		one := llvm.ConstInt(llvm.IntType(2), 1, false /*signed*/)
		v := b.builder.CreateLShr(optVal, one, "")
		// Truncate to the same size bits
		return b.builder.CreateTrunc(v, b.typeBuilder.boolT, "derefsome")
	case *types.String, *types.Fun, *types.Array, *types.Tuple:
		return optVal
	case *types.Option, *types.Unit:
		return b.builder.CreateExtractValue(optVal, 1, "derefsome")
	default:
		panic("unreachable")
	}
}

func (b *blockBuilder) buildVal(ident string, val mir.Val) llvm.Value {
	switch val := val.(type) {
	case *mir.Unit:
		return b.unitVal
	case *mir.Bool:
		c := uint64(1)
		if !val.Const {
			c = 0
		}
		return llvm.ConstInt(b.typeBuilder.boolT, c, false /*sign extend*/)
	case *mir.Int:
		return llvm.ConstInt(b.typeBuilder.intT, uint64(val.Const), true /*sign extend*/)
	case *mir.Float:
		return llvm.ConstFloat(b.typeBuilder.floatT, val.Const)
	case *mir.String:
		strVal := b.buildAlloca(b.typeBuilder.stringT, "")

		charsVal := b.builder.CreateGlobalStringPtr(val.Const, "")
		charsPtr := b.builder.CreateStructGEP(strVal, 0, "")
		b.builder.CreateStore(charsVal, charsPtr)

		sizeVal := llvm.ConstInt(b.typeBuilder.intT, uint64(len(val.Const)), true /*signed*/)
		sizePtr := b.builder.CreateStructGEP(strVal, 1, "str.size")
		b.builder.CreateStore(sizeVal, sizePtr)

		return b.builder.CreateLoad(strVal, "str")
	case *mir.Unary:
		child := b.resolve(val.Child)
		switch val.Op {
		case mir.NEG:
			return b.builder.CreateNeg(child, "neg")
		case mir.FNEG:
			return b.builder.CreateFNeg(child, "fneg")
		case mir.NOT:
			return b.builder.CreateNot(child, "not")
		default:
			panic("unreachable")
		}
	case *mir.Binary:
		lhs := b.resolve(val.LHS)
		rhs := b.resolve(val.RHS)
		switch val.Op {
		case mir.ADD:
			return b.builder.CreateAdd(lhs, rhs, "add")
		case mir.SUB:
			return b.builder.CreateSub(lhs, rhs, "sub")
		case mir.MUL:
			return b.builder.CreateMul(lhs, rhs, "mul")
		case mir.DIV:
			return b.builder.CreateSDiv(lhs, rhs, "div")
		case mir.MOD:
			return b.builder.CreateSRem(lhs, rhs, "mod")
		case mir.FADD:
			return b.builder.CreateFAdd(lhs, rhs, "fadd")
		case mir.FSUB:
			return b.builder.CreateFSub(lhs, rhs, "fsub")
		case mir.FMUL:
			return b.builder.CreateFMul(lhs, rhs, "fmul")
		case mir.FDIV:
			return b.builder.CreateFDiv(lhs, rhs, "fdiv")
		case mir.LT, mir.LTE, mir.GT, mir.GTE:
			return b.buildLess(val, lhs, rhs)
		case mir.EQ, mir.NEQ:
			return b.buildEq(b.typeOf(val.LHS), val, lhs, rhs)
		case mir.AND:
			return b.builder.CreateAnd(lhs, rhs, "andl")
		case mir.OR:
			return b.builder.CreateOr(lhs, rhs, "orl")
		default:
			panic("unreachable")
		}
	case *mir.Ref:
		reg, ok := b.registers[val.Ident]
		if !ok {
			panic("Value not found for ref: " + val.Ident)
		}
		return reg
	case *mir.If:
		parent := b.builder.GetInsertBlock().Parent()
		thenBlock := llvm.AddBasicBlock(parent, "if.then")
		elseBlock := llvm.AddBasicBlock(parent, "if.else")
		endBlock := llvm.AddBasicBlock(parent, "if.end")

		ty := b.typeBuilder.fromMIR(b.typeOf(ident))
		cond := b.resolve(val.Cond)
		b.builder.CreateCondBr(cond, thenBlock, elseBlock)

		b.builder.SetInsertPointAtEnd(thenBlock)
		thenVal := b.buildBlock(val.Then)
		b.builder.CreateBr(endBlock)
		thenLastBlock := b.builder.GetInsertBlock()

		elseBlock.MoveAfter(thenLastBlock)
		b.builder.SetInsertPointAtEnd(elseBlock)
		elseVal := b.buildBlock(val.Else)
		b.builder.CreateBr(endBlock)
		elseLastBlock := b.builder.GetInsertBlock()

		endBlock.MoveAfter(elseLastBlock)
		b.builder.SetInsertPointAtEnd(endBlock)
		phi := b.builder.CreatePHI(ty, "if.merge")
		phi.AddIncoming([]llvm.Value{thenVal, elseVal}, []llvm.BasicBlock{thenLastBlock, elseLastBlock})
		return phi
	case *mir.Fun:
		panic("unreachable because IR was closure-transformed")
	case *mir.App:
		argsLen := len(val.Args)
		if val.Kind == mir.CLOSURE_CALL {
			argsLen++
		}
		argVals := make([]llvm.Value, 0, argsLen)

		table := b.funcTable
		callee := val.Callee
		if val.Kind == mir.EXTERNAL_CALL {
			table = b.globalTable
			callee = b.env.Externals[val.Callee].CName
		}

		// Find function pointer for invoking a function directly
		funVal, funFound := table[callee]
		if !funFound && val.Kind != mir.CLOSURE_CALL {
			panic("Value for function is not found in table: " + callee)
		}

		if val.Kind == mir.CLOSURE_CALL {
			closureVal := b.resolve(val.Callee)

			// Extract function pointer from closure instance if callee does not indicates well-known function
			if !funFound {
				funVal = b.builder.CreateExtractValue(closureVal, 0, "funptr")
			}

			// Extract pointer to captures object
			capturesPtr := b.builder.CreateExtractValue(closureVal, 1, "capturesptr")
			argVals = append(argVals, capturesPtr)
		}

		for _, a := range val.Args {
			argVals = append(argVals, b.resolve(a))
		}

		// Note:
		// Call inst cannot have a name when the return type is void.
		ret := b.builder.CreateCall(funVal, argVals, "")
		if ret.Type().TypeKind() == llvm.VoidTypeKind {
			// When returned value is void
			ret = b.unitVal
		}
		return ret
	case *mir.Tuple:
		// Note:
		// Type of tuple is a pointer to struct. To obtain the value for tuple, we need underlying
		// struct type because 'alloca' instruction returns the pointer to allocated memory.
		ptrTy := b.typeBuilder.fromMIR(b.typeOf(ident))
		allocTy := ptrTy.ElementType()

		ptr := b.buildMalloc(allocTy, ident)
		for i, e := range val.Elems {
			v := b.resolve(e)
			p := b.builder.CreateStructGEP(ptr, i, fmt.Sprintf("%s.%d", ident, i))
			b.builder.CreateStore(v, p)
		}
		return ptr
	case *mir.Array:
		t, ok := b.typeOf(ident).(*types.Array)
		if !ok {
			panic("Type of array instruction is not array")
		}

		// Copy second argument to all elements of allocated array
		// Initialize array object {ptr, size}
		elemTy := b.typeBuilder.fromMIR(t.Elem)
		arr := llvm.Undef(b.typeBuilder.fromMIR(t))

		sizeVal := b.resolve(val.Size)
		arrVal := b.buildArrayMalloc(elemTy, sizeVal, "array.ptr")
		arr = b.builder.CreateInsertValue(arr, arrVal, 0, "")

		// Prepare 2nd argument value and iteration variable for the loop
		elemVal := b.resolve(val.Elem)
		iterPtr := b.buildAlloca(b.typeBuilder.intT, "arr.init.iter")
		b.builder.CreateStore(llvm.ConstInt(b.typeBuilder.intT, 0, false), iterPtr)

		// Start of the initialization loop
		parent := b.builder.GetInsertBlock().Parent()
		condBlock := llvm.AddBasicBlock(parent, "arr.init.cond")
		loopBlock := llvm.AddBasicBlock(parent, "arr.init.setelem")
		endBlock := llvm.AddBasicBlock(parent, "arr.init.end")
		b.builder.CreateBr(condBlock)
		b.builder.SetInsertPointAtEnd(condBlock)

		iterVal := b.builder.CreateLoad(iterPtr, "")
		compVal := b.builder.CreateICmp(llvm.IntEQ, iterVal, sizeVal, "")
		b.builder.CreateCondBr(compVal, endBlock, loopBlock)

		// Copy 2nd argument to each element
		b.builder.SetInsertPointAtEnd(loopBlock)
		elemPtr := b.builder.CreateInBoundsGEP(arrVal, []llvm.Value{iterVal}, "")
		b.builder.CreateStore(elemVal, elemPtr)
		iterVal = b.builder.CreateAdd(iterVal, llvm.ConstInt(b.typeBuilder.intT, 1, false), "arr.init.inc")
		b.builder.CreateStore(iterVal, iterPtr)
		b.builder.CreateBr(condBlock)

		// No need to use endBlock.MoveAfter() because no block was inserted
		// between loopBlock and endBlock
		b.builder.SetInsertPointAtEnd(endBlock)

		// Set size value
		arr = b.builder.CreateInsertValue(arr, sizeVal, 1, "")
		return arr
	case *mir.ArrLit:
		t, ok := b.typeOf(ident).(*types.Array)
		if !ok {
			panic("Type of arrlit instruction is not array")
		}

		arr := llvm.Undef(b.typeBuilder.fromMIR(t))
		sizeVal := llvm.ConstInt(b.typeBuilder.intT, uint64(len(val.Elems)), false /*signed*/)
		arr = b.builder.CreateInsertValue(arr, sizeVal, 1, "")

		if len(val.Elems) == 0 {
			return arr
		}

		elemTy := b.typeBuilder.fromMIR(t.Elem)
		arrPtr := b.buildArrayMalloc(elemTy, sizeVal, "array.ptr")
		arr = b.builder.CreateInsertValue(arr, arrPtr, 0, "")

		for i, elem := range val.Elems {
			indices := []llvm.Value{llvm.ConstInt(b.typeBuilder.intT, uint64(i), false /*sized*/)}
			elemVal := b.resolve(elem)
			elemPtr := b.builder.CreateInBoundsGEP(arrPtr, indices, fmt.Sprintf("array.elem.%d", i))
			b.builder.CreateStore(elemVal, elemPtr)
		}

		return arr
	case *mir.TplLoad:
		from := b.resolve(val.From)
		p := b.builder.CreateStructGEP(from, val.Index, "")
		return b.builder.CreateLoad(p, "tplload")
	case *mir.ArrLoad:
		fromVal := b.resolve(val.From)
		idxVal := b.resolve(val.Index)
		arrPtr := b.builder.CreateExtractValue(fromVal, 0, "")
		elemPtr := b.builder.CreateInBoundsGEP(arrPtr, []llvm.Value{idxVal}, "")
		return b.builder.CreateLoad(elemPtr, "arrload")
	case *mir.ArrStore:
		toVal := b.resolve(val.To)
		idxVal := b.resolve(val.Index)
		rhsVal := b.resolve(val.RHS)
		arrPtr := b.builder.CreateExtractValue(toVal, 0, "")
		elemPtr := b.builder.CreateInBoundsGEP(arrPtr, []llvm.Value{idxVal}, "")
		b.builder.CreateStore(rhsVal, elemPtr)
		return b.unitVal
	case *mir.ArrLen:
		fromVal := b.resolve(val.Array)
		return b.builder.CreateExtractValue(fromVal, 1, "arrsize")
	case *mir.XRef:
		ext, ok := b.env.Externals[val.Ident]
		if !ok {
			panic("Type for external value not found: " + val.Ident)
		}

		funTy, ok := ext.Type.(*types.Fun)
		if !ok {
			x, ok := b.globalTable[ext.CName]
			if !ok {
				panic("Value for external value not found: " + ext.CName)
			}
			return b.builder.CreateLoad(x, val.Ident)
		}

		// When external function is used as variable, it must be wrapped as closure
		// instead of global value itself.
		funVal := b.buildExternalClosureWrapper(val.Ident, funTy, ext.CName)
		clsTy := b.context.StructType([]llvm.Type{funVal.Type(), b.typeBuilder.voidPtrT}, false /*packed*/)
		alloc := b.buildAlloca(clsTy, "")
		funPtr := b.builder.CreateStructGEP(alloc, 0, "")
		b.builder.CreateStore(funVal, funPtr)
		return b.builder.CreateLoad(alloc, val.Ident+".cls")
	case *mir.MakeCls:
		closure, ok := b.closures[val.Fun]
		if !ok {
			panic("Closure for function not found: " + val.Fun)
		}

		funcT, ok := b.env.DeclTable[val.Fun].(*types.Fun)
		if !ok {
			panic(fmt.Sprintf("Type of function '%s' not found!", val.Fun))
		}
		funPtrTy := llvm.PointerType(b.typeBuilder.buildFun(funcT, false), 0 /*address space*/)

		closureTy := b.context.StructCreateNamed(fmt.Sprintf("%s.clsobj", val.Fun))
		capturesTy := b.typeBuilder.buildClosureCaptures(val.Fun, closure)
		closureTy.StructSetBody([]llvm.Type{funPtrTy, llvm.PointerType(capturesTy, 0 /*address space*/)}, false /*packed*/)

		closureVal := b.buildAlloca(closureTy, "")

		// Set function pointer to first field of closure
		funPtr, ok := b.funcTable[val.Fun]
		if !ok {
			panic("Value for function not found: " + val.Fun)
		}
		b.builder.CreateStore(funPtr, b.builder.CreateStructGEP(closureVal, 0, ""))

		capturesVal := b.buildMalloc(capturesTy, fmt.Sprintf("captures.%s", val.Fun))
		for i, v := range val.Vars {
			ptr := b.builder.CreateStructGEP(capturesVal, i, "")
			freevar := b.resolve(v)
			b.builder.CreateStore(freevar, ptr)
		}
		b.builder.CreateStore(capturesVal, b.builder.CreateStructGEP(closureVal, 1, ""))

		castedTy := llvm.PointerType(
			b.context.StructType([]llvm.Type{funPtrTy, b.typeBuilder.voidPtrT}, false /*packed*/),
			0, /*address space*/
		)
		castedVal := b.builder.CreateBitCast(closureVal, castedTy, "")

		return b.builder.CreateLoad(castedVal, fmt.Sprintf("closure.%s", val.Fun))
	case *mir.Some:
		elemVal := b.resolve(val.Elem)
		ty, ok := b.typeOf(ident).(*types.Option)
		if !ok {
			panic("Type of Some is not an option type: " + b.typeOf(ident).String())
		}

		switch ty.Elem.(type) {
		case *types.Int, *types.Bool:
			tyVal := b.typeBuilder.buildOption(ty)
			// Extend 1 bit for flag
			extended := b.builder.CreateZExt(elemVal, tyVal, "")
			// Lowest bit is a flag. So shift left by 1 bit
			shifted := b.builder.CreateShl(extended, llvm.ConstInt(tyVal, 1, false /*signed*/), "")
			// Set flag to 1
			return b.builder.CreateOr(shifted, llvm.ConstInt(tyVal, 1, false /*signed*/), "")
		case *types.Float:
			// Similar to Int or Bool cases, but bitcast is required
			tyVal := b.typeBuilder.buildOption(ty)
			casted := b.builder.CreateBitCast(elemVal, llvm.Int64Type(), "")
			extended := b.builder.CreateZExt(casted, tyVal, "")
			shifted := b.builder.CreateShl(extended, llvm.ConstInt(tyVal, 1, false /*signed*/), "")
			return b.builder.CreateOr(shifted, llvm.ConstInt(tyVal, 1, false /*signed*/), "")
		case *types.String, *types.Fun, *types.Array, *types.Tuple:
			// They use NULL pointer for 'None' value. So nothing to do to make 'Some' value.
			return elemVal
		case *types.Option, *types.Unit:
			v := llvm.Undef(b.typeBuilder.buildOption(ty))
			v = b.builder.CreateInsertValue(v, llvm.ConstInt(b.typeBuilder.boolT, 1, false), 0, "some.flag")
			v = b.builder.CreateInsertValue(v, elemVal, 1, "some.elem")
			return v
		default:
			panic("unreachable")
		}
	case *mir.None:
		ty, ok := b.typeOf(ident).(*types.Option)
		if !ok {
			panic("Type of None is not an option type: " + b.typeOf(ident).String())
		}

		tyVal := b.typeBuilder.buildOption(ty)
		switch ty.Elem.(type) {
		case *types.Int, *types.Bool, *types.Float:
			return llvm.ConstInt(tyVal, 0, false)
		case *types.String, *types.Fun, *types.Array:
			v := llvm.Undef(tyVal)
			null := llvm.ConstPointerNull(tyVal.StructElementTypes()[0])
			v = b.builder.CreateInsertValue(v, null, 0, "none.flag")
			return v
		case *types.Tuple:
			return llvm.ConstPointerNull(tyVal)
		case *types.Option, *types.Unit:
			v := llvm.Undef(b.typeBuilder.buildOption(ty))
			v = b.builder.CreateInsertValue(v, llvm.ConstInt(b.typeBuilder.boolT, 0, false), 0, "none.flag")
			return v
		default:
			panic("unreachable")
		}
	case *mir.IsSome:
		optVal := b.resolve(val.OptVal)
		ty, ok := b.typeOf(val.OptVal).(*types.Option)
		if !ok {
			panic("Type of IsSome is not an option type: " + b.typeOf(val.OptVal).String())
		}
		return b.buildIsSome(optVal, b.typeBuilder.buildOption(ty), ty)
	case *mir.DerefSome:
		optVal := b.resolve(val.SomeVal)
		ty, ok := b.typeOf(val.SomeVal).(*types.Option)
		if !ok {
			panic("Type of DerefSome is not an option type: " + b.typeOf(val.SomeVal).String())
		}
		return b.buildDerefSome(optVal, ty)
	case *mir.NOP:
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (b *blockBuilder) buildInsn(insn *mir.Insn) llvm.Value {
	if b.debug != nil {
		b.debug.setLocation(b.builder, insn.Pos)
	}
	v := b.buildVal(insn.Ident, insn.Val)
	b.registers[insn.Ident] = v
	return v
}

func (b *blockBuilder) buildBlock(block *mir.Block) llvm.Value {
	i := block.Top.Next
	for {
		v := b.buildInsn(i)
		i = i.Next
		if i.Next == nil {
			return v
		}
	}
}
