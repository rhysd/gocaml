package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/types"
)

type variantResolver struct {
	env  *types.Env
	poly *polyEnv
}

func newVariantResolver(env *types.Env) *variantResolver {
	return &variantResolver{env, nil}
}

func (res *variantResolver) VisitTopdown(node ast.Expr) ast.Visitor {
	switch node := node.(type) {
	case *ast.Let:
		origin := res.poly
		res.poly = res.poly.nest()
		ast.Visit(res, node.Body)
		bodyPoly := res.poly
		res.poly = origin
		res.poly = res.poly.nest()
		ast.Visit(res, node.Bound)
		res.poly = res.poly.merge(bodyPoly, origin)
		fmt.Println(node.Symbol.Name, "at", node.Pos().String(), "\n", res.poly.flatten().String())
		return nil
	case *ast.LetRec:
		origin := res.poly
		res.poly = res.poly.nest()
		ast.Visit(res, node.Body)
		bodyPoly := res.poly
		res.poly = origin
		res.poly = res.poly.nest()
		ast.Visit(res, node.Func.Body)
		res.poly = res.poly.merge(bodyPoly, origin)
		fmt.Println(node.Func.Symbol.Name, "at", node.Pos().String(), "\n", res.poly.flatten().String())
		return nil
	case *ast.LetTuple:
		origin := res.poly
		res.poly = res.poly.nest()
		ast.Visit(res, node.Body)
		bodyPoly := res.poly
		res.poly = origin
		res.poly = res.poly.nest()
		ast.Visit(res, node.Bound)
		res.poly = res.poly.merge(bodyPoly, origin)
		fmt.Println("LetTuple at", node.Pos().String(), "\n", res.poly.flatten().String())
		return nil
	case *ast.VarRef:
		if inst, ok := res.env.RefInsts[node]; ok {
			for _, m := range inst.Mapping {
				res.poly.addPoly(m.ID, m.Type)
			}
		}
	}
	return res
}

func (res *variantResolver) VisitBottomup(node ast.Expr) {
	// switch node := node.(type) {
	// case *ast.Let, *ast.LetRec, *ast.LetTuple:
	// 	fmt.Println(node.Name(), "at", node.Pos().String(), "\n", res.poly.String())
	// }
}

func tmpResolvePolyVariant(root ast.Expr, env *types.Env) {
	env.DumpDebug()
	fmt.Println()
	res := newVariantResolver(env)
	res.poly = res.poly.nest()
	ast.Visit(res, root)
	fmt.Println("Poly Variant:", res.poly.flatten())
}
