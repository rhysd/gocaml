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
	block.Println(os.Stdout)
	// Output:
	// BEGIN: program
	// ack$t1 = fun x$t2,y$t3
	// BEGIN: body (ack$t1)
	// $k1 = int 0 ; type=int
	// $k2 = ref x$t2 ; type=int
	// $k3 = binary < $k1 $k2 ; type=bool
	// $k4 = unary not $k3 ; type=bool
	// $k30 = if $k4
	// BEGIN: then
	// $k5 = ref y$t3 ; type=int
	// $k6 = int 1 ; type=int
	// $k7 = binary + $k5 $k6 ; type=int
	// END: then
	// BEGIN: else
	// $k8 = int 0 ; type=int
	// $k9 = ref y$t3 ; type=int
	// $k10 = binary < $k8 $k9 ; type=bool
	// $k11 = unary not $k10 ; type=bool
	// $k29 = if $k11
	// BEGIN: then
	// $k12 = ref ack$t1 ; type=int -> int -> int
	// $k13 = ref x$t2 ; type=int
	// $k14 = int 1 ; type=int
	// $k15 = binary - $k13 $k14 ; type=int
	// $k16 = int 1 ; type=int
	// $k17 = app $k12 $k15,$k16 ; type=int
	// END: then
	// BEGIN: else
	// $k18 = ref ack$t1 ; type=int -> int -> int
	// $k19 = ref x$t2 ; type=int
	// $k20 = int 1 ; type=int
	// $k21 = binary - $k19 $k20 ; type=int
	// $k22 = ref ack$t1 ; type=int -> int -> int
	// $k23 = ref x$t2 ; type=int
	// $k24 = ref y$t3 ; type=int
	// $k25 = int 1 ; type=int
	// $k26 = binary - $k24 $k25 ; type=int
	// $k27 = app $k22 $k23,$k26 ; type=int
	// $k28 = app $k18 $k21,$k27 ; type=int
	// END: else
	//  ; type=int
	// END: else
	//  ; type=int
	// END: body (ack$t1)
	//  ; type=int -> int -> int
	// $k31 = xref print_int ; type=int -> ()
	// $k32 = ref ack$t1 ; type=int -> int -> int
	// $k33 = int 3 ; type=int
	// $k34 = int 10 ; type=int
	// $k35 = app $k32 $k33,$k34 ; type=int
	// $k36 = app $k31 $k35 ; type=()
	// END: program

}
