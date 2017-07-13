package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/types"
)

type tmpPoly struct {
	ctx     *polyVariantsContext
	env     *types.Env
	schemes schemes
	assign  map[types.VarID]types.Type
}

func (poly *tmpPoly) VisitTopdown(node ast.Expr) ast.Visitor {
	switch node := node.(type) {
	case *ast.Let:
		origin := poly.ctx
		poly.ctx = poly.ctx.nest()
		ast.Visit(poly, node.Body)
		bodyCtx := poly.ctx
		poly.ctx = origin
		poly.ctx = poly.ctx.nest()
		ast.Visit(poly, node.Bound)
		poly.ctx = origin.merge(bodyCtx, poly.ctx)
		fmt.Println(node.Symbol.Name, "at", node.Pos().String())
		fmt.Println(poly.ctx.allVariants().String())
		fmt.Println()
		return nil
	case *ast.LetRec:
		origin := poly.ctx
		poly.ctx = poly.ctx.nest()
		ast.Visit(poly, node.Body)
		bodyCtx := poly.ctx
		poly.ctx = origin
		poly.ctx = poly.ctx.nest()
		ast.Visit(poly, node.Func.Body)
		poly.ctx = origin.merge(bodyCtx, poly.ctx)
		fmt.Println(node.Func.Symbol.Name, "at", node.Pos().String())
		fmt.Println(poly.ctx.allVariants().String())
		fmt.Println()
		return nil
	case *ast.LetTuple:
		origin := poly.ctx
		poly.ctx = poly.ctx.nest()
		ast.Visit(poly, node.Body)
		bodyCtx := poly.ctx
		poly.ctx = origin
		poly.ctx = poly.ctx.nest()
		ast.Visit(poly, node.Bound)
		poly.ctx = origin.merge(bodyCtx, poly.ctx)
		fmt.Println("LetTuple at", node.Pos().String())
		fmt.Println(poly.ctx.allVariants().String())
		fmt.Println()
		return nil
	case *ast.VarRef:
		if inst, ok := poly.env.RefInsts[node]; ok {
			poly.ctx.add(inst)
		}
	}
	return poly
}

func (poly *tmpPoly) VisitBottomup(node ast.Expr) {
	// TODO?
}

func tmpPolyVisit(root ast.Expr, env *types.Env, schemes schemes) {
	vis := &tmpPoly{nil, env, schemes, map[types.VarID]types.Type{}}
	ast.Visit(vis, root)
	fmt.Println("Result of tmpPolyVisit:\n")
	fmt.Println(vis.ctx.allVariants().String())
}
