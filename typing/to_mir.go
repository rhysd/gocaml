package typing

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
)

// Convert AST into MIR with K-Normalization

type emitter struct {
	count uint
	types *types.Env
	err   *locerr.Error
}

func (e *emitter) genID() string {
	e.count++
	return fmt.Sprintf("$k%d", e.count)
}

func (e *emitter) typeOf(i *mir.Insn) types.Type {
	t, ok := e.types.Table[i.Ident]
	if !ok {
		panic(fmt.Sprintf("Type for '%s' not found for %v (bug)", i.Ident, *i))
	}
	return t
}

func (e *emitter) semanticError(msg string, node ast.Expr) {
	if e.err == nil {
		e.err = locerr.ErrorIn(node.Pos(), node.End(), msg)
		return
	}
	e.err = e.err.NoteAt(node.Pos(), msg)
}

func (e *emitter) emitBinaryInsn(op mir.OperatorKind, lhs ast.Expr, rhs ast.Expr) (types.Type, mir.Val, *mir.Insn) {
	l := e.emitInsn(lhs)
	r := e.emitInsn(rhs)
	r.Append(l)
	return e.typeOf(l), &mir.Binary{op, l.Ident, r.Ident}, r
}

func (e *emitter) emitLetInsn(node *ast.Let) *mir.Insn {
	// Note:
	// Instroduce shortcut about symbol to reduce number of instruction nodes.
	//
	// Before:
	//   $k1 = some_insn
	//   $sym$t1 = ref $k1
	//
	// After:
	//   $sym$t1 = some_insn
	bound := e.emitInsn(node.Bound)
	t, found := e.types.Table[bound.Ident]
	delete(e.types.Table, bound.Ident)

	bound.Ident = node.Symbol.Name
	if found {
		e.types.Table[bound.Ident] = t
	}

	body := e.emitInsn(node.Body)
	body.Append(bound)
	return body
}

func (e *emitter) emitFunInsn(node *ast.LetRec) *mir.Insn {
	name := node.Func.Symbol.Name

	ty, ok := e.types.Table[name]
	if !ok {
		// Note: Symbol in LetRec cannot be an external symbol.
		panic(fmt.Sprintf("Unknown function %s", name))
	}

	params := make([]string, 0, len(node.Func.Params))
	for _, s := range node.Func.Params {
		params = append(params, s.Ident.Name)
	}

	blk, _ := e.emitBlock(fmt.Sprintf("body (%s)", name), node.Func.Body)

	val := &mir.Fun{
		params,
		blk,
		false,
	}

	e.types.Table[name] = ty
	insn := mir.NewInsn(name, val, node.Pos())

	body := e.emitInsn(node.Body)
	body.Append(insn)
	return body
}

func (e *emitter) emitMatchInsn(node *ast.Match) (types.Type, mir.Val, *mir.Insn) {
	pos := node.Pos()
	matched := e.emitInsn(node.Target)
	id := e.genID()
	e.types.Table[id] = types.BoolType
	cond := mir.Concat(mir.NewInsn(id, &mir.IsSome{matched.Ident}, pos), matched)

	matchedTy, ok := e.types.Table[matched.Ident].(*types.Option)
	if !ok {
		panic("Type of 'match' expression target not found")
	}
	name := node.SomeIdent.Name
	e.types.Table[name] = matchedTy.Elem

	derefInsn := mir.NewInsn(name, &mir.DerefSome{matched.Ident}, pos)
	someBlk, _ := e.emitBlock("then", node.IfSome)
	someBlk.Prepend(derefInsn)

	noneBlk, armTy := e.emitBlock("else", node.IfNone)

	return armTy, &mir.If{cond.Ident, someBlk, noneBlk}, cond
}

