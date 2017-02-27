package typing

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
	"strings"
)

func typeError(err error, where string, pos token.Position) error {
	return errors.Wrapf(err, "Type error: %s (line:%d, column:%d)\n", where, pos.Line, pos.Column)
}

func (env *Env) checkNodeType(where string, node ast.Expr, expected Type) error {
	t, err := env.infer(node)
	if err != nil {
		return err
	}
	if err = Unify(expected, t); err != nil {
		return typeError(err, fmt.Sprintf("%s must be '%s'", where, expected.String()), node.Pos())
	}
	return nil
}

func (env *Env) inferArithmeticBinOp(op string, left, right ast.Expr, operand Type) (Type, error) {
	l, err := env.infer(left)
	if err != nil {
		return nil, err
	}
	r, err := env.infer(right)
	if err != nil {
		return nil, err
	}
	if err = Unify(operand, l); err != nil {
		return nil, typeError(err, fmt.Sprintf("left hand of operator '%s' must be %s", op, operand.String()), left.Pos())
	}
	if err = Unify(operand, r); err != nil {
		return nil, typeError(err, fmt.Sprintf("right hand of operator '%s' must be %s", op, operand.String()), right.Pos())
	}
	// Returns the same type as operands
	return operand, nil
}

func (env *Env) inferRelationalBinOp(op string, left, right ast.Expr) (Type, error) {
	l, err := env.infer(left)
	if err != nil {
		return nil, err
	}
	r, err := env.infer(right)
	if err != nil {
		return nil, err
	}
	if err = Unify(l, r); err != nil {
		return nil, typeError(err, fmt.Sprintf("type mismatch of operands at rational operator '%s'", op), left.Pos())
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
		if err := env.checkNodeType("operand of operator 'not'", n.Child, BoolType); err != nil {
			return nil, err
		}
		return BoolType, nil
	case *ast.Neg:
		if err := env.checkNodeType("operand of unary operator '-'", n.Child, IntType); err != nil {
			return nil, err
		}
		return IntType, nil
	case *ast.Add:
		return env.inferArithmeticBinOp("+", n.Left, n.Right, IntType)
	case *ast.Sub:
		return env.inferArithmeticBinOp("-", n.Left, n.Right, IntType)
	case *ast.FNeg:
		if err := env.checkNodeType("operand of unary operator '-.'", n.Child, FloatType); err != nil {
			return nil, err
		}
		return FloatType, nil
	case *ast.FAdd:
		return env.inferArithmeticBinOp("+.", n.Left, n.Right, FloatType)
	case *ast.FSub:
		return env.inferArithmeticBinOp("-.", n.Left, n.Right, FloatType)
	case *ast.FMul:
		return env.inferArithmeticBinOp("*.", n.Left, n.Right, FloatType)
	case *ast.FDiv:
		return env.inferArithmeticBinOp("/.", n.Left, n.Right, FloatType)
	case *ast.Eq:
		return env.inferRelationalBinOp("=", n.Left, n.Right)
	case *ast.Less:
		return env.inferRelationalBinOp("<", n.Left, n.Right)
	case *ast.LessEq:
		return env.inferRelationalBinOp("<=", n.Left, n.Right)
	case *ast.If:
		if err := env.checkNodeType("condition of 'if' expression", n.Cond, BoolType); err != nil {
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
			return nil, typeError(err, "mismatch of types for 'then' clause and 'else' clause in 'if' expression", n.Pos())
		}

		return t, nil
	case *ast.Let:
		bound, err := env.infer(n.Bound)
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(n.Symbol.DisplayName, "$unused") {
			// Parser expands `foo; bar` to `let $unused = foo in bar`. In this situation,
			// type of the variable will never be determined because it's unused.
			// So skipping it in order to avoid unknown type error for the unused variable.
			return env.infer(n.Body)
		}

		t := &Var{}
		if err = Unify(t, bound); err != nil {
			return nil, typeError(err, fmt.Sprintf("type of variable '%s'", n.Symbol.DisplayName), n.Body.Pos())
		}

		env.Table[n.Symbol.Name] = bound
		return env.infer(n.Body)
	case *ast.VarRef:
		if t, ok := env.Table[n.Symbol.Name]; ok {
			return t, nil
		}
		if t, ok := env.Externals[n.Symbol.Name]; ok {
			return t, nil
		}
		// Assume as free variable. If free variable's type is not identified,
		// It falls into compilation error
		t := &Var{}
		env.Externals[n.Symbol.DisplayName] = t
		return t, nil
	case *ast.LetRec:
		f := &Var{}
		// Need to register function here because of recursive functions
		env.Table[n.Func.Symbol.Name] = f

		// Register parameters of function as variables to table
		params := make([]Type, len(n.Func.Params))
		for i, p := range n.Func.Params {
			// Types of parameters are unknown at definition
			t := &Var{}
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
			return nil, typeError(err, fmt.Sprintf("function '%s'", n.Func.Symbol.DisplayName), n.Pos())
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
		ret := &Var{}
		fun := &Fun{
			Ret:    ret,
			Params: args,
		}

		callee, err := env.infer(n.Callee)
		if err != nil {
			return nil, err
		}

		if err = Unify(callee, fun); err != nil {
			return nil, typeError(err, "type of called function", n.Pos())
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
			t := &Var{}
			env.Table[sym.Name] = t
			elems[i] = t
		}

		// Bound value must be tuple
		if err := env.checkNodeType("bound tuple value at 'let'", n.Bound, &Tuple{Elems: elems}); err != nil {
			return nil, err
		}

		return env.infer(n.Body)
	case *ast.ArrayCreate:
		if err := env.checkNodeType("size at array creation", n.Size, IntType); err != nil {
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
		elem := &Var{}
		array := &Array{Elem: elem}

		if err := env.checkNodeType("array value in index access", n.Array, array); err != nil {
			return nil, err
		}

		if err := env.checkNodeType("index access to array", n.Index, IntType); err != nil {
			return nil, err
		}

		return elem, nil
	case *ast.Put:
		if err := env.checkNodeType("index at assignment to an element of array", n.Index, IntType); err != nil {
			return nil, err
		}
		assignee, err := env.infer(n.Assignee)
		if err != nil {
			return nil, err
		}

		// Type of assigned value must be the same as element type of the array
		array := &Array{Elem: assignee}
		if err := env.checkNodeType("assignment to an element of array", n.Array, array); err != nil {
			return nil, err
		}

		// Assign to array does not have a value, so return unit type
		return UnitType, nil
	}
	panic(fmt.Sprintf("Unreachable: %s %v %v", e.Name(), e.Pos(), e.End()))
}
