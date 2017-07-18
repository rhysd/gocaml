package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
)

// Convert AST into MIR with K-Normalization

type emitter struct {
	count    uint
	env      *types.Env
	inferred InferredTypes
	insts    refInsts
}

func (e *emitter) genID() string {
	e.count++
	return fmt.Sprintf("$k%d", e.count)
}

func (e *emitter) typeOf(node ast.Expr) types.Type {
	t, ok := e.inferred[node]
	if !ok {
		panic("FATAL: Type was not inferred for node '" + node.Name() + "' at " + node.Pos().String())
	}
	return t
}

func (e *emitter) insn(val mir.Val, prev *mir.Insn, node ast.Expr) *mir.Insn {
	id := e.genID()
	e.env.DeclTable[id] = e.typeOf(node)
	return mir.Concat(mir.NewInsn(id, val, node.Pos()), prev)
}

func (e *emitter) emitBinaryInsn(op mir.OperatorKind, lhs, rhs, node ast.Expr) *mir.Insn {
	l := e.emitInsn(lhs)
	r := e.emitInsn(rhs)
	r.Append(l)
	return e.insn(&mir.Binary{op, l.Ident, r.Ident}, r, node)
}

func (e *emitter) emitLetInsn(node *ast.Let) *mir.Insn {
	// TODO: Do not emit insn if it's unused generic decl

	// Note:
	// Introduce shortcut about symbol to reduce number of instruction nodes.
	//
	// Before:
	//   $k1 = some_insn
	//   $sym$t1 = ref $k1
	//
	// After:
	//   $sym$t1 = some_insn
	//
	// Here `$k1` is `bound.Ident`.
	bound := e.emitInsn(node.Bound)
	t, _ := e.env.DeclTable[bound.Ident]
	delete(e.env.DeclTable, bound.Ident)
	if inst, ok := e.env.RefInsts[bound.Ident]; ok {
		delete(e.env.RefInsts, bound.Ident)
		e.env.RefInsts[node.Symbol.Name] = inst
	}

	bound.Ident = node.Symbol.Name
	e.env.DeclTable[bound.Ident] = t

	body := e.emitInsn(node.Body)
	body.Append(bound)
	return body
}

func (e *emitter) emitFunInsn(node *ast.LetRec) *mir.Insn {
	// TODO: Do not emit insn if it's unused generic function

	name := node.Func.Symbol.Name
	ty, ok := e.env.DeclTable[name]
	if !ok {
		panic("FATAL: Unknown function: " + name)
	}

	params := make([]string, 0, len(node.Func.Params))
	for _, s := range node.Func.Params {
		params = append(params, s.Ident.Name)
	}

	blk := e.emitBlock(fmt.Sprintf("body (%s)", name), node.Func.Body)

	val := &mir.Fun{
		params,
		blk,
		false,
	}

	e.env.DeclTable[name] = ty
	insn := mir.NewInsn(name, val, node.Pos())

	body := e.emitInsn(node.Body)
	body.Append(insn)
	return body
}

func (e *emitter) emitMatchInsn(node *ast.Match) *mir.Insn {
	pos := node.Pos()
	matched := e.emitInsn(node.Target)
	id := e.genID()
	e.env.DeclTable[id] = types.BoolType
	cond := mir.Concat(mir.NewInsn(id, &mir.IsSome{matched.Ident}, pos), matched)

	// TODO: Do not emit insn if it's unused generic decl
	matchedTy, ok := e.env.DeclTable[matched.Ident].(*types.Option)
	if !ok {
		panic("Type of 'match' expression target not found")
	}
	name := node.SomeIdent.Name
	e.env.DeclTable[name] = matchedTy.Elem

	derefInsn := mir.NewInsn(name, &mir.DerefSome{matched.Ident}, pos)
	someBlk := e.emitBlock("then", node.IfSome)
	someBlk.Prepend(derefInsn)

	noneBlk := e.emitBlock("else", node.IfNone)

	return e.insn(&mir.If{cond.Ident, someBlk, noneBlk}, cond, node)
}

func (e *emitter) emitLetTupleInsn(node *ast.LetTuple) *mir.Insn {
	if len(node.Symbols) == 0 {
		panic("FATAL: LetTuple node must contain at least one symbol")
	}

	bound := e.emitInsn(node.Bound)
	boundTy, ok := e.typeOf(node.Bound).(*types.Tuple)
	if !ok {
		panic("FATAL: LetTuple node did not bind symbols to tuple value")
	}

	insn := bound
	for i, sym := range node.Symbols {
		// TODO: Do not emit insn if it's unused generic decl
		name := sym.Name
		e.env.DeclTable[name] = boundTy.Elems[i]
		insn = mir.Concat(mir.NewInsn(
			name,
			&mir.TplLoad{
				From:  bound.Ident,
				Index: i,
			},
			node.Pos(),
		), insn)
	}

	body := e.emitInsn(node.Body)
	body.Append(insn)
	return body
}

