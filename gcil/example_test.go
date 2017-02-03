package gcil

import (
	"github.com/rhysd/gocaml/alpha"
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
	// Returned block represents the root block of program
	block := EmitIR(root, env)

	// Instructions are represented as list of instructions.
	// Block has pointers to access to the head and tail of the list.
	//
	_ = block.Top
	_ = block.Bottom

	// For debug purpose, .Println() method can output instruction sequences
	block.Println(os.Stdout, env)
	// BEGIN: program
	// ack$t1 = fun x$t2,y$t3 ; int -> int -> int
	//   BEGIN: body (ack$t1)
	//   $k1 = int 0 ; int
	//   $k2 = ref x$t2 ; int
	//   $k3 = binary < $k1 $k2 ; bool
	//   $k4 = unary not $k3 ; bool
	//   $k30 = if $k4 ; int
	//     BEGIN: then
	//     $k5 = ref y$t3 ; int
	//     $k6 = int 1 ; int
	//     $k7 = binary + $k5 $k6 ; int
	//     END: then
	//     BEGIN: else
	//     $k8 = int 0 ; int
	//     $k9 = ref y$t3 ; int
	//     $k10 = binary < $k8 $k9 ; bool
	//     $k11 = unary not $k10 ; bool
	//     $k29 = if $k11 ; int
	//       BEGIN: then
	//       $k12 = ref ack$t1 ; int -> int -> int
	//       $k13 = ref x$t2 ; int
	//       $k14 = int 1 ; int
	//       $k15 = binary - $k13 $k14 ; int
	//       $k16 = int 1 ; int
	//       $k17 = app $k12 $k15,$k16 ; int
	//       END: then
	//       BEGIN: else
	//       $k18 = ref ack$t1 ; int -> int -> int
	//       $k19 = ref x$t2 ; int
	//       $k20 = int 1 ; int
	//       $k21 = binary - $k19 $k20 ; int
	//       $k22 = ref ack$t1 ; int -> int -> int
	//       $k23 = ref x$t2 ; int
	//       $k24 = ref y$t3 ; int
	//       $k25 = int 1 ; int
	//       $k26 = binary - $k24 $k25 ; int
	//       $k27 = app $k22 $k23,$k26 ; int
	//       $k28 = app $k18 $k21,$k27 ; int
	//       END: else
	//     END: else
	//   END: body (ack$t1)
	// $k31 = xref print_int ; int -> ()
	// $k32 = ref ack$t1 ; int -> int -> int
	// $k33 = int 3 ; int
	// $k34 = int 10 ; int
	// $k35 = app $k32 $k33,$k34 ; int
	// $k36 = app $k31 $k35 ; ()
	// END: program

}
