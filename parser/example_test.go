package parser

import (
	"fmt"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/token"
	"path/filepath"
)

func Example() {
	file := filepath.FromSlash("../testdata/from-mincaml/ack.ml")
	src, err := token.NewSourceFromFile(file)
	if err != nil {
		// File not found
		panic(err)
	}

	lex := lexer.NewLexer(src)
	go lex.Lex()

	// Parse() takes channel of token which is usually given from lexer
	// And returns the root of AST.
	r, err := Parse(lex.Tokens)
	if err != nil {
		// When parse failed
		panic(err)
	}
	fmt.Printf("AST: %v\n", r)
}
