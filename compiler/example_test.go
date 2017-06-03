package compiler

import (
	"github.com/rhysd/loc"
	"path/filepath"
)

func Example() {
	// Compile testdata/from-mincaml/ack.ml
	file := filepath.FromSlash("../testdata/from-mincaml/ack.ml")
	src, err := loc.NewSourceFromFile(file)
	if err != nil {
		// File not found
		panic(err)
	}

	c := Compiler{}

	// Show list of tokens
	c.PrintTokens(src)

	// Show AST nodes
	c.PrintAST(src)

	// Parse file into AST
	ast, err := c.Parse(src)
	if err != nil {
		panic(err)
	}

	// Do semantic analysis (type check and inference)
	env, err := c.SemanticAnalysis(ast)
	if err != nil {
		panic(err)
	}

	// Show environment of type analysis
	env.Dump()

	// TODO: LLVM IR code generation
}
