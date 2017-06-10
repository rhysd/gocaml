package driver

import (
	"fmt"
	"github.com/rhysd/locerr"
	"path/filepath"
)

func Example() {
	// Compile testdata/from-mincaml/ack.ml
	file := filepath.FromSlash("../testdata/from-mincaml/ack.ml")
	src, err := locerr.NewSourceFromFile(file)
	if err != nil {
		// File not found
		panic(err)
	}

	d := Driver{}

	// Show list of tokens
	d.PrintTokens(src)

	// Show AST nodes
	d.PrintAST(src)

	// Parse file into AST
	ast, err := d.Parse(src)
	if err != nil {
		panic(err)
	}

	// Resolving symbols, type analysis and converting AST into MIR instruction block
	env, err := d.SemanticAnalysis(ast)
	if err != nil {
		panic(err)
	}

	// Show environment of type analysis
	env.Dump()

	// Show LLVM IR for the source
	ir, err := d.EmitLLVMIR(src)
	if err != nil {
		panic(err)
	}
	fmt.Println(ir)

	// Show native assembly code for the source
	asm, err := d.EmitAsm(src)
	if err != nil {
		panic(err)
	}
	fmt.Println(asm)

	// Compile the source into an executable
	if err := d.Compile(src); err != nil {
		panic(err)
	}
}
