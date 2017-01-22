package main

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"os"
)

func getSource(args []string) (*token.Source, error) {
	if len(args) <= 1 {
		return token.NewSourceFromStdin()
	} else {
		return token.NewSourceFromFile(args[1])
	}
}

func main() {
	src, err := getSource(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error on opening source: %s\n", err.Error())
		os.Exit(4)
	}

	ch := make(chan token.Token)
	l := lexer.NewLexer(src, ch)
	go l.Lex()

	root, err := parser.Parse(ch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(5)
	}

	a := ast.AST{
		root,
		src,
	}

	ast.Print(a)
}
