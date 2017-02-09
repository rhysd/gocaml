package gcil

import (
	"fmt"
	"github.com/rhysd/gocaml/typing"
	"os"
	"strings"
)

// Program representation. Program can be obtained after closure transform because
// all functions must be at the top.
type Program struct {
	Toplevel map[string]*Fun     // Mapping from function name to its instruction
	Closures map[string][]string // Mapping from closure name to it free variables
	Entry    *Block
}

func (prog *Program) Dump(env *typing.Env) {

	fmt.Println("TOPLEVELS:")
	p := printer{env, os.Stdout, ""}
	for n, f := range prog.Toplevel {
		p.printlnInsn(NewInsn(n, f))
		fmt.Println()
	}

	fmt.Println("CLOSURES:")
	for c, fv := range prog.Closures {
		fmt.Printf("%s:\t%s\n", c, strings.Join(fv, ","))
	}
	fmt.Println()

	fmt.Println("ENTRY:")
	prog.Entry.Println(os.Stdout, env)
}
