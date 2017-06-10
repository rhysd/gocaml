package syntax

import (
	"fmt"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/locerr"
	"path/filepath"
)

func ExampleLexer_Lex() {
	file := filepath.FromSlash("../testdata/from-mincaml/ack.ml")
	src, err := locerr.NewSourceFromFile(file)
	if err != nil {
		// File not found
		panic(err)
	}

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

func ExampleParse() {
	file := filepath.FromSlash("../testdata/from-mincaml/ack.ml")
	src, err := locerr.NewSourceFromFile(file)
	if err != nil {
		// File not found
		panic(err)
	}

	// Create lexer instance for the source
	lex := NewLexer(src)
	go lex.Lex()

	// ParseTokens() takes channel of token which is usually given from lexer
	// And returns the root of AST.
	tree, err := ParseTokens(lex.Tokens)
	if err != nil {
		// When parse failed
		panic(err)
	}

	fmt.Printf("AST: %v\n", tree)

	// If you want to parse a source code into AST directly, simply call Parse() function.
	tree, err = Parse(src)
	if err != nil {
		// When lexing or parsing failed
		panic(err)
	}

	fmt.Printf("AST: %v\n", tree)
}
