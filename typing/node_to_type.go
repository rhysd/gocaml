package typing

import (
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/locerr"
)

type nodeTypeConv struct {
	aliases map[string]Type
}

func newNodeTypeConv(decls []*ast.TypeDecl) (*nodeTypeConv, error) {
	conv := &nodeTypeConv{make(map[string]Type, len(decls)+5 /*primitives*/)}
	conv.aliases["unit"] = UnitType
	conv.aliases["int"] = IntType
	conv.aliases["bool"] = BoolType
	conv.aliases["float"] = FloatType
	conv.aliases["string"] = StringType

	for _, decl := range decls {
		if decl.Ident == "_" {
			return nil, locerr.ErrorAt(decl.Pos(), "Cannot declare '_' type name")
		}
		if t, ok := conv.aliases[decl.Ident]; ok {
			return nil, locerr.ErrorfAt(decl.Pos(), "Type name '%s' was already declared as type '%s' at (line:%d, column:%d)", decl.Ident, t.String())
		}
		t, err := conv.nodeToType(decl.Type)
		if err != nil {
			return nil, locerr.NotefAt(decl.Pos(), err, "Type declaration '%s'", decl.Ident)
		}
		conv.aliases[decl.Ident] = t
	}
	return conv, nil
}

func (conv *nodeTypeConv) nodesToTypes(nodes []ast.Expr) ([]Type, error) {
	types := make([]Type, 0, len(nodes))
	for _, n := range nodes {
		t, err := conv.nodeToType(n)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, nil
}

func (conv *nodeTypeConv) nodeToType(node ast.Expr) (Type, error) {
	switch n := node.(type) {
	case *ast.FuncType:
		params, err := conv.nodesToTypes(n.ParamTypes)
		if err != nil {
			return nil, err
		}

		ret, err := conv.nodeToType(n.RetType)
		if err != nil {
			return nil, err
		}

		return &Fun{ret, params}, nil
	case *ast.TupleType:
		elems, err := conv.nodesToTypes(n.ElemTypes)
		return &Tuple{elems}, err
	case *ast.CtorType:
		len := len(n.ParamTypes)
		if len == 0 {
			if n.Ctor == "_" {
				// '_' accepts any type.
				return &Var{}, nil
			}
			if t, ok := conv.aliases[n.Ctor]; ok {
				// TODO: Currently
				return t, nil
			}
		}

		// TODO: Currently only built-in array and option types are supported
		switch n.Ctor {
		case "array":
			if len != 1 {
				return nil, locerr.ErrorAt(n.Pos(), "Invalid array type. 'array' only has 1 type parameter.")
			}
			elem, err := conv.nodeToType(n.ParamTypes[0])
			return &Array{elem}, err
		case "option":
			if len != 1 {
				return nil, locerr.ErrorAt(n.Pos(), "Invalid option type. 'option' only has 1 type parameter.")
			}
			elem, err := conv.nodeToType(n.ParamTypes[0])
			return &Option{elem}, err
		default:
			return nil, locerr.ErrorfAt(n.Pos(), "Unknown type constructor '%s'. Primitive types, aliased types, 'array', 'option' and '_' are supported", n.Ctor)
		}
	default:
		panic("FATAL: Cannot convert non-type AST node into type values: " + node.Name())
	}
}
