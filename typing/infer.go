package typing

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/common"
	"github.com/rhysd/gocaml/token"
)

func typeError(err error, where string, pos token.Position) error {
	return errors.Wrapf(err, "Type error: %s (line:%d, column:%d)\n", where, pos.Line, pos.Column)
}

// Inferer is a visitor to infer types in the AST
type Inferer struct {
	env  *Env
	conv *nodeTypeConv
}

// NewInferer creates a new Inferer instance
func NewInferer() *Inferer {
	return &Inferer{NewEnv(), nil}
}

func (inf *Inferer) checkNodeType(where string, node ast.Expr, expected Type) error {
	t, err := inf.infer(node)
	if err != nil {
		return err
	}
	if err = Unify(expected, t); err != nil {
		return typeError(err, fmt.Sprintf("%s must be '%s'", where, expected.String()), node.Pos())
	}
	return nil
}

func (inf *Inferer) inferArithmeticBinOp(op string, left, right ast.Expr, operand Type) (Type, error) {
	l, err := inf.infer(left)
	if err != nil {
		return nil, err
	}
	r, err := inf.infer(right)
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

func (inf *Inferer) inferRelationalBinOp(op string, left, right ast.Expr) (Type, error) {
	l, err := inf.infer(left)
	if err != nil {
		return nil, err
	}
	r, err := inf.infer(right)
	if err != nil {
		return nil, err
	}
	if err = Unify(l, r); err != nil {
		return nil, typeError(err, fmt.Sprintf("type mismatch at operands of relational operator '%s'", op), left.Pos())
	}
	return BoolType, nil
}

func (inf *Inferer) inferLogicalOp(op string, left, right ast.Expr) (Type, error) {
	for i, e := range []ast.Expr{left, right} {
		t, err := inf.infer(e)
		if err != nil {
			return nil, err
		}
		if err = Unify(BoolType, t); err != nil {
			return nil, typeError(err, fmt.Sprintf("type mismatch at %dth operand of logical operator '%s'", i+1, op), e.Pos())
		}
	}
	return BoolType, nil
}

func (inf *Inferer) infer(e ast.Expr) (Type, error) {
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
		if err := inf.checkNodeType("operand of operator 'not'", n.Child, BoolType); err != nil {
			return nil, err
		}
		return BoolType, nil
	case *ast.Neg:
		if err := inf.checkNodeType("operand of unary operator '-'", n.Child, IntType); err != nil {
			return nil, err
		}
		return IntType, nil
	case *ast.Add:
		return inf.inferArithmeticBinOp("+", n.Left, n.Right, IntType)
	case *ast.Sub:
		return inf.inferArithmeticBinOp("-", n.Left, n.Right, IntType)
	case *ast.Mul:
		return inf.inferArithmeticBinOp("*", n.Left, n.Right, IntType)
	case *ast.Div:
		return inf.inferArithmeticBinOp("/", n.Left, n.Right, IntType)
	case *ast.Mod:
		return inf.inferArithmeticBinOp("%", n.Left, n.Right, IntType)
	case *ast.FNeg:
		if err := inf.checkNodeType("operand of unary operator '-.'", n.Child, FloatType); err != nil {
			return nil, err
		}
		return FloatType, nil
	case *ast.FAdd:
		return inf.inferArithmeticBinOp("+.", n.Left, n.Right, FloatType)
	case *ast.FSub:
		return inf.inferArithmeticBinOp("-.", n.Left, n.Right, FloatType)
	case *ast.FMul:
		return inf.inferArithmeticBinOp("*.", n.Left, n.Right, FloatType)
	case *ast.FDiv:
		return inf.inferArithmeticBinOp("/.", n.Left, n.Right, FloatType)
	case *ast.Eq:
		return inf.inferRelationalBinOp("=", n.Left, n.Right)
	case *ast.NotEq:
		return inf.inferRelationalBinOp("<>", n.Left, n.Right)
	case *ast.Less:
		return inf.inferRelationalBinOp("<", n.Left, n.Right)
	case *ast.LessEq:
		return inf.inferRelationalBinOp("<=", n.Left, n.Right)
	case *ast.Greater:
		return inf.inferRelationalBinOp(">", n.Left, n.Right)
	case *ast.GreaterEq:
		return inf.inferRelationalBinOp(">=", n.Left, n.Right)
	case *ast.And:
		return inf.inferLogicalOp("&&", n.Left, n.Right)
	case *ast.Or:
		return inf.inferLogicalOp("||", n.Left, n.Right)
	case *ast.If:
		if err := inf.checkNodeType("condition of 'if' expression", n.Cond, BoolType); err != nil {
			return nil, err
		}

		t, err := inf.infer(n.Then)
		if err != nil {
			return nil, err
		}

		e, err := inf.infer(n.Else)
		if err != nil {
			return nil, err
		}

		if err = Unify(t, e); err != nil {
			return nil, typeError(err, "mismatch of types for 'then' clause and 'else' clause in 'if' expression", n.Pos())
		}

		return t, nil
	case *ast.Let:
		bound, err := inf.infer(n.Bound)
		if err != nil {
			return nil, err
		}

		var t Type
		if n.Type != nil {
			// When let x: type = ...
			t, err = inf.conv.nodeToType(n.Type)
			if err != nil {
				return nil, err
			}
		} else {
			t = &Var{}
		}

		if err = Unify(t, bound); err != nil {
			return nil, typeError(err, fmt.Sprintf("type of variable '%s'", n.Symbol.DisplayName), n.Body.Pos())
		}

		inf.env.Table[n.Symbol.Name] = bound
		return inf.infer(n.Body)
	case *ast.VarRef:
		if t, ok := inf.env.Table[n.Symbol.Name]; ok {
			return t, nil
		}
		if t, ok := inf.env.Externals[n.Symbol.Name]; ok {
			return t, nil
		}
		// Assume as free variable. If free variable's type is not identified,
		// It falls into compilation error
		t := &Var{}
		inf.env.Externals[n.Symbol.DisplayName] = t
		return t, nil
	case *ast.LetRec:
		f := &Var{}
		// Need to register function here because of recursive functions
		inf.env.Table[n.Func.Symbol.Name] = f

		// Register parameters of function as variables to table
		params := make([]Type, len(n.Func.Params))
		for i, p := range n.Func.Params {
			var t Type
			var err error
			if p.Type != nil {
				t, err = inf.conv.nodeToType(p.Type)
				if err != nil {
					return nil, typeError(err, fmt.Sprintf("%s parameter of function", common.Ordinal(i+1)), p.Type.Pos())
				}
			} else {
				t = &Var{}
			}
			inf.env.Table[p.Ident.Name] = t
			params[i] = t
		}

		// Infer return type of function from its body
		ret, err := inf.infer(n.Func.Body)
		if err != nil {
			return nil, err
		}

		if n.Func.RetType != nil {
			t, err := inf.conv.nodeToType(n.Func.RetType)
			if err != nil {
				return nil, typeError(err, "return type of function", n.Func.RetType.Pos())
			}
			if err = Unify(t, ret); err != nil {
				return nil, typeError(err, "return type of function", n.Func.RetType.Pos())
			}
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

		return inf.infer(n.Body)
	case *ast.Apply:
		args := make([]Type, len(n.Args))
		for i, a := range n.Args {
			t, err := inf.infer(a)
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

		callee, err := inf.infer(n.Callee)
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
			t, err := inf.infer(e)
			if err != nil {
				return nil, err
			}
			elems[i] = t
		}
		return &Tuple{Elems: elems}, nil
	case *ast.LetTuple:
		var t Type

		if n.Type != nil {
			var err error
			t, err = inf.conv.nodeToType(n.Type)
			if err != nil {
				return nil, err
			}
			tpl, ok := t.(*Tuple)
			if !ok {
				p := n.Type.Pos()
				return nil, errors.Errorf("Type error: Bound value of 'let (...) =' must be tuple, but found '%s' (line:%d, column:%d)", t.String(), p.Line, p.Column)
			}
			if len(tpl.Elems) != len(n.Symbols) {
				p := n.Type.Pos()
				return nil, errors.Errorf("Type error: Mismatch numbers of elements of specified tuple type and symbols in 'let (...)' expression: %d vs %d (line:%d, column:%d)", len(tpl.Elems), len(n.Symbols), p.Line, p.Column)
			}
			for i, sym := range n.Symbols {
				inf.env.Table[sym.Name] = tpl.Elems[i]
			}
		} else {
			elems := make([]Type, len(n.Symbols))
			for i, sym := range n.Symbols {
				// Bound elements' types are unknown in this point
				v := &Var{}
				inf.env.Table[sym.Name] = v
				elems[i] = v
			}
			t = &Tuple{Elems: elems}
		}

		// Bound value must be tuple
		if err := inf.checkNodeType("bound tuple value at 'let'", n.Bound, t); err != nil {
			return nil, err
		}

		return inf.infer(n.Body)
	case *ast.ArrayCreate:
		if err := inf.checkNodeType("size at array creation", n.Size, IntType); err != nil {
			return nil, err
		}
		elem, err := inf.infer(n.Elem)
		if err != nil {
			return nil, err
		}
		return &Array{Elem: elem}, nil
	case *ast.ArraySize:
		if err := inf.checkNodeType("argument of 'Array.length'", n.Target, &Array{Elem: &Var{}}); err != nil {
			return nil, err
		}
		return IntType, nil
	case *ast.Get:
		// Lhs of Get must be array but its element type is unknown.
		// So introduce new type variable for it.
		elem := &Var{}
		array := &Array{Elem: elem}

		if err := inf.checkNodeType("array value in index access", n.Array, array); err != nil {
			return nil, err
		}

		if err := inf.checkNodeType("index access to array", n.Index, IntType); err != nil {
			return nil, err
		}

		return elem, nil
	case *ast.Put:
		if err := inf.checkNodeType("index at assignment to an element of array", n.Index, IntType); err != nil {
			return nil, err
		}
		assignee, err := inf.infer(n.Assignee)
		if err != nil {
			return nil, err
		}

		// Type of assigned value must be the same as element type of the array
		array := &Array{Elem: assignee}
		if err := inf.checkNodeType("assignment to an element of array", n.Array, array); err != nil {
			return nil, err
		}

		// Assign to array does not have a value, so return unit type
		return UnitType, nil
	case *ast.Some:
		elem, err := inf.infer(n.Child)
		if err != nil {
			return nil, err
		}
		return &Option{elem}, nil
	case *ast.None:
		t := &Option{&Var{}}
		inf.env.NoneTypes[n] = t
		return t, nil
	case *ast.Match:
		elem := &Var{}
		matched := &Option{elem}
		if err := inf.checkNodeType("matching target in 'match' expression", n.Target, matched); err != nil {
			return nil, err
		}

		inf.env.Table[n.SomeIdent.Name] = elem
		some, err := inf.infer(n.IfSome)
		if err != nil {
			return nil, err
		}
		none, err := inf.infer(n.IfNone)
		if err != nil {
			return nil, err
		}
		if err = Unify(some, none); err != nil {
			return nil, typeError(err, "mismatch of types between 'Some' arm and 'None' arm in 'match' expression", n.Pos())
		}
		return some, nil
	case *ast.Typed:
		child, err := inf.infer(n.Child)
		if err != nil {
			return nil, err
		}

		t, err := inf.conv.nodeToType(n.Type)
		if err != nil {
			return nil, err
		}

		if err = Unify(t, child); err != nil {
			return nil, typeError(err, "mismatch between inferred type and specified type", n.Pos())
		}

		return child, nil
	default:
		panic(fmt.Sprintf("FATAL: Unreachable: %s %v %v", e.Name(), e.Pos(), e.End()))
	}
}

// Infer infers types in given AST and returns error when detecting type errors
func (inferer *Inferer) Infer(parsed *ast.AST) error {
	var err error
	inferer.conv, err = newNodeTypeConv(parsed.TypeDecls)
	if err != nil {
		return err
	}

	root, err := inferer.infer(parsed.Root)
	if err != nil {
		return err
	}

	if err := Unify(UnitType, root); err != nil {
		return errors.Wrap(err, "Type of root expression of program must be unit\n")
	}

	// While dereferencing type variables in table, we can detect type variables
	// which does not have exact type and raise an error for that.
	// External variables must be well-typed also.
	return derefTypeVars(inferer.env, parsed.Root)
}

// TypeInferernce applies type inference and returns a type environment as the result
func TypeInferernce(parsed *ast.AST) (*Env, error) {
	inferer := NewInferer()
	if err := inferer.Infer(parsed); err != nil {
		return nil, err
	}
	return inferer.env, nil
}
