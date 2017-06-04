package lexer

import (
	"fmt"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/locerr"
	"path/filepath"
)

func Example() {
	file := filepath.FromSlash("../testdata/from-mincaml/ack.ml")
	src, err := locerr.NewSourceFromFile(file)
	if err != nil {
		// File not found
		panic(err)
	}

	// Create lexer instance for the source
	lex := NewLexer(src)

	// Start to lex the source in other goroutine
	go lex.Lex()

	// tokens will be sent from lex.Tokens channel
	for {
		select {
		case tok := <-lex.Tokens:
			switch tok.Kind {
			case token.ILLEGAL:
				fmt.Printf("Lexing invalid token at %v\n", tok.Start)
				return
			case token.EOF:
				fmt.Println("End of input")
				return
			default:
				fmt.Printf("Token: %s", tok.String())
			}
		}
	}
}
