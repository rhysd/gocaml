package typing

import (
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
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

	lex := lexer.NewLexer(src)
	go lex.Lex()

	ast, err := parser.Parse(lex.Tokens)
	if err != nil {
		// When parse failed
		panic(err)
	}

	if err = alpha.Transform(ast.Root); err != nil {
		// When some some duplicates found
		panic(err)
	}

	// Apply type inference. After this, all symbols in AST should have exact
	// types. It also checks types are valid and all types are determined by
	// inference. It returns a type environment object as the result.
	env, err := TypeCheck(ast)
	if err != nil {
		// Type error detected
		panic(err)
	}

	// You can dump the type table
	env.Dump()
}
