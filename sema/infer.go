package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/common"
	. "github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
)

// InferredTypes is a dictonary from an AST nodes to inferred types.
type InferredTypes map[ast.Expr]Type

// Type schemes for generic types
type schemes map[Type]boundVarIDs

type refInsts map[*ast.VarRef]*Instantiation

// Inferer is a visitor to infer types in the AST
type Inferer struct {
	Env      *Env
	conv     *nodeTypeConv
	inferred InferredTypes
	// Map from generic type to bound type variables in the generic type
	schemes schemes
	insts   refInsts
}

// NewInferer creates a new Inferer instance
func NewInferer(env *Env) *Inferer {
	return &Inferer{
		env,
		nil,
		map[ast.Expr]Type{},
		map[Type]boundVarIDs{},
		refInsts{},
	}
}

func (inf *Inferer) generalize(t Type, level int) Type {
	t, bounds := generalize(t, level)
	if len(bounds) > 0 {
		inf.schemes[t] = bounds
	}
	return t
}

func (inf *Inferer) checkNodeType(where string, node ast.Expr, expected Type, level int) error {
	t, err := inf.infer(node, level)
	if err != nil {
		return err
	}
	if err := Unify(expected, t); err != nil {
		return err.In(node.Pos(), node.End()).NotefAt(node.Pos(), "Type error: %s must be '%s'", where, expected.String())
	}
	return nil
}

func (inf *Inferer) inferArithmeticBinOp(op string, left, right ast.Expr, operand Type, level int) (Type, error) {
	l, err := inf.infer(left, level)
	if err != nil {
		return nil, err
	}
	r, err := inf.infer(right, level)
	if err != nil {
		return nil, err
	}
	if err := Unify(operand, l); err != nil {
		return nil, err.In(left.Pos(), left.End()).NotefAt(left.Pos(), "Left hand of operator '%s' must be %s", op, operand.String())
	}
	if err := Unify(operand, r); err != nil {
		return nil, err.In(right.Pos(), right.End()).NotefAt(right.Pos(), "Right hand of operator '%s' must be %s", op, operand.String())
	}
	// Returns the same type as operands
	return operand, nil
}

func (inf *Inferer) inferRelationalBinOp(op string, left, right ast.Expr, level int) (Type, error) {
	l, err := inf.infer(left, level)
	if err != nil {
		return nil, err
	}
	r, err := inf.infer(right, level)
	if err != nil {
		return nil, err
	}
	if err := Unify(l, r); err != nil {
		return nil, err.In(left.Pos(), right.End()).NotefAt(left.Pos(), "Type mismatch at operands of relational operator '%s'", op)
	}
	return BoolType, nil
}

func (inf *Inferer) inferLogicalOp(op string, left, right ast.Expr, level int) (Type, error) {
	for i, e := range []ast.Expr{left, right} {
		t, err := inf.infer(e, level)
		if err != nil {
			return nil, err
		}
		if err := Unify(BoolType, t); err != nil {
			return nil, err.In(left.Pos(), right.End()).NotefAt(e.Pos(), "Type mismatch at %dth operand of logical operator '%s'", i+1, op)
		}
	}
	return BoolType, nil
}

