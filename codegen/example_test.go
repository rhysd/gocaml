package codegen

import (
	"fmt"
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/closure"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
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
	if err = alpha.Transform(root); err != nil {
		// When some some duplicates found
		panic(err)
	}

	// Type analysis
	env := typing.NewEnv()
	if err := env.ApplyTypeAnalysis(root); err != nil {
		// Type error detected
		panic(err)
	}

	// Convert AST into GCIL instruction block
	block, err := gcil.FromAST(root, env)
	if err != nil {
		panic(err)
	}
	gcil.ElimRefs(block, env)

	// Create GCIL compilation unit
	program := closure.Transform(block)

	// Make options to emit the result
	options := EmitOptions{
		Optimization: OptimizeDefault,             // Optimization level
		Triple:       "x86_64-apple-darwin16.4.0", // Compilation target (Empty string means default target on your machine)
		DebugInfo:    true,                        // Add debug information to the result or not
	}

	// Emitter object, which compiles GCIL to LLVM IR and emits assembly, object file or executable
	// In factory function, given GCIL code is already converted to LLVM IR
	emitter, err := NewEmitter(program, env, src, options)
	if err != nil {
		panic(err)
	}

	// You need to defer finalization
	defer emitter.Dispose()

	// Run LLVM IR level optimizations
	emitter.RunOptimizationPasses()

	// Show LLVM IR compiled from `program`
	fmt.Println("LLVMIR:\n" + emitter.EmitLLVMIR())

	// Emit platform-dependant assembly file
	asm, err := emitter.EmitAsm()
	if err != nil {
		panic(err)
	}
	fmt.Println("Assembly:\n" + asm)

	// Emit object file contents as bytes (GCIL -> LLVM IR -> object file)
	object, err := emitter.EmitObject()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Object file:\n%v\n", object)

	// Emit executable file as "a.out". This is the final result we want!
	// It links the object file and runtime with a linker.
	// (GCIL -> LLVM IR -> assembly -> object -> executable)
	if err := emitter.EmitExecutable("a.out"); err != nil {
		panic(err)
	}
}
