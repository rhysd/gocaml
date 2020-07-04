package mir

import (
	"fmt"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
	"io"
	"strings"
)

// Closures is a map from closure name to its captures
type Closures map[string][]string

type FunInsn struct {
	Name string
	Val  *Fun
	Pos  locerr.Pos
}

type Toplevel map[string]FunInsn

func NewToplevel() Toplevel {
	return map[string]FunInsn{}
}

func (top Toplevel) Add(n string, f *Fun, p locerr.Pos) {
	top[n] = FunInsn{n, f, p}
}

// Program representation. Program can be obtained after closure transform because
// all functions must be at the top.
type Program struct {
	Toplevel Toplevel // Mapping from function name to its instruction
	Closures Closures // Mapping from closure name to it free variables
	Entry    *Block
}

func (prog *Program) PrintToplevels(out io.Writer, env *types.Env) {
	p := printer{env, out, ""}
	for n, f := range prog.Toplevel {
		p.printlnInsn(NewInsn(n, f.Val, f.Pos))
		fmt.Fprintln(out)
	}
}

func (prog *Program) Dump(out io.Writer, env *types.Env) {
	fmt.Fprintf(out, "[TOPLEVELS (%d)]\n", len(prog.Toplevel))
	prog.PrintToplevels(out, env)

	fmt.Fprintf(out, "[CLOSURES (%d)]\n", len(prog.Closures))
	for c, fv := range prog.Closures {
		fmt.Fprintf(out, "%s:\t%s\n", c, strings.Join(fv, ","))
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "[ENTRY]")
	prog.Entry.Println(out, env)
}

func (prog *Program) Println(out io.Writer, env *types.Env) {
	prog.PrintToplevels(out, env)
	prog.Entry.Println(out, env)
}