func (inf *Inferer) inferNode(e ast.Expr, level int) (Type, error) {
	switch n := e.(type) {
	case *ast.Unit:
		return UnitType, nil
	case *ast.Int:
		return IntType, nil
	case *ast.Float:
		return FloatType, nil
	case *ast.String:
		return StringType, nil
	case *ast.Bool:
		return BoolType, nil
	case *ast.Not:
		if err := inf.checkNodeType("operand of operator 'not'", n.Child, BoolType, level); err != nil {
			return nil, err
		}
		return BoolType, nil
	case *ast.Neg:
		if err := inf.checkNodeType("operand of unary operator '-'", n.Child, IntType, level); err != nil {
			return nil, err
		}
		return IntType, nil
	case *ast.Add:
		return inf.inferArithmeticBinOp("+", n.Left, n.Right, IntType, level)
	case *ast.Sub:
		return inf.inferArithmeticBinOp("-", n.Left, n.Right, IntType, level)
	case *ast.Mul:
		return inf.inferArithmeticBinOp("*", n.Left, n.Right, IntType, level)
	case *ast.Div:
		return inf.inferArithmeticBinOp("/", n.Left, n.Right, IntType, level)
	case *ast.Mod:
		return inf.inferArithmeticBinOp("%", n.Left, n.Right, IntType, level)
	case *ast.FNeg:
		if err := inf.checkNodeType("operand of unary operator '-.'", n.Child, FloatType, level); err != nil {
			return nil, err
		}
		return FloatType, nil
	case *ast.FAdd:
		return inf.inferArithmeticBinOp("+.", n.Left, n.Right, FloatType, level)
	case *ast.FSub:
		return inf.inferArithmeticBinOp("-.", n.Left, n.Right, FloatType, level)
	case *ast.FMul:
		return inf.inferArithmeticBinOp("*.", n.Left, n.Right, FloatType, level)
	case *ast.FDiv:
		return inf.inferArithmeticBinOp("/.", n.Left, n.Right, FloatType, level)
	case *ast.Eq:
		return inf.inferRelationalBinOp("=", n.Left, n.Right, level)
	case *ast.NotEq:
		return inf.inferRelationalBinOp("<>", n.Left, n.Right, level)
	case *ast.Less:
		return inf.inferRelationalBinOp("<", n.Left, n.Right, level)
	case *ast.LessEq:
		return inf.inferRelationalBinOp("<=", n.Left, n.Right, level)
	case *ast.Greater:
		return inf.inferRelationalBinOp(">", n.Left, n.Right, level)
	case *ast.GreaterEq:
		return inf.inferRelationalBinOp(">=", n.Left, n.Right, level)
	case *ast.And:
		return inf.inferLogicalOp("&&", n.Left, n.Right, level)
	case *ast.Or:
		return inf.inferLogicalOp("||", n.Left, n.Right, level)
	case *ast.If:
		if err := inf.checkNodeType("condition of 'if' expression", n.Cond, BoolType, level); err != nil {
			return nil, err
		}

		t, err := inf.infer(n.Then, level)
		if err != nil {
			return nil, err
		}

		e, err := inf.infer(n.Else, level)
		if err != nil {
			return nil, err
		}

		if err := Unify(t, e); err != nil {
			return nil, err.In(n.Pos(), n.End()).NoteAt(n.Pos(), "Mismatch of types for 'then' clause and 'else' clause in 'if' expression")
		}

		return t, nil
	case *ast.Let:
		bound, err := inf.infer(n.Bound, level+1)
		if err != nil {
			return nil, err
		}

		if n.Type != nil {
			// When let x: type = ...
			t, err := inf.conv.nodeToType(n.Type, level)
			if err != nil {
				return nil, err
			}
			if err := Unify(t, bound); err != nil {
				b := n.Body
				return nil, err.In(b.Pos(), b.End()).NotefAt(b.Pos(), "Type of variable '%s'", n.Symbol.DisplayName)
			}
		}
		inf.Env.DeclTable[n.Symbol.Name] = inf.generalize(bound, level)

		return inf.infer(n.Body, level)
	case *ast.VarRef:
		if t, ok := inf.Env.DeclTable[n.Symbol.Name]; ok {
			inst := instantiate(t, level)
			if inst == nil {
				return t, nil
			}
			inf.insts[n] = inst
			return inst.To, nil
		}
		if e, ok := inf.Env.Externals[n.Symbol.Name]; ok {
			return e.Type, nil
		}
		panic("FATAL: Unknown symbol must be checked in alpha transform: " + n.Symbol.Name)
	case *ast.LetRec:
		// Note:
		// LetRec is different from other Let or LetTuple because it may be recursive.
		// It's the point to separe LetRec to recursive variable declaration and function expression.
		//   before: let rec f a b = a > b in ...
		//   after:  let rec f = fun a b -> a > b in ...
		// It means that type variables of parameters should be made with level + 1. And type variable
		// of return type is also. Then type of `f` should be generalized with level.

		// Register parameters of function as variables to table
		params := make([]Type, len(n.Func.Params))
		for i, p := range n.Func.Params {
			var t Type
			var err error
			if p.Type != nil {
				t, err = inf.conv.nodeToType(p.Type, level+1)
				if err != nil {
					return nil, locerr.NotefAt(p.Type.Pos(), err, "%s parameter of function '%s'", common.Ordinal(i+1), n.Func.Symbol.DisplayName)
				}
			} else {
				t = NewVar(nil, level+1)
			}
			inf.Env.DeclTable[p.Ident.Name] = t
			params[i] = t
		}

		var ret Type
		if n.Func.RetType != nil {
			r := n.Func.RetType
			t, err := inf.conv.nodeToType(r, level+1)
			if err != nil {
				return nil, locerr.NotefAt(r.Pos(), err, "Return type of function '%s'", n.Func.Symbol.DisplayName)
			}
			ret = t
		} else {
			ret = NewVar(nil, level+1)
		}

		// Considering recursive function call, register function name before inferring type of its
		// body. Register the function as a type variable here and later update the type with the
		// result of type inference for body of function.
		// Type of recursive function is *NOT* generic while inferring type of its body. For example,
		// `let rec f x = f 10 in f true` causes compilation error because of mismatch between 'int'
		// and 'bool'.
		fun := &Fun{ret, params}
		inf.Env.DeclTable[n.Func.Symbol.Name] = fun

		// Infer return type of function from its body
		ret2, err := inf.infer(n.Func.Body, level+1)
		if err != nil {
			return nil, err
		}

		if err := Unify(ret2, ret); err != nil {
			return nil, err.In(n.Pos(), n.End()).NotefAt(n.Pos(), "Return type of function '%s'", n.Func.Symbol.DisplayName)
		}

		// Update the return type with the result of type inference of function body. The function was
		// registered as non-polymorphic type for recursive call before inferring its body.
		inf.Env.DeclTable[n.Func.Symbol.Name] = inf.generalize(fun, level)

		return inf.infer(n.Body, level)
	case *ast.Apply:
		args := make([]Type, len(n.Args))
		for i, a := range n.Args {
			t, err := inf.infer(a, level)
			if err != nil {
				return nil, err
			}
			args[i] = t
		}

		// Return type of callee is unknown in this point.
		// So make a new type variable and allocate it as return type.
		ret := NewVar(nil, level)
		fun := &Fun{
			Ret:    ret,
			Params: args,
		}

		callee, err := inf.infer(n.Callee, level)
		if err != nil {
			return nil, err
		}

		if err := Unify(callee, fun); err != nil {
			return nil, err.In(n.Pos(), n.End()).NoteAt(n.Pos(), "Type of called function")
		}

		return ret, nil
	case *ast.Tuple:
		elems := make([]Type, len(n.Elems))
		for i, e := range n.Elems {
			t, err := inf.infer(e, level)
			if err != nil {
				return nil, err
			}
			elems[i] = t
		}
		return &Tuple{Elems: elems}, nil
	case *ast.LetTuple:
		var t *Tuple

		if n.Type != nil {
			ty, err := inf.conv.nodeToType(n.Type, level)
			if err != nil {
				return nil, err
			}

			var ok bool
			t, ok = ty.(*Tuple)
			if !ok {
				return nil, locerr.ErrorfIn(n.Type.Pos(), n.Type.End(), "Type error: Bound value of 'let (...) =' must be tuple, but found '%s'", t.String())
			}
			if len(t.Elems) != len(n.Symbols) {
				return nil, locerr.ErrorfIn(n.Type.Pos(), n.Type.End(), "Type error: Mismatch numbers of elements of specified tuple type and symbols in 'let (...)' expression: %d vs %d", len(t.Elems), len(n.Symbols))
			}
		} else {
			elems := make([]Type, len(n.Symbols))
			for i := range n.Symbols {
				// Bound elements' types are unknown in this point
				elems[i] = NewVar(nil, level+1)
			}
			t = &Tuple{Elems: elems}
		}

		bound, err := inf.infer(n.Bound, level+1)
		if err != nil {
			return nil, err
		}

		for i, sym := range n.Symbols {
			inf.Env.DeclTable[sym.Name] = inf.generalize(t.Elems[i], level)
		}

		// Bound value must be tuple
		if err := Unify(t, bound); err != nil {
			return nil, err.In(n.Pos(), n.End()).NotefAt(n.Pos(), "Type error: bound tuple value at 'let' must be '%s'", t.String())
		}

		return inf.infer(n.Body, level)
	case *ast.ArrayMake:
		if err := inf.checkNodeType("size at array creation", n.Size, IntType, level); err != nil {
			return nil, err
		}
		elem, err := inf.infer(n.Elem, level)
		if err != nil {
			return nil, err
		}
		return &Array{Elem: elem}, nil
	case *ast.ArraySize:
		if err := inf.checkNodeType("argument of 'Array.length'", n.Target, &Array{Elem: NewVar(nil, level)}, level); err != nil {
			return nil, err
		}
		return IntType, nil
	case *ast.ArrayGet:
		// Lhs of Get must be array but its element type is unknown.
		// So introduce new type variable for it.
		elem := NewVar(nil, level)
		array := &Array{Elem: elem}

		if err := inf.checkNodeType("array value in index access", n.Array, array, level); err != nil {
			return nil, err
		}

		if err := inf.checkNodeType("index access to array", n.Index, IntType, level); err != nil {
			return nil, err
		}

		return elem, nil
	case *ast.ArrayPut:
		if err := inf.checkNodeType("index at assignment to an element of array", n.Index, IntType, level); err != nil {
			return nil, err
		}
		assignee, err := inf.infer(n.Assignee, level)
		if err != nil {
			return nil, err
		}

		// Type of assigned value must be the same as element type of the array
		array := &Array{Elem: assignee}
		if err := inf.checkNodeType("assignment to an element of array", n.Array, array, level); err != nil {
			return nil, err
		}

		// Assign to array does not have a value, so return unit type
		return UnitType, nil
	case *ast.ArrayLit:
		if len(n.Elems) == 0 {
			// Array is empty. Cannot infer type of elements.
			return &Array{NewVar(nil, level)}, nil
		}
		elem, err := inf.infer(n.Elems[0], level)
		if err != nil {
			return nil, locerr.NoteAt(n.Pos(), err, "1st element type of array literal is incorrect")
		}
		for i, e := range n.Elems[1:] {
			t, err := inf.infer(e, level)
			if err != nil {
				return nil, locerr.NotefAt(e.Pos(), err, "%s element type of array literal is incorrect", common.Ordinal(i+2))
			}
			if err := Unify(elem, t); err != nil {
				return nil, err.In(e.Pos(), e.End()).NotefAt(e.Pos(), "Mismatch between 1st element and %s element in array literal", common.Ordinal(i+2))
			}
		}
		return &Array{elem}, nil
	case *ast.Some:
		elem, err := inf.infer(n.Child, level)
		if err != nil {
			return nil, err
		}
		return &Option{elem}, nil
	case *ast.None:
		return &Option{NewVar(nil, level)}, nil
	case *ast.Match:
		elem := NewVar(nil, level)
		matched := &Option{elem}
		if err := inf.checkNodeType("matching target in 'match' expression", n.Target, matched, level); err != nil {
			return nil, err
		}

		inf.Env.DeclTable[n.SomeIdent.Name] = elem
		some, err := inf.infer(n.IfSome, level)
		if err != nil {
			return nil, err
		}
		none, err := inf.infer(n.IfNone, level)
		if err != nil {
			return nil, err
		}
		if err := Unify(some, none); err != nil {
			return nil, err.In(n.Pos(), n.End()).NoteAt(n.Pos(), "Mismatch of types between 'Some' arm and 'None' arm in 'match' expression")
		}
		return some, nil
	case *ast.Typed:
		child, err := inf.infer(n.Child, level)
		if err != nil {
			return nil, err
		}

		t, err := inf.conv.nodeToType(n.Type, level)
		if err != nil {
			return nil, err
		}

		if err := Unify(t, child); err != nil {
			return nil, err.In(n.Pos(), n.End()).NoteAt(n.Pos(), "Mismatch between inferred type and specified type")
		}

		return child, nil
	default:
		panic(fmt.Sprintf("FATAL: Unreachable: %s %v %v", e.Name(), e.Pos(), e.End()))
	}
}

