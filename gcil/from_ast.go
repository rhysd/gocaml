package gcil

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"strings"
)

// Convert AST into GCIL with K-Normalization

type emitter struct {
	count  uint
	types  *typing.Env
	errors []string
}

func (e *emitter) genID() string {
	e.count++
	return fmt.Sprintf("$k%d", e.count)
}

func (e *emitter) typeOf(i *Insn) typing.Type {
	t, ok := e.types.Table[i.Ident]
	if !ok {
		panic(fmt.Sprintf("Type for '%s' not found for %v (bug)", i.Ident, *i))
	}
	return t
}

func (e *emitter) semanticError(msg string, pos token.Position) {
	e.errors = append(
		e.errors, fmt.Sprintf("%s at (line:%d, column:%d)", msg, pos.Line, pos.Column),
	)
}

func (e *emitter) emitBinaryInsn(op OperatorKind, lhs ast.Expr, rhs ast.Expr) (typing.Type, Val, *Insn) {
	l := e.emitInsn(lhs)
	r := e.emitInsn(rhs)
	r.Append(l)
	return e.typeOf(l), &Binary{op, l.Ident, r.Ident}, r
}

func (e *emitter) emitLetInsn(node *ast.Let) *Insn {
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

func (e *emitter) emitFunInsn(node *ast.LetRec) *Insn {
	name := node.Func.Symbol.Name

	ty, ok := e.types.Table[name]
	if !ok {
		// Note: Symbol in LetRec cannot be an external symbol.
		panic(fmt.Sprintf("Unknown function %s", name))
	}

	params := make([]string, 0, len(node.Func.Params))
	for _, s := range node.Func.Params {
		params = append(params, s.Name)
	}

	blk, _ := e.emitBlock(fmt.Sprintf("body (%s)", name), node.Func.Body)

	val := &Fun{
		params,
		blk,
		false,
	}

	e.types.Table[name] = ty
	insn := NewInsn(name, val, node.Pos())

	body := e.emitInsn(node.Body)
	body.Append(insn)
	return body
}

func (e *emitter) emitLetTupleInsn(node *ast.LetTuple) *Insn {
	if len(node.Symbols) == 0 {
		panic("LetTuple node must contain at least one symbol")
	}

	bound := e.emitInsn(node.Bound)
	boundTy, ok := e.typeOf(bound).(*typing.Tuple)
	if !ok {
		panic("LetTuple node does not bound symbols to tuple value")
	}

	insn := bound
	for i, sym := range node.Symbols {
		name := sym.Name
		e.types.Table[name] = boundTy.Elems[i]
		insn = Concat(NewInsn(
			name,
			&TplLoad{
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

func (e *emitter) emitLessInsn(kind OperatorKind, lhs, rhs ast.Expr) (typing.Type, Val, *Insn) {
	operand, val, prev := e.emitBinaryInsn(kind, lhs, rhs)
	// Note:
	// This type constraint may be useful for type inference. But current HM type inference algorithm cannot
	// handle a union type. In this context, the operand should be `int | float`
	switch operand.(type) {
	case *typing.Unit, *typing.Bool, *typing.Fun, *typing.Tuple, *typing.Array:
		e.semanticError(fmt.Sprintf("'%s' can't be compared with operator '%s'", operand.String(), OpTable[kind]), lhs.Pos())
	}
	return typing.BoolType, val, prev
}

func (e *emitter) emitEqInsn(kind OperatorKind, lhs, rhs ast.Expr) (typing.Type, Val, *Insn) {
	operand, val, prev := e.emitBinaryInsn(kind, lhs, rhs)
	// Note:
	// This type constraint may be useful for type inference. But current HM type inference algorithm cannot
	// handle a union type. In this context, the operand should be `() | bool | int | float | fun<R, TS...> | tuple<Args...>`
	if _, ok := operand.(*typing.Array); ok {
		e.semanticError(fmt.Sprintf("'%s' can't be compared with operator '%s'", operand.String(), OpTable[kind]), lhs.Pos())
	}
	return typing.BoolType, val, prev
}

func (e *emitter) emitInsn(node ast.Expr) *Insn {
	var prev *Insn = nil
	var val Val
	var ty typing.Type

	switch n := node.(type) {
	case *ast.Unit:
		ty = typing.UnitType
		val = UnitVal
	case *ast.Bool:
		ty = typing.BoolType
		val = &Bool{n.Value}
	case *ast.Int:
		ty = typing.IntType
		val = &Int{n.Value}
	case *ast.Float:
		ty = typing.FloatType
		val = &Float{n.Value}
	case *ast.Not:
		i := e.emitInsn(n.Child)
		ty, val = e.typeOf(i), &Unary{NOT, i.Ident}
		prev = i
	case *ast.Neg:
		i := e.emitInsn(n.Child)
		ty, val = e.typeOf(i), &Unary{NEG, i.Ident}
		prev = i
	case *ast.FNeg:
		i := e.emitInsn(n.Child)
		ty, val = e.typeOf(i), &Unary{FNEG, i.Ident}
		prev = i
	case *ast.Add:
		ty, val, prev = e.emitBinaryInsn(ADD, n.Left, n.Right)
	case *ast.Sub:
		ty, val, prev = e.emitBinaryInsn(SUB, n.Left, n.Right)
	case *ast.Mul:
		ty, val, prev = e.emitBinaryInsn(MUL, n.Left, n.Right)
	case *ast.Div:
		ty, val, prev = e.emitBinaryInsn(DIV, n.Left, n.Right)
	case *ast.FAdd:
		ty, val, prev = e.emitBinaryInsn(FADD, n.Left, n.Right)
	case *ast.FSub:
		ty, val, prev = e.emitBinaryInsn(FSUB, n.Left, n.Right)
	case *ast.FMul:
		ty, val, prev = e.emitBinaryInsn(FMUL, n.Left, n.Right)
	case *ast.FDiv:
		ty, val, prev = e.emitBinaryInsn(FDIV, n.Left, n.Right)
	case *ast.Less:
		ty, val, prev = e.emitLessInsn(LT, n.Left, n.Right)
	case *ast.LessEq:
		ty, val, prev = e.emitLessInsn(LTE, n.Left, n.Right)
	case *ast.Greater:
		ty, val, prev = e.emitLessInsn(GT, n.Left, n.Right)
	case *ast.GreaterEq:
		ty, val, prev = e.emitLessInsn(GTE, n.Left, n.Right)
	case *ast.And:
		ty, val, prev = e.emitBinaryInsn(AND, n.Left, n.Right)
	case *ast.Or:
		ty, val, prev = e.emitBinaryInsn(OR, n.Left, n.Right)
	case *ast.Eq:
		ty, val, prev = e.emitEqInsn(EQ, n.Left, n.Right)
	case *ast.NotEq:
		ty, val, prev = e.emitEqInsn(NEQ, n.Left, n.Right)
	case *ast.If:
		prev = e.emitInsn(n.Cond)
		thenBlk, t := e.emitBlock("then", n.Then)
		elseBlk, _ := e.emitBlock("else", n.Else)
		ty = t
		val = &If{
			prev.Ident,
			thenBlk,
			elseBlk,
		}
	case *ast.Let:
		return e.emitLetInsn(n)
	case *ast.VarRef:
		if t, ok := e.types.Table[n.Symbol.Name]; ok {
			ty = t
			val = &Ref{n.Symbol.Name}
		} else if t, ok := e.types.Externals[n.Symbol.Name]; ok {
			ty = t
			val = &XRef{n.Symbol.Name}
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
		val = &App{callee.Ident, args, DIRECT_CALL}
		f, ok := e.typeOf(callee).(*typing.Fun)
		if !ok {
			panic(fmt.Sprintf("Callee of Apply node is not typed as function!: %s", e.typeOf(callee).String()))
		}
		ty = f.Ret
	case *ast.Tuple:
		if len(n.Elems) == 0 {
			panic("Tuple must not be empty!")
		}
		prev = e.emitInsn(n.Elems[0])
		elems := []string{prev.Ident}
		types := []typing.Type{e.typeOf(prev)}
		for _, elem := range n.Elems[1:] {
			elemInsn := e.emitInsn(elem)
			elemInsn.Append(prev)
			elems = append(elems, elemInsn.Ident)
			types = append(types, e.typeOf(elemInsn))
			prev = elemInsn
		}
		ty = &typing.Tuple{types}
		val = &Tuple{elems}
	case *ast.LetTuple:
		return e.emitLetTupleInsn(n)
	case *ast.ArrayCreate:
		size := e.emitInsn(n.Size)
		elem := e.emitInsn(n.Elem)
		elem.Append(size)
		prev = elem
		ty = &typing.Array{e.typeOf(elem)}
		val = &Array{size.Ident, elem.Ident}
	case *ast.Get:
		array := e.emitInsn(n.Array)
		arrayTy, ok := e.typeOf(array).(*typing.Array)
		if !ok {
			panic("'Get' node does not access to array!")
		}
		index := e.emitInsn(n.Index)
		index.Append(array)
		prev = index
		ty = arrayTy.Elem
		val = &ArrLoad{array.Ident, index.Ident}
	case *ast.Put:
		array := e.emitInsn(n.Array)
		arrayTy, ok := e.typeOf(array).(*typing.Array)
		if !ok {
			panic("'Put' node does not access to array!")
		}
		index := e.emitInsn(n.Index)
		index.Append(array)
		rhs := e.emitInsn(n.Assignee)
		rhs.Append(index)
		prev = rhs
		ty = arrayTy.Elem
		val = &ArrStore{array.Ident, index.Ident, rhs.Ident}
	}

	// Note:
	// ty may be nil when it's for unused variable introduced in semicolon
	// expression.
	if val == nil {
		panic("Value in instruction must not be nil!")
	}
	id := e.genID()
	e.types.Table[id] = ty
	return Concat(NewInsn(id, val, node.Pos()), prev)
}

// Return Block instance and its type
func (e *emitter) emitBlock(name string, node ast.Expr) (*Block, typing.Type) {
	lastInsn := e.emitInsn(node)
	firstInsn := Reverse(lastInsn)
	// emitInsn() emits instructions in descending order.
	// Reverse the order to iterate instractions ascending order.
	return NewBlock(name, firstInsn, lastInsn), e.typeOf(lastInsn)
}

func FromAST(root ast.Expr, types *typing.Env) (*Block, error) {
	e := &emitter{0, types, []string{}}
	b, _ := e.emitBlock("program", root)
	if len(e.errors) > 0 {
		return nil, errors.New("Semantics error:\n  " + strings.Join(e.errors, "\n  "))
	}
	return b, nil
}