func (e *emitter) emitAppInsn(node *ast.Apply) *mir.Insn {
	var prev *mir.Insn
	var inst *types.Instantiation
	var ident string
	if ref, ok := node.Callee.(*ast.VarRef); ok {
		// Note:
		// When calling a variable directly, it may be direct call of a known function.
		// Known function is optimized in closure transform. So we set name of variable
		// reference directly to the callee of 'app' instruction. When callee is a polymorphic
		// function and needs to be instantiated, what type the callee is instantiated should
		// be maintained for monomorphization. Here we set the identifier of 'app' instruction
		// as the key of the instantiation. It is used to know how the callee was instantiated
		// while monomorphization.
		if _, ok := e.env.DeclTable[ref.Symbol.Name]; ok {
			ident = ref.Symbol.Name
			inst, _ = e.insts[ref]
		} else if _, ok := e.env.Externals[ref.Symbol.Name]; ok {
			prev = e.insn(&mir.XRef{ref.Symbol.Name}, nil, ref)
			ident = prev.Ident
		} else {
			panic("FATAL: Unknown identifier: " + ref.Symbol.Name)
		}
	} else {
		prev = e.emitInsn(node.Callee)
		ident = prev.Ident
	}
	args := make([]string, 0, len(node.Args))
	for _, a := range node.Args {
		arg := e.emitInsn(a)
		arg.Append(prev)
		args = append(args, arg.Ident)
		prev = arg
	}
	insn := e.insn(&mir.App{ident, args, mir.DIRECT_CALL}, prev, node)
	if inst != nil {
		e.env.RefInsts[insn.Ident] = inst
	}
	return insn
}

