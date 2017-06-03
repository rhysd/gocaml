package typing

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/ast"
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
			pos := decl.Pos()
			return nil, errors.Errorf("Cannot declare '_' type name at (line:%d, column:%d)", pos.Line, pos.Column)
		}
		if t, ok := conv.aliases[decl.Ident]; ok {
			pos := decl.Pos()
			return nil, errors.Errorf("Type name '%s' was already declared as type '%s' at (line:%d, column:%d)", decl.Ident, t.String(), pos.Line, pos.Column)
		}
		t, err := conv.nodeToType(decl.Type)
		if err != nil {
			return nil, typeError(err, fmt.Sprintf("Type declaration '%s'", decl.Ident), decl.Pos())
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
				p := n.Pos()
				return nil, errors.Errorf("Invalid array type at (line:%d,column:%d). 'array' only has 1 type parameter.", p.Line, p.Column)
			}
			elem, err := conv.nodeToType(n.ParamTypes[0])
			return &Array{elem}, err
		case "option":
			if len != 1 {
				p := n.Pos()
				return nil, errors.Errorf("Invalid option type at (line:%d,column:%d). 'option' only has 1 type parameter.", p.Line, p.Column)
			}
			elem, err := conv.nodeToType(n.ParamTypes[0])
			return &Option{elem}, err
		default:
			p := n.Pos()
			return nil, errors.Errorf("Unknown type constructor '%s' at (line:%d,column:%d). Primitive types, aliased types, 'array', 'option' and '_' are supported", n.Ctor, p.Line, p.Column)
		}
	default:
		panic("FATAL: Cannot convert non-type AST node into type values: " + node.Name())
	}
}
