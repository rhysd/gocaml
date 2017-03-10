package ast

import (
	"fmt"
	"strings"
)

type Printer struct {
	indent int
}

func (p Printer) Visit(e Expr) Visitor {
	fmt.Printf("\n%s%s (%d:%d-%d:%d)", strings.Repeat("-   ", p.indent), e.Name(), e.Pos().Line, e.Pos().Column, e.End().Line, e.End().Column)
	return Printer{p.indent + 1}
}

// Print outputs a structure of AST to stdout.
func Print(a *AST) {
	fmt.Printf("AST for %s:", a.File.Name)
	// printExpr(a.Root, 1)
	p := Printer{1}
	Visit(p, a.Root)
}

// Println does the same as Print and append newline at the end of output.
func Println(a *AST) {
	Print(a)
	fmt.Println("")
}
