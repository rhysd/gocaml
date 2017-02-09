package main

import (
	"flag"
	"fmt"
	"github.com/rhysd/gocaml/compiler"
	"github.com/rhysd/gocaml/token"
	"os"
)

var (
	help       = flag.Bool("help", false, "Show this help")
	showTokens = flag.Bool("tokens", false, "Show tokens for input")
	showAST    = flag.Bool("ast", false, "Show AST for input")
	showGCIL   = flag.Bool("gcil", false, "Emit GoCaml Intermediate Language representation to stdout")
	externals  = flag.Bool("externals", false, "Display external symbols")
)

const usageHeader = `Usage: gocaml [flags] [file]

  Compiler for GoCaml.
  When file is given as argument, compiler will targets it. Otherwise, compiler
  attempt to read from STDIN as source code to target.

Flags:`

func usage() {
	fmt.Fprintln(os.Stderr, usageHeader)
	flag.PrintDefaults()
}

func getSource(args []string) (*token.Source, error) {
	if len(args) == 0 {
		return token.NewSourceFromStdin()
	}
	return token.NewSourceFromFile(args[1])
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
		os.Exit(0)
	}

	var src *token.Source
	var err error

	if flag.NArg() == 0 {
		src, err = token.NewSourceFromStdin()
	} else {
		src, err = token.NewSourceFromFile(flag.Arg(0))
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error on opening source: %s\n", err.Error())
		os.Exit(4)
	}

	c := compiler.Compiler{}

	switch {
	case *showTokens:
		c.PrintTokens(src)
	case *showAST:
		c.PrintAST(src)
	case *showGCIL:
		prog, env, err := c.EmitGCIL(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
		prog.Dump(env)
	default:
		prog, env, err := c.EmitGCIL(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
		prog.Dump(env)
	}
}
