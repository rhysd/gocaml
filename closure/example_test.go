package closure

import (
	"os"
	"path/filepath"

	"github.com/rhysd/gocaml/sema"
	"github.com/rhysd/gocaml/syntax"
	"github.com/rhysd/locerr"
)

func Example() {
	file := filepath.FromSlash("../testdata/from-mincaml/ack.ml")
	src, err := locerr.NewSourceFromFile(file)
	if err != nil {
		// File not found
		panic(err)
	}

	ast, err := syntax.Parse(src)
	if err != nil {
		// When parse failed
		panic(err)
	}

	// Resolving symbols, type analysis and converting AST into MIR instruction block
	env, block, err := sema.SemanticsCheck(ast)
	if err != nil {
		// Type error detected
		panic(err)
	}

	// Closure transform.
	// Move all nested function to toplevel with resolving closures and known
	// function optimization.
	// Returned value will represents converted whole program.
	// It contains entry point of program, some toplevel functions and closure
	// information.
	program := Transform(block)

	// For debug purpose, you can show MIR representation after conversion
	program.Println(os.Stdout, env)
	// Output:
	// ack$t1 = recfun x$t2,y$t3 ; type=int -> int -> int
	//   BEGIN: body (ack$t1)
	//   $k2 = int 0 ; type=int
	//   $k3 = binary <= x$t2 $k2 ; type=bool
	//   $k28 = if $k3 ; type=int
	//     BEGIN: then
	//     $k5 = int 1 ; type=int
	//     $k6 = binary + y$t3 $k5 ; type=int
	//     END: then
	//     BEGIN: else
	//     $k8 = int 0 ; type=int
	//     $k9 = binary <= y$t3 $k8 ; type=bool
	//     $k27 = if $k9 ; type=int
	//       BEGIN: then
	//       $k12 = int 1 ; type=int
	//       $k13 = binary - x$t2 $k12 ; type=int
	//       $k14 = int 1 ; type=int
	//       $k15 = app ack$t1 $k13,$k14 ; type=int
	//       END: then
	//       BEGIN: else
	//       $k18 = int 1 ; type=int
	//       $k19 = binary - x$t2 $k18 ; type=int
	//       $k23 = int 1 ; type=int
	//       $k24 = binary - y$t3 $k23 ; type=int
	//       $k25 = app ack$t1 x$t2,$k24 ; type=int
	//       $k26 = app ack$t1 $k19,$k25 ; type=int
	//       END: else
	//     END: else
	//   END: body (ack$t1)
	//
	// BEGIN: program
	// $k31 = int 3 ; type=int
	// $k32 = int 10 ; type=int
	// $k33 = app ack$t1 $k31,$k32 ; type=int
	// $k34 = appx print_int $k33 ; type=unit
	// END: program
}
