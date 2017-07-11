package ast

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Printer is a visitor to print AST to io.Writer
type Printer struct {
	indent int
	out    io.Writer
}

func (p Printer) VisitTopdown(e Expr) Visitor {
	fmt.Fprintf(p.out, "\n%s%s (%d:%d-%d:%d)", strings.Repeat("-   ", p.indent), e.Name(), e.Pos().Line, e.Pos().Column, e.End().Line, e.End().Column)
	return Printer{p.indent + 1, p.out}
}

func (p Printer) VisitBottomup(Expr) {
	return
}

// Fprint outputs a structure of AST to given io.Writer object
func Fprint(out io.Writer, a *AST) {
	fmt.Fprintf(out, "AST for %s:", a.File().Path)
	p := Printer{1, out}
	for _, t := range a.TypeDecls {
		Visit(p, t)
	}
	for _, e := range a.Externals {
		Visit(p, e)
	}
	Visit(p, a.Root)
}

// Print outputs a structure of AST to stdout.
func Print(a *AST) {
	Fprint(os.Stdout, a)
}

// Println does the same as Print and append newline at the end of output.
func Println(a *AST) {
	Print(a)
	fmt.Println()
}