func (e *emitter) emitLetTupleInsn(node *ast.LetTuple) *mir.Insn {
	if len(node.Symbols) == 0 {
		panic("LetTuple node must contain at least one symbol")
	}

	bound := e.emitInsn(node.Bound)
	boundTy, ok := e.typeOf(bound).(*types.Tuple)
	if !ok {
		panic("LetTuple node does not bound symbols to tuple value")
	}

	insn := bound
	for i, sym := range node.Symbols {
		name := sym.Name
		e.types.Table[name] = boundTy.Elems[i]
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

func (e *emitter) emitLessInsn(kind mir.OperatorKind, lhs, rhs, parent ast.Expr) (types.Type, mir.Val, *mir.Insn) {
	operand, val, prev := e.emitBinaryInsn(kind, lhs, rhs)
	// Note:
	// This type constraint may be useful for type inference. But current HM type inference algorithm cannot
	// handle a union type. In this context, the operand should be `int | float`
	switch operand.(type) {
	case *types.Unit, *types.Bool, *types.String, *types.Fun, *types.Tuple, *types.Array, *types.Option:
		e.semanticError(fmt.Sprintf("'%s' can't be compared with operator '%s'", operand.String(), mir.OpTable[kind]), parent)
	}
	return types.BoolType, val, prev
}

func (e *emitter) emitEqInsn(kind mir.OperatorKind, lhs, rhs, parent ast.Expr) (types.Type, mir.Val, *mir.Insn) {
	operand, val, prev := e.emitBinaryInsn(kind, lhs, rhs)
	// Note:
	// This type constraint may be useful for type inference. But current HM type inference algorithm cannot
	// handle a union type. In this context, the operand should be `() | bool | int | float | fun<R, TS...> | tuple<Args...>`
	if _, ok := operand.(*types.Array); ok {
		e.semanticError(fmt.Sprintf("'%s' can't be compared with operator '%s'", operand.String(), mir.OpTable[kind]), parent)
	}
	return types.BoolType, val, prev
}

func (e *emitter) emitInsn(node ast.Expr) *mir.Insn {
	var prev *mir.Insn
	var val mir.Val
	var ty types.Type

	switch n := node.(type) {
	case *ast.Unit:
		ty = types.UnitType
		val = mir.UnitVal
	case *ast.Bool:
		ty = types.BoolType
		val = &mir.Bool{n.Value}
	case *ast.Int:
		ty = types.IntType
		val = &mir.Int{n.Value}
	case *ast.Float:
		ty = types.FloatType
		val = &mir.Float{n.Value}
	case *ast.String:
		ty = types.StringType
		val = &mir.String{n.Value}
	case *ast.Not:
		i := e.emitInsn(n.Child)
		ty, val = e.typeOf(i), &mir.Unary{mir.NOT, i.Ident}
		prev = i
	case *ast.Neg:
		i := e.emitInsn(n.Child)
		ty, val = e.typeOf(i), &mir.Unary{mir.NEG, i.Ident}
		prev = i
	case *ast.FNeg:
		i := e.emitInsn(n.Child)
		ty, val = e.typeOf(i), &mir.Unary{mir.FNEG, i.Ident}
		prev = i
	case *ast.Add:
		ty, val, prev = e.emitBinaryInsn(mir.ADD, n.Left, n.Right)
	case *ast.Sub:
		ty, val, prev = e.emitBinaryInsn(mir.SUB, n.Left, n.Right)
	case *ast.Mul:
		ty, val, prev = e.emitBinaryInsn(mir.MUL, n.Left, n.Right)
	case *ast.Div:
		ty, val, prev = e.emitBinaryInsn(mir.DIV, n.Left, n.Right)
	case *ast.Mod:
		ty, val, prev = e.emitBinaryInsn(mir.MOD, n.Left, n.Right)
	case *ast.FAdd:
		ty, val, prev = e.emitBinaryInsn(mir.FADD, n.Left, n.Right)
	case *ast.FSub:
		ty, val, prev = e.emitBinaryInsn(mir.FSUB, n.Left, n.Right)
	case *ast.FMul:
		ty, val, prev = e.emitBinaryInsn(mir.FMUL, n.Left, n.Right)
	case *ast.FDiv:
		ty, val, prev = e.emitBinaryInsn(mir.FDIV, n.Left, n.Right)
	case *ast.Less:
		ty, val, prev = e.emitLessInsn(mir.LT, n.Left, n.Right, n)
	case *ast.LessEq:
		ty, val, prev = e.emitLessInsn(mir.LTE, n.Left, n.Right, n)
	case *ast.Greater:
		ty, val, prev = e.emitLessInsn(mir.GT, n.Left, n.Right, n)
	case *ast.GreaterEq:
		ty, val, prev = e.emitLessInsn(mir.GTE, n.Left, n.Right, n)
	case *ast.And:
		ty, val, prev = e.emitBinaryInsn(mir.AND, n.Left, n.Right)
	case *ast.Or:
		ty, val, prev = e.emitBinaryInsn(mir.OR, n.Left, n.Right)
	case *ast.Eq:
		ty, val, prev = e.emitEqInsn(mir.EQ, n.Left, n.Right, n)
	case *ast.NotEq:
		ty, val, prev = e.emitEqInsn(mir.NEQ, n.Left, n.Right, n)
	case *ast.If:
		prev = e.emitInsn(n.Cond)
		thenBlk, t := e.emitBlock("then", n.Then)
		elseBlk, _ := e.emitBlock("else", n.Else)
		ty = t
		val = &mir.If{
			prev.Ident,
			thenBlk,
			elseBlk,
		}
	case *ast.Let:
		return e.emitLetInsn(n)
	case *ast.VarRef:
		if t, ok := e.types.Table[n.Symbol.Name]; ok {
			ty = t
			val = &mir.Ref{n.Symbol.Name}
		} else if t, ok := e.types.Externals[n.Symbol.Name]; ok {
			ty = t
			val = &mir.XRef{n.Symbol.Name}
		} else {
			panic(fmt.Sprintf("Unknown identifier %s", n.Symbol.Name))
		}
	case *ast.LetRec:
		return e.emitFunInsn(n)
	case *ast.Apply:
		callee := e.emitInsn(n.Callee)
		prev = callee
		args := make([]string, 0, len(n.Args))
		for _, a := range n.Args {
			arg := e.emitInsn(a)
			arg.Append(prev)
			args = append(args, arg.Ident)
			prev = arg
		}
		val = &mir.App{callee.Ident, args, mir.DIRECT_CALL}
		f, ok := e.typeOf(callee).(*types.Fun)
		if !ok {
			panic(fmt.Sprintf("Callee of Apply node is not typed as function!: %s", e.typeOf(callee).String()))
		}
		ty = f.Ret
	case *ast.Tuple:
		len := len(n.Elems)
		if len == 0 {
			panic("Tuple must not be empty!")
		}
		elems := make([]string, 0, len)
		elemTypes := make([]types.Type, 0, len)
		for _, elem := range n.Elems {
			i := e.emitInsn(elem)
			i.Append(prev)
			elems = append(elems, i.Ident)
			elemTypes = append(elemTypes, e.typeOf(i))
			prev = i
		}
		ty = &types.Tuple{elemTypes}
		val = &mir.Tuple{elems}
	case *ast.ArrayLit:
		if len(n.Elems) == 0 {
			// Cannot know the type of empty array by bottom-up deduction. So we need to depend on
			// type hint here.
			var ok bool
			ty, ok = e.types.TypeHints[n]
			if !ok {
				panic("Type of empty array literal is unknown")
			}
			val = &mir.ArrLit{}
			break
		}
		elems := make([]string, 0, len(n.Elems))
		for _, elem := range n.Elems {
			i := e.emitInsn(elem)
			i.Append(prev)
			elems = append(elems, i.Ident)
			prev = i
		}
		ty = &types.Array{e.typeOf(prev)}
		val = &mir.ArrLit{elems}
	case *ast.LetTuple:
		return e.emitLetTupleInsn(n)
	case *ast.ArrayCreate:
		size := e.emitInsn(n.Size)
		elem := e.emitInsn(n.Elem)
		elem.Append(size)
		prev = elem
		ty = &types.Array{e.typeOf(elem)}
		val = &mir.Array{size.Ident, elem.Ident}
	case *ast.Get:
		array := e.emitInsn(n.Array)
		arrayTy, ok := e.typeOf(array).(*types.Array)
		if !ok {
			panic("'Get' node does not access to array!")
		}
		index := e.emitInsn(n.Index)
		index.Append(array)
		prev = index
		ty = arrayTy.Elem
		val = &mir.ArrLoad{array.Ident, index.Ident}
	case *ast.Put:
		array := e.emitInsn(n.Array)
		arrayTy, ok := e.typeOf(array).(*types.Array)
		if !ok {
			panic("'Put' node does not access to array!")
		}
		index := e.emitInsn(n.Index)
		index.Append(array)
		rhs := e.emitInsn(n.Assignee)
		rhs.Append(index)
		prev = rhs
		ty = arrayTy.Elem
		val = &mir.ArrStore{array.Ident, index.Ident, rhs.Ident}
	case *ast.ArraySize:
		array := e.emitInsn(n.Target)
		prev = array
		ty = types.IntType
		val = &mir.ArrLen{array.Ident}
	case *ast.Some:
		child := e.emitInsn(n.Child)
		prev = child
		childTy, ok := e.types.Table[child.Ident]
		if !ok {
			panic("Child type for 'Some' value is unknown")
		}
		ty = &types.Option{childTy}
		val = &mir.Some{child.Ident}
	case *ast.None:
		var ok bool
		ty, ok = e.types.TypeHints[n]
		if !ok {
			panic("Type of 'None' value is unknown")
		}
		val = mir.NoneVal
	case *ast.Match:
		ty, val, prev = e.emitMatchInsn(n)
	case *ast.Typed:
		return e.emitInsn(n.Child)
	}

	// Note:
	// ty may be nil when it's for unused variable introduced in semicolon
	// expression.
	if val == nil {
		panic("Value in instruction must not be nil!")
	}
	id := e.genID()
	e.types.Table[id] = ty
	return mir.Concat(mir.NewInsn(id, val, node.Pos()), prev)
}

// Returns mir.Block instance and its type
func (e *emitter) emitBlock(name string, node ast.Expr) (*mir.Block, types.Type) {
	lastInsn := e.emitInsn(node)
	firstInsn := mir.Reverse(lastInsn)
	// emitInsn() emits instructions in descending order.
	// Reverse the order to iterate instractions ascending order.
	return mir.NewBlock(name, firstInsn, lastInsn), e.typeOf(lastInsn)
}

// Convert given AST into MIR with type environment
func ToMIR(root ast.Expr, env *types.Env) (*mir.Block, error) {
	e := &emitter{0, env, nil}
	b, _ := e.emitBlock("program", root)
	if e.err != nil {
		return nil, e.err.Note("Semantics error while MIR generation")
	}
	return b, nil
}