func (inf *Inferer) infer(e ast.Expr, level int) (Type, error) {
	t, err := inf.inferNode(e, level)
	if err != nil {
		return nil, err
	}
	inf.inferred[e] = t
	return t, nil
}

// Infer infers types in given AST and returns error when detecting type errors
func (inf *Inferer) Infer(parsed *ast.AST) error {
	var err error

	// TODO:
	// Move creating inf.conv to newInferer(). newInferer should receive *ast.AST and make
	// Inferer instance to call Infer().
	inf.conv, err = newNodeTypeConv(parsed.TypeDecls)
	if err != nil {
		return err
	}

	inf.conv.acceptsAnyType = false
	for _, ext := range parsed.Externals {
		t, err := inf.conv.nodeToType(ext.Type, -1)
		if err != nil {
			err = locerr.NotefAt(ext.Pos(), err, "Invalid type annotation at 'external' declaration '%s'", ext.Ident.Name)
			err = locerr.NoteAt(ext.Pos(), err, "'_' is not permitted in type of external symbol")
			return err
		}
		inf.Env.Externals[ext.Ident.Name] = &External{t, ext.C}
	}
	inf.conv.acceptsAnyType = true

	root, err := inf.infer(parsed.Root, 0)
	if err != nil {
		return err
	}

	if err := Unify(UnitType, root); err != nil {
		return err.At(parsed.Root.Pos()).Note("Type of root expression of program must be unit")
	}

	if err := derefTypeVars(inf.Env, parsed.Root, inf.inferred, inf.schemes, inf.insts); err != nil {
		return err
	}

	return nil
}
