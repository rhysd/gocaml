package ast

import (
	"fmt"
	"strings"
)

type Printer struct {
	indent int
}

func Print(a *AST) {
	fmt.Printf("AST for %s:", a.File.Name)
	printExpr(a.Root, 1)
}

func printExpr(e Expr, indent int) {
	fmt.Printf("\n%s<%s:%d:%d-%d:%d>", strings.Repeat("  ", indent), e.Pos().Line, e.Pos().Column, e.End().Line, e.End().Column)

	i := indent + 1

	switch n := e.(type) {
	case *Not:
		printExpr(n.Child, i)
	case *Neg:
		printExpr(n.Child, i)
	case *Add:
		printExpr(n.Left, i)
		printExpr(n.Right, i)
	case *Sub:
		printExpr(n.Left, i)
		printExpr(n.Right, i)
	case *FNeg:
		printExpr(n.Child, i)
	case *FAdd:
		printExpr(n.Left, i)
		printExpr(n.Right, i)
	case *FSub:
		printExpr(n.Left, i)
		printExpr(n.Right, i)
	case *FMul:
		printExpr(n.Left, i)
		printExpr(n.Right, i)
	case *FDiv:
		printExpr(n.Left, i)
		printExpr(n.Right, i)
	case *Eq:
		printExpr(n.Left, i)
		printExpr(n.Right, i)
	case *Less:
		printExpr(n.Left, i)
		printExpr(n.Right, i)
	case *If:
		printExpr(n.Cond, i)
		printExpr(n.Then, i)
		printExpr(n.Else, i)
	case *Let:
		printExpr(n.Bound, i)
		printExpr(n.Body, i)
	case *LetRec:
		printExpr(n.Func.Body, i)
		printExpr(n.Body, i)
	case *Apply:
		printExpr(n.Callee, i)
		for _, e := range n.Args {
			printExpr(e, i)
		}
	case *Tuple:
		for _, e := range n.Elems {
			printExpr(e, i)
		}
	case *LetTuple:
		printExpr(n.Bound, i)
		printExpr(n.Body, i)
	case *Array:
		printExpr(n.Size, i)
		printExpr(n.Elem, i)
	case *Get:
		printExpr(n.Array, i)
		printExpr(n.Index, i)
	case *Put:
		printExpr(n.Array, i)
		printExpr(n.Index, i)
		printExpr(n.Assignee, i)
	}
}
