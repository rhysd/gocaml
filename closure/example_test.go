package closure

import (
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"os"
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

	// Closure transform.
	// Move all nested function to toplevel with resolving closures and known
	// function optimization.
	// Returned value will represents converted whole program.
	// It contains entry point of program, some toplevel functions and closure
	// information.
	program := Transform(block)

	// For debug purpose, you can show GCIL representation after conversion
	program.Println(os.Stdout, env)
	// Output:
	// ack$t1 = recfun x$t2,y$t3 ; type=(int, int) -> int
	//   BEGIN: body (ack$t1)
	//   $k1 = int 0 ; type=int
	//   $k3 = binary < $k1 x$t2 ; type=bool
	//   $k4 = unary not $k3 ; type=bool
	//   $k30 = if $k4 ; type=int
	//     BEGIN: then
	//     $k6 = int 1 ; type=int
	//     $k7 = binary + y$t3 $k6 ; type=int
	//     END: then
	//     BEGIN: else
	//     $k8 = int 0 ; type=int
	//     $k10 = binary < $k8 y$t3 ; type=bool
	//     $k11 = unary not $k10 ; type=bool
	//     $k29 = if $k11 ; type=int
	//       BEGIN: then
	//       $k14 = int 1 ; type=int
	//       $k15 = binary - x$t2 $k14 ; type=int
	//       $k16 = int 1 ; type=int
	//       $k17 = app ack$t1 $k15,$k16 ; type=int
	//       END: then
	//       BEGIN: else
	//       $k20 = int 1 ; type=int
	//       $k21 = binary - x$t2 $k20 ; type=int
	//       $k25 = int 1 ; type=int
	//       $k26 = binary - y$t3 $k25 ; type=int
	//       $k27 = app ack$t1 x$t2,$k26 ; type=int
	//       $k28 = app ack$t1 $k21,$k27 ; type=int
	//       END: else
	//     END: else
	//   END: body (ack$t1)
	//
	// BEGIN: program
	// $k33 = int 3 ; type=int
	// $k34 = int 10 ; type=int
	// $k35 = app ack$t1 $k33,$k34 ; type=int
	// $k36 = appx print_int $k35 ; type=()
	// END: program
}
