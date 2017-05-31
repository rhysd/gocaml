package typing

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
)

var Primitives = map[string]Type{
	"unit":   UnitType,
	"int":    IntType,
	"bool":   BoolType,
	"float":  FloatType,
	"string": StringType,
}

func nodesToTypes(nodes []ast.Expr) ([]Type, error) {
	types := make([]Type, 0, len(nodes))
	for _, n := range nodes {
		t, err := nodeToType(n)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, nil
}

func nodeToType(node ast.Expr) (Type, error) {
	switch n := node.(type) {
	case *ast.FuncType:
		params, err := nodesToTypes(n.ParamTypes)
		if err != nil {
			return nil, err
		}

		ret, err := nodeToType(n.RetType)
		if err != nil {
			return nil, err
		}

		return &Fun{ret, params}, nil
	case *ast.TupleType:
		elems, err := nodesToTypes(n.ElemTypes)
		return &Tuple{elems}, err
	case *ast.CtorType:
		len := len(n.ParamTypes)
		if len == 0 {
			if n.Ctor == "_" {
				// '_' accepts any type.
				return &Var{}, nil
			}
			if prim, ok := Primitives[n.Ctor]; ok {
				return prim, nil
			}
		}

		// TODO: Currently only built-in array and option types are supported
		switch n.Ctor {
		case "array":
			if len != 1 {
				p := n.Pos()
				return nil, fmt.Errorf("Invalid array type at (line:%d,column:%d). 'array' only has 1 type parameter.", p.Line, p.Column)
			}
			elem, err := nodeToType(n.ParamTypes[0])
			return &Array{elem}, err
		case "option":
			if len != 1 {
				p := n.Pos()
				return nil, fmt.Errorf("Invalid option type at (line:%d,column:%d). 'option' only has 1 type parameter.", p.Line, p.Column)
			}
			elem, err := nodeToType(n.ParamTypes[0])
			return &Option{elem}, err
		default:
			p := n.Pos()
			return nil, fmt.Errorf("Unknown type constructor '%s' at (line:%d,column:%d). Currently only primitive types, 'array', 'option' and '_' are supported as built-in.", n.Ctor, p.Line, p.Column)
		}
	default:
		panic("FATAL: Cannot convert non-type AST node into type values: " + node.Name())
	}
}
