package typing

import (
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
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

	root, err := parser.Parse(lex.Tokens)
	if err != nil {
		// When parse failed
		panic(err)
	}

	// Create new type analysis environment
	// (symbol table and external variables table)
	env := NewEnv()

	// Apply type inference. After this, all symbols in AST should have exact
	// types. It also checks types are valid and all types are determined by
	// inference
	if err := env.ApplyTypeAnalysis(root); err != nil {
		// Type error detected
		panic(err)
	}
}
