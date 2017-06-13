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
	parsed, err := d.Parse(src)
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed)

	// Resolving symbols and type analysis
	env, inferred, err := d.SemanticAnalysis(src)
	if err != nil {
		panic(err)
	}

	// Show environment of type analysis
	env.Dump()

	// Show inferred types of all AST nodes
	for e, t := range inferred {
		fmt.Printf("Node '%s' at '%s' => Type '%s'", e.Name(), e.Pos(), t.String())
	}

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
