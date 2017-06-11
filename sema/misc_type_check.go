package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
)

func checkLess(op string, opTy types.Type) string {
	switch opTy.(type) {
	case *types.Unit, *types.Bool, *types.String, *types.Fun, *types.Tuple, *types.Array, *types.Option:
		return fmt.Sprintf("'%s' can't be compared with operator '%s'", opTy.String(), op)
	default:
		return ""
	}
}

func checkEq(op string, opTy types.Type) string {
	if a, ok := opTy.(*types.Array); ok {
		return fmt.Sprintf("Array type '%s' can't be compared with operator '%s'", a.String(), op)
	}
	return ""
}

func typeAt(n ast.Expr, ts exprTypes) types.Type {
	t, ok := ts[n]
	if !ok {
		panic("FATAL: Unknown type at " + n.Pos().String())
	}
	return t
}

func checkNode(n ast.Expr, t types.Type, ts exprTypes) string {
	switch n := n.(type) {
	case *ast.Less:
		return checkLess("<", typeAt(n.Left, ts))
	case *ast.LessEq:
		return checkLess("<=", typeAt(n.Left, ts))
	case *ast.Greater:
		return checkLess(">", typeAt(n.Left, ts))
	case *ast.GreaterEq:
		return checkLess(">=", typeAt(n.Left, ts))
	case *ast.Eq:
		return checkEq("=", typeAt(n.Left, ts))
	case *ast.NotEq:
		return checkEq("<>", typeAt(n.Left, ts))
	default:
		return ""
	}
}

func MiscTypeCheck(ts exprTypes) (err *locerr.Error) {
	for e, t := range ts {
		if msg := checkNode(e, t, ts); msg != "" {
			if err == nil {
				err = locerr.ErrorIn(e.Pos(), e.End(), msg)
			} else {
				err = err.NoteAt(e.Pos(), msg)
			}
		}
	}
	return
}