func (e *emitter) emitInsn(node ast.Expr) *mir.Insn {
	switch n := node.(type) {
	case *ast.Unit:
		return e.insn(mir.UnitVal, nil, node)
	case *ast.Bool:
		return e.insn(&mir.Bool{n.Value}, nil, node)
	case *ast.Int:
		return e.insn(&mir.Int{n.Value}, nil, node)
	case *ast.Float:
		return e.insn(&mir.Float{n.Value}, nil, node)
	case *ast.String:
		return e.insn(&mir.String{n.Value}, nil, node)
	case *ast.Not:
		i := e.emitInsn(n.Child)
		return e.insn(&mir.Unary{mir.NOT, i.Ident}, i, node)
	case *ast.Neg:
		i := e.emitInsn(n.Child)
		return e.insn(&mir.Unary{mir.NEG, i.Ident}, i, node)
	case *ast.FNeg:
		i := e.emitInsn(n.Child)
		return e.insn(&mir.Unary{mir.FNEG, i.Ident}, i, node)
	case *ast.Add:
		return e.emitBinaryInsn(mir.ADD, n.Left, n.Right, node)
	case *ast.Sub:
		return e.emitBinaryInsn(mir.SUB, n.Left, n.Right, node)
	case *ast.Mul:
		return e.emitBinaryInsn(mir.MUL, n.Left, n.Right, node)
	case *ast.Div:
		return e.emitBinaryInsn(mir.DIV, n.Left, n.Right, node)
	case *ast.Mod:
		return e.emitBinaryInsn(mir.MOD, n.Left, n.Right, node)
	case *ast.FAdd:
		return e.emitBinaryInsn(mir.FADD, n.Left, n.Right, node)
	case *ast.FSub:
		return e.emitBinaryInsn(mir.FSUB, n.Left, n.Right, node)
	case *ast.FMul:
		return e.emitBinaryInsn(mir.FMUL, n.Left, n.Right, node)
	case *ast.FDiv:
		return e.emitBinaryInsn(mir.FDIV, n.Left, n.Right, node)
	case *ast.Less:
		return e.emitBinaryInsn(mir.LT, n.Left, n.Right, node)
	case *ast.LessEq:
		return e.emitBinaryInsn(mir.LTE, n.Left, n.Right, node)
	case *ast.Greater:
		return e.emitBinaryInsn(mir.GT, n.Left, n.Right, node)
	case *ast.GreaterEq:
		return e.emitBinaryInsn(mir.GTE, n.Left, n.Right, node)
	case *ast.And:
		return e.emitBinaryInsn(mir.AND, n.Left, n.Right, node)
	case *ast.Or:
		return e.emitBinaryInsn(mir.OR, n.Left, n.Right, node)
	case *ast.Eq:
		return e.emitBinaryInsn(mir.EQ, n.Left, n.Right, node)
	case *ast.NotEq:
		return e.emitBinaryInsn(mir.NEQ, n.Left, n.Right, node)
	case *ast.If:
		prev := e.emitInsn(n.Cond)
		thenBlk := e.emitBlock("then", n.Then)
		elseBlk := e.emitBlock("else", n.Else)
		val := &mir.If{
			prev.Ident,
			thenBlk,
			elseBlk,
		}
		return e.insn(val, prev, node)
	case *ast.Let:
		return e.emitLetInsn(n)
	case *ast.VarRef:
		if _, ok := e.env.DeclTable[n.Symbol.Name]; ok {
			insn := e.insn(&mir.Ref{n.Symbol.Name}, nil, node)
			if inst, ok := e.insts[n]; ok {
				e.env.RefInsts[insn.Ident] = inst
			}
			return insn
		} else if _, ok := e.env.Externals[n.Symbol.Name]; ok {
			return e.insn(&mir.XRef{n.Symbol.Name}, nil, node)
		} else {
			panic("FATAL: Unknown identifier: " + n.Symbol.Name)
		}
	case *ast.LetRec:
		return e.emitFunInsn(n)
	case *ast.Apply:
		return e.emitAppInsn(n)
	case *ast.Tuple:
		var prev *mir.Insn
		len := len(n.Elems)
		elems := make([]string, 0, len)
		for _, elem := range n.Elems {
			i := e.emitInsn(elem)
			i.Append(prev)
			elems = append(elems, i.Ident)
			prev = i
		}
		return e.insn(&mir.Tuple{elems}, prev, node)
	case *ast.ArrayLit:
		if len(n.Elems) == 0 {
			return e.insn(&mir.ArrLit{}, nil, node)
		}
		var prev *mir.Insn
		elems := make([]string, 0, len(n.Elems))
		for _, elem := range n.Elems {
			i := e.emitInsn(elem)
			i.Append(prev)
			elems = append(elems, i.Ident)
			prev = i
		}
		return e.insn(&mir.ArrLit{elems}, prev, node)
	case *ast.LetTuple:
		return e.emitLetTupleInsn(n)
	case *ast.ArrayMake:
		size := e.emitInsn(n.Size)
		elem := e.emitInsn(n.Elem)
		elem.Append(size)
		return e.insn(&mir.Array{size.Ident, elem.Ident}, elem, node)
	case *ast.ArrayGet:
		array := e.emitInsn(n.Array)
		index := e.emitInsn(n.Index)
		index.Append(array)
		return e.insn(&mir.ArrLoad{array.Ident, index.Ident}, index, node)
	case *ast.ArrayPut:
		array := e.emitInsn(n.Array)
		index := e.emitInsn(n.Index)
		index.Append(array)
		rhs := e.emitInsn(n.Assignee)
		rhs.Append(index)
		return e.insn(&mir.ArrStore{array.Ident, index.Ident, rhs.Ident}, rhs, node)
	case *ast.ArraySize:
		array := e.emitInsn(n.Target)
		return e.insn(&mir.ArrLen{array.Ident}, array, node)
	case *ast.Some:
		child := e.emitInsn(n.Child)
		return e.insn(&mir.Some{child.Ident}, child, node)
	case *ast.None:
		return e.insn(mir.NoneVal, nil, node)
	case *ast.Match:
		return e.emitMatchInsn(n)
	case *ast.Typed:
		return e.emitInsn(n.Child)
	default:
		panic("FATAL: Unknown node: " + node.Name())
	}
}

// Returns mir.Block instance and its type
func (e *emitter) emitBlock(name string, node ast.Expr) *mir.Block {
	lastInsn := e.emitInsn(node)
	// emitInsn() emits instructions in descending order.
	// Reverse the order to iterate instractions ascending order.
	firstInsn := mir.Reverse(lastInsn)
	return mir.NewBlock(name, firstInsn, lastInsn)
}

// ToMIR converts given AST into MIR with type environment
func ToMIR(root ast.Expr, env *types.Env, inferred InferredTypes, insts refInsts) *mir.Block {
	e := &emitter{0, env, inferred, insts}
	return e.emitBlock("program", root)
}
