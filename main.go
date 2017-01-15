package main

import (
	"fmt"
	"github.com/rhysd/mincaml-parser/lexer"
	"github.com/rhysd/mincaml-parser/token"
	"io"
	"os"
)

func getSource(args []string) (io.Reader, string, error) {
	if len(args) <= 1 {
		return os.Stdin, "<stdin>", nil
	} else {
		n := args[0]
		f, err := os.Open(n)
		if err != nil {
			return nil, n, err
		}
		return f, n, nil
	}
}

func main() {
	r, f, err := getSource(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error on opening %s: %s\n", f, err.Error())
		os.Exit(4)
	}

	ch := make(chan token.Token)
	l := lexer.NewLexer(f, r, ch)
	go l.Lex()
	for {
		select {
		case t := <-ch:
			switch t.Kind {
			case token.EOF:
				os.Exit(0)
			case token.ILLEGAL:
				fmt.Fprintf(os.Stderr, "Error at %s", t.String())
				os.Exit(5)
			default:
				fmt.Println(t.String())
			}
		}
	}
}
