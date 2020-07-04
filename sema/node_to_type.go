package sema

import (
	"github.com/rhysd/gocaml/ast"
	. "github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
)

type nodeTypeConv struct {
	aliases        map[string]Type
	acceptsAnyType bool
}

func newNodeTypeConv(decls []*ast.TypeDecl) (*nodeTypeConv, error) {
	conv := &nodeTypeConv{make(map[string]Type, len(decls)+5 /*primitives*/), true}
	conv.aliases["unit"] = UnitType
	conv.aliases["int"] = IntType
	conv.aliases["bool"] = BoolType
	conv.aliases["float"] = FloatType
	conv.aliases["string"] = StringType

	for _, decl := range decls {
		t, err := conv.nodeToType(decl.Type, -1)
		if err != nil {
			return nil, locerr.NotefAt(decl.Pos(), err, "Type declaration '%s'", decl.Ident.Name)
		}
		conv.aliases[decl.Ident.Name] = t
	}
	return conv, nil
}

func (conv *nodeTypeConv) nodesToTypes(nodes []ast.Expr, level int) ([]Type, error) {
	types := make([]Type, 0, len(nodes))
	for _, n := range nodes {
		t, err := conv.nodeToType(n, level)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, nil
}

func (conv *nodeTypeConv) nodeToType(node ast.Expr, level int) (Type, error) {
	switch n := node.(type) {
	case *ast.FuncType:
		params, err := conv.nodesToTypes(n.ParamTypes, level)
		if err != nil {
			return nil, err
		}

		ret, err := conv.nodeToType(n.RetType, level)
		if err != nil {
			return nil, err
		}

		return &Fun{ret, params}, nil
	case *ast.TupleType:
		elems, err := conv.nodesToTypes(n.ElemTypes, level)
		return &Tuple{elems}, err
	case *ast.CtorType:
		len := len(n.ParamTypes)
		if len == 0 {
			if n.Ctor.Name == "_" {
				if !conv.acceptsAnyType {
					return nil, locerr.ErrorIn(n.Pos(), n.End(), "'_' is not permitted for type annotation in this context")
				}
				// '_' accepts any type.
				return &Var{Level: level}, nil
			}
			if t, ok := conv.aliases[n.Ctor.Name]; ok {
				return t, nil
			}
		}

		// TODO: Currently only built-in array and option types are supported
		switch n.Ctor.Name {
		case "array":
			if len != 1 {
				return nil, locerr.ErrorIn(n.Pos(), n.End(), "Invalid array type. 'array' only has 1 type parameter")
			}
			elem, err := conv.nodeToType(n.ParamTypes[0], level)
			return &Array{elem}, err
		case "option":
			if len != 1 {
				return nil, locerr.ErrorIn(n.Pos(), n.End(), "Invalid option type. 'option' only has 1 type parameter")
			}
			elem, err := conv.nodeToType(n.ParamTypes[0], level)
			return &Option{elem}, err
		default:
			return nil, locerr.ErrorfIn(n.Pos(), n.End(), "Unknown type constructor '%s'. Primitive types, aliased types, 'array', 'option' and '_' are supported", n.Ctor.DisplayName)
		}
	default:
		panic("FATAL: Cannot convert non-type AST node into type values: " + node.Name())
	}
}
