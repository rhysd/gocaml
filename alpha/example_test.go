package alpha

import (
	"fmt"
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

	// Run alpha transform against the root of AST
	if err = Transform(root); err != nil {
		// When some some duplicates found
		panic(err)
	}

	// Now all symbols in the AST have unique names
	// e.g. abc -> abc$t1
	// And now all variable references (VarRef) point a symbol instance of the definition node.
	// By checking the pointer of symbol, we can know where the variable reference are defined
	// in source.
	fmt.Printf("%v\n", root)
}
