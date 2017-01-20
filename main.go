package main

import (
	"fmt"
	"github.com/rhysd/mincaml-parser/lexer"
	"github.com/rhysd/mincaml-parser/source"
	"github.com/rhysd/mincaml-parser/token"
	"os"
)

func getSource(args []string) (*source.Source, error) {
	if len(args) <= 1 {
		return source.FromStdin()
	} else {
		return source.FromFile(args[1])
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
	for {
		select {
		case t := <-ch:
			switch t.Kind {
			case token.EOF:
				os.Exit(0)
			case token.ILLEGAL:
				os.Exit(5)
			default:
				fmt.Println(t.String())
			}
		}
	}
}
