package typing

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/ast"
)

func typeError(err error, node ast.Expr) error {
	pos := node.Pos()
	return errors.Wrapf(err, "Type error at node '%s' (line:%d, column:%d)\n", node.Name(), pos.Line, pos.Column)
}

func (env *Env) checkNodeType(node ast.Expr, expected Type) error {
	t, err := env.infer(node)
	if err != nil {
		return err
	}
	if err = Unify(expected, t); err != nil {
		return typeError(err, node)
	}
	return nil
}

func (env *Env) inferArithmeticBinOp(left, right ast.Expr, operand Type) (Type, error) {
	l, err := env.infer(left)
	if err != nil {
		return nil, err
	}
	r, err := env.infer(right)
	if err != nil {
		return nil, err
	}
	if err = Unify(operand, l); err != nil {
		return nil, typeError(err, left)
	}
	if err = Unify(operand, r); err != nil {
		return nil, typeError(err, right)
	}
	// Returns the same type as operands
	return operand, nil
}

func (env *Env) inferRelationalBinOp(left, right ast.Expr) (Type, error) {
	l, err := env.infer(left)
	if err != nil {
		return nil, err
	}
	r, err := env.infer(right)
	if err != nil {
		return nil, err
	}
	if err = Unify(l, r); err != nil {
		return nil, typeError(err, left)
	}
	return BoolType, nil
}

func (env *Env) infer(e ast.Expr) (Type, error) {
	switch n := e.(type) {
	case *ast.Unit:
		return UnitType, nil
	case *ast.Int:
		return IntType, nil
	case *ast.Float:
		return FloatType, nil
	case *ast.Bool:
		return BoolType, nil
	case *ast.Not:
		if err := env.checkNodeType(n.Child, BoolType); err != nil {
			return nil, err
		}
		return BoolType, nil
	case *ast.Neg:
		if err := env.checkNodeType(n.Child, IntType); err != nil {
			return nil, err
		}
		return IntType, nil
	case *ast.Add:
		return env.inferArithmeticBinOp(n.Left, n.Right, IntType)
	case *ast.Sub:
		return env.inferArithmeticBinOp(n.Left, n.Right, IntType)
	case *ast.FNeg:
		if err := env.checkNodeType(n.Child, FloatType); err != nil {
			return nil, err
		}
		return FloatType, nil
	case *ast.FAdd:
		return env.inferArithmeticBinOp(n.Left, n.Right, FloatType)
	case *ast.FSub:
		return env.inferArithmeticBinOp(n.Left, n.Right, FloatType)
	case *ast.FMul:
		return env.inferArithmeticBinOp(n.Left, n.Right, FloatType)
	case *ast.FDiv:
		return env.inferArithmeticBinOp(n.Left, n.Right, FloatType)
	case *ast.Eq:
		return env.inferRelationalBinOp(n.Left, n.Right)
	case *ast.Less:
		return env.inferRelationalBinOp(n.Left, n.Right)
	case *ast.If:
		if err := env.checkNodeType(n.Cond, BoolType); err != nil {
			return nil, err
		}

		t, err := env.infer(n.Then)
		if err != nil {
			return nil, err
		}
		e, err := env.infer(n.Else)
		if err != nil {
			return nil, err
		}

		if err = Unify(t, e); err != nil {
			return nil, typeError(err, n)
		}

		return t, nil
	case *ast.Let:
		bound, err := env.infer(n.Bound)
		if err != nil {
			return nil, err
		}

		t := NewVar()
		if err = Unify(t, bound); err != nil {
			return nil, typeError(err, n.Body)
		}

		env.Table[n.Symbol.Name] = bound
		return env.infer(n.Body)
	case *ast.Var:
		if t, ok := env.Table[n.Ident]; ok {
			return t, nil
		}
		if t, ok := env.Externals[n.Ident]; ok {
			return t, nil
		}
		// Assume as free variable. If free variable's type is not identified,
		// It falls into compilation error
		t := NewVar()
		env.Externals[n.Ident] = t
		return t, nil
	case *ast.LetRec:
		f := NewVar()
		// Need to register function here because of recursive functions
		env.Table[n.Func.Symbol.Name] = f

		// Register parameters of function as variables to table
		params := make([]Type, len(n.Func.Params))
		for i, p := range n.Func.Params {
			// Types of parameters are unknown at definition
			t := NewVar()
			env.Table[p.Name] = t
			params[i] = t
		}

		// Infer return type of function from its body
		ret, err := env.infer(n.Func.Body)
		if err != nil {
			return nil, err
		}

		fun := &Fun{
			Params: params,
			Ret:    ret,
		}

		// n.Func.Type represents its function type. So unify it with
		// inferred function type from its parameters and body.
		if err = Unify(fun, f); err != nil {
			return nil, typeError(err, n)
		}

		return env.infer(n.Body)
	case *ast.Apply:
		args := make([]Type, len(n.Args))
		for i, a := range n.Args {
			t, err := env.infer(a)
			if err != nil {
				return nil, err
			}
			args[i] = t
		}

		// Return type of callee is unknown in this point.
		// So make a new type variable and allocate it as return type.
		ret := NewVar()
		fun := &Fun{
			Ret:    ret,
			Params: args,
		}

		callee, err := env.infer(n.Callee)
		if err != nil {
			return nil, err
		}

		if err = Unify(callee, fun); err != nil {
			return nil, typeError(err, n)
		}

		return ret, nil
	case *ast.Tuple:
		elems := make([]Type, len(n.Elems))
		for i, e := range n.Elems {
			t, err := env.infer(e)
			if err != nil {
				return nil, err
			}
			elems[i] = t
		}
		return &Tuple{Elems: elems}, nil
	case *ast.LetTuple:
		elems := make([]Type, len(n.Symbols))
		for i, sym := range n.Symbols {
			// Bound elements' types are unknown in this point
			t := NewVar()
			env.Table[sym.Name] = t
			elems[i] = t
		}

		// Bound value must be tuple
		if err := env.checkNodeType(n.Bound, &Tuple{Elems: elems}); err != nil {
			return nil, err
		}

		return env.infer(n.Body)
	case *ast.Array:
		if err := env.checkNodeType(n.Size, IntType); err != nil {
			return nil, err
		}
		elem, err := env.infer(n.Elem)
		if err != nil {
			return nil, err
		}
		return &Array{Elem: elem}, nil
	case *ast.Get:
		// Lhs of Get must be array but its element type is unknown.
		// So introduce new type variable for it.
		elem := NewVar()
		array := &Array{Elem: elem}

		if err := env.checkNodeType(n.Array, array); err != nil {
			return nil, err
		}

		if err := env.checkNodeType(n.Index, IntType); err != nil {
			return nil, err
		}

		return elem, nil
	case *ast.Put:
		if err := env.checkNodeType(n.Index, IntType); err != nil {
			return nil, err
		}
		assignee, err := env.infer(n.Assignee)
		if err != nil {
			return nil, err
		}

		// Type of assigned value must be the same as element type of the array
		array := &Array{Elem: assignee}
		if err := env.checkNodeType(n.Array, array); err != nil {
			return nil, err
		}

		// Assign to array does not have a value, so return unit type
		return UnitType, nil
	}
	panic(fmt.Sprintf("Unreachable: %s %v %v", e.Name(), e.Pos(), e.End()))
}
