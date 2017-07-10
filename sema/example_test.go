package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/syntax"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
	"os"
	"path/filepath"
)

func ExampleInferer_Infer() {
	// Type check example

	// Analyzing target
	src, err := locerr.NewSourceFromFile(filepath.FromSlash("../testdata/from-mincaml/ack.ml"))
	if err != nil {
		// File not found
		panic(err)
	}

	parsed, err := syntax.Parse(src)
	if err != nil {
		// When parse failed
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// Type environment for analysis
	env := types.NewEnv()

	// First, resolve all symbols by alpha transform
	if err := AlphaTransform(parsed, env); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// Second, run unification on all nodes and dereference type variables

	// Make a visitor to do type inferernce
	inferer := NewInferer(env)

	// Do type inference. It returns error if type mismatch was detected.
	if err := inferer.Infer(parsed); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// No error found!
	fmt.Println("OK")
	// Output: OK
}

func ExampleSemanticsCheck() {
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

	// Resolve symbols by alpha transform.
	// Then apply type inference. After this, all symbols in AST should have exact types. It also checks
	// types are valid and all types are determined by inference. It returns a type environment object
	// and converted MIR as the result.
	env, ir, err := SemanticsCheck(ast)
	if err != nil {
		// Type error detected
		panic(err)
	}

	// You can dump the type table
	env.Dump()

	ir.Println(os.Stdout, env)
	// Output:
	// ack$t1 = recfun x$t2,y$t3 ; type=(int, int) -> int
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

	// Optimization for eliminate unnecessary 'ref' instructions and classify
	// 'app' instruction for external function calls.
	mir.ElimRefs(ir, env)
}
