package ast

import (
	"fmt"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/locerr"
)

// Visitor which counts number of nodes in AST
type printPath struct {
	total int
}

// VisitTopdown method is called before children are visited
func (v *printPath) VisitTopdown(e Expr) Visitor {
	fmt.Printf("\n -> %s (topdown)", e.Name())
	return v
}

// VisitBottomup method is called after children were visited
func (v *printPath) VisitBottomup(e Expr) {
	fmt.Printf("\n -> %s (bottomup)", e.Name())
}

func Example() {
	src := locerr.NewDummySource("")

	// AST which usually comes from syntax.Parse() function.
	rootOfAST := &Let{
		LetToken: &token.Token{File: src},
		Symbol:   NewSymbol("test"),
		Bound: &Int{
			Token: &token.Token{File: src},
			Value: 42,
		},
		Body: &Add{
			Left: &VarRef{
				Token:  &token.Token{File: src},
				Symbol: NewSymbol("test"),
			},
			Right: &Float{
				Token: &token.Token{File: src},
				Value: 3.14,
			},
		},
	}

	ast := &AST{Root: rootOfAST}

	// Apply visitor to root node of AST
	v := &printPath{0}
	fmt.Println("ROOT")

	Visit(v, ast.Root)
	// Output:
	// ROOT
	//  -> Let (test) (topdown)
	//  -> Int (topdown)
	//  -> Int (bottomup)
	//  -> Add (topdown)
	//  -> VarRef (test) (topdown)
	//  -> VarRef (test) (bottomup)
	//  -> Float (topdown)
	//  -> Float (bottomup)
	//  -> Add (bottomup)
	//  -> Let (test) (bottomup)

	// Print AST
	Println(ast)
}
