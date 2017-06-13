package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
)

// Convert AST into MIR with K-Normalization

type emitter struct {
	count    uint
	env      *types.Env
	inferred exprTypes
	err      *locerr.Error
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

func (e *emitter) emitBinaryInsn(op mir.OperatorKind, lhs ast.Expr, rhs ast.Expr) (mir.Val, *mir.Insn) {
	l := e.emitInsn(lhs)
	r := e.emitInsn(rhs)
	r.Append(l)
	return &mir.Binary{op, l.Ident, r.Ident}, r
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
	t, found := e.env.Table[bound.Ident]
	delete(e.env.Table, bound.Ident)

	bound.Ident = node.Symbol.Name
	if found {
		e.env.Table[bound.Ident] = t
	}

	body := e.emitInsn(node.Body)
	body.Append(bound)
	return body
}

func (e *emitter) emitFunInsn(node *ast.LetRec) *mir.Insn {
	name := node.Func.Symbol.Name
	ty, ok := e.env.Table[name]
	if !ok {
		// Note: Symbol in LetRec cannot be an external symbol.
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

	e.env.Table[name] = ty
	insn := mir.NewInsn(name, val, node.Pos())

	body := e.emitInsn(node.Body)
	body.Append(insn)
	return body
}

func (e *emitter) emitMatchInsn(node *ast.Match) (mir.Val, *mir.Insn) {
	pos := node.Pos()
	matched := e.emitInsn(node.Target)
	id := e.genID()
	e.env.Table[id] = types.BoolType
	cond := mir.Concat(mir.NewInsn(id, &mir.IsSome{matched.Ident}, pos), matched)

	matchedTy, ok := e.env.Table[matched.Ident].(*types.Option)
	if !ok {
		panic("Type of 'match' expression target not found")
	}
	name := node.SomeIdent.Name
	e.env.Table[name] = matchedTy.Elem

	derefInsn := mir.NewInsn(name, &mir.DerefSome{matched.Ident}, pos)
	someBlk := e.emitBlock("then", node.IfSome)
	someBlk.Prepend(derefInsn)

	noneBlk := e.emitBlock("else", node.IfNone)

	return &mir.If{cond.Ident, someBlk, noneBlk}, cond
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
		name := sym.Name
		e.env.Table[name] = boundTy.Elems[i]
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

func (e *emitter) emitInsn(node ast.Expr) *mir.Insn {
	var prev *mir.Insn
	var val mir.Val

	switch n := node.(type) {
	case *ast.Unit:
		val = mir.UnitVal
	case *ast.Bool:
		val = &mir.Bool{n.Value}
	case *ast.Int:
		val = &mir.Int{n.Value}
	case *ast.Float:
		val = &mir.Float{n.Value}
	case *ast.String:
		val = &mir.String{n.Value}
	case *ast.Not:
		i := e.emitInsn(n.Child)
		val = &mir.Unary{mir.NOT, i.Ident}
		prev = i
	case *ast.Neg:
		i := e.emitInsn(n.Child)
		val = &mir.Unary{mir.NEG, i.Ident}
		prev = i
	case *ast.FNeg:
		i := e.emitInsn(n.Child)
		val = &mir.Unary{mir.FNEG, i.Ident}
		prev = i
	case *ast.Add:
		val, prev = e.emitBinaryInsn(mir.ADD, n.Left, n.Right)
	case *ast.Sub:
		val, prev = e.emitBinaryInsn(mir.SUB, n.Left, n.Right)
	case *ast.Mul:
		val, prev = e.emitBinaryInsn(mir.MUL, n.Left, n.Right)
	case *ast.Div:
		val, prev = e.emitBinaryInsn(mir.DIV, n.Left, n.Right)
	case *ast.Mod:
		val, prev = e.emitBinaryInsn(mir.MOD, n.Left, n.Right)
	case *ast.FAdd:
		val, prev = e.emitBinaryInsn(mir.FADD, n.Left, n.Right)
	case *ast.FSub:
		val, prev = e.emitBinaryInsn(mir.FSUB, n.Left, n.Right)
	case *ast.FMul:
		val, prev = e.emitBinaryInsn(mir.FMUL, n.Left, n.Right)
	case *ast.FDiv:
		val, prev = e.emitBinaryInsn(mir.FDIV, n.Left, n.Right)
	case *ast.Less:
		val, prev = e.emitBinaryInsn(mir.LT, n.Left, n.Right)
	case *ast.LessEq:
		val, prev = e.emitBinaryInsn(mir.LTE, n.Left, n.Right)
	case *ast.Greater:
		val, prev = e.emitBinaryInsn(mir.GT, n.Left, n.Right)
	case *ast.GreaterEq:
		val, prev = e.emitBinaryInsn(mir.GTE, n.Left, n.Right)
	case *ast.And:
		val, prev = e.emitBinaryInsn(mir.AND, n.Left, n.Right)
	case *ast.Or:
		val, prev = e.emitBinaryInsn(mir.OR, n.Left, n.Right)
	case *ast.Eq:
		val, prev = e.emitBinaryInsn(mir.EQ, n.Left, n.Right)
	case *ast.NotEq:
		val, prev = e.emitBinaryInsn(mir.NEQ, n.Left, n.Right)
	case *ast.If:
		prev = e.emitInsn(n.Cond)
		thenBlk := e.emitBlock("then", n.Then)
		elseBlk := e.emitBlock("else", n.Else)
		val = &mir.If{
			prev.Ident,
			thenBlk,
			elseBlk,
		}
	case *ast.Let:
		return e.emitLetInsn(n)
	case *ast.VarRef:
		if _, ok := e.env.Table[n.Symbol.Name]; ok {
			val = &mir.Ref{n.Symbol.Name}
		} else if _, ok := e.env.Externals[n.Symbol.Name]; ok {
			val = &mir.XRef{n.Symbol.Name}
		} else {
			panic("FATAL: Unknown identifier: " + n.Symbol.Name)
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
	case *ast.Tuple:
		len := len(n.Elems)
		if len == 0 {
			panic("Tuple must not be empty!")
		}
		elems := make([]string, 0, len)
		for _, elem := range n.Elems {
			i := e.emitInsn(elem)
			i.Append(prev)
			elems = append(elems, i.Ident)
			prev = i
		}
		val = &mir.Tuple{elems}
	case *ast.ArrayLit:
		if len(n.Elems) == 0 {
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
		val = &mir.ArrLit{elems}
	case *ast.LetTuple:
		return e.emitLetTupleInsn(n)
	case *ast.ArrayMake:
		size := e.emitInsn(n.Size)
		elem := e.emitInsn(n.Elem)
		elem.Append(size)
		prev = elem
		val = &mir.Array{size.Ident, elem.Ident}
	case *ast.Get:
		array := e.emitInsn(n.Array)
		index := e.emitInsn(n.Index)
		index.Append(array)
		prev = index
		val = &mir.ArrLoad{array.Ident, index.Ident}
	case *ast.Put:
		array := e.emitInsn(n.Array)
		index := e.emitInsn(n.Index)
		index.Append(array)
		rhs := e.emitInsn(n.Assignee)
		rhs.Append(index)
		prev = rhs
		val = &mir.ArrStore{array.Ident, index.Ident, rhs.Ident}
	case *ast.ArraySize:
		array := e.emitInsn(n.Target)
		prev = array
		val = &mir.ArrLen{array.Ident}
	case *ast.Some:
		child := e.emitInsn(n.Child)
		prev = child
		val = &mir.Some{child.Ident}
	case *ast.None:
		val = mir.NoneVal
	case *ast.Match:
		val, prev = e.emitMatchInsn(n)
	case *ast.Typed:
		return e.emitInsn(n.Child)
	}

	if val == nil {
		panic("FATAL: Value in instruction must not be nil!")
	}

	ty := e.typeOf(node)
	id := e.genID()
	e.env.Table[id] = ty
	return mir.Concat(mir.NewInsn(id, val, node.Pos()), prev)
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
func ToMIR(root ast.Expr, env *types.Env, inferred exprTypes) (*mir.Block, error) {
	e := &emitter{0, env, inferred, nil}
	b := e.emitBlock("program", root)
	if e.err != nil {
		return nil, e.err.Note("Semantics error while MIR generation")
	}
	return b, nil
}
