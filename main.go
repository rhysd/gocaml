package main

import (
	"flag"
	"fmt"
	"github.com/rhysd/gocaml/codegen"
	"github.com/rhysd/gocaml/driver"
	"github.com/rhysd/locerr"
	"os"
	"strings"
)

var (
	help        = flag.Bool("help", false, "Show this help")
	showTokens  = flag.Bool("tokens", false, "Show tokens for input")
	showAST     = flag.Bool("ast", false, "Show AST for input")
	analyze     = flag.Bool("analyze", false, "Dump analyzed symbols and types information to stdout")
	showMIR     = flag.Bool("mir", false, "Emit GoCaml Intermediate Language representation to stdout")
	check       = flag.Bool("check", false, "Check code (syntax, types, ...) and report errors if exist")
	llvm        = flag.Bool("llvm", false, "Emit LLVM IR to stdout")
	asm         = flag.Bool("asm", false, "Emit assembler code to stdout")
	opt         = flag.Int("opt", -1, "Optimization level (0~3). 0: none, 1: less, 2: default, 3: aggressive")
	obj         = flag.Bool("obj", false, "Compile to object file")
	ldflags     = flag.String("ldflags", "", "Flags passed to underlying linker")
	debug       = flag.Bool("g", false, "Compile with debug information")
	target      = flag.String("target", "", "Target architecture triple")
	showTargets = flag.Bool("show-targets", false, "Show all available targets")
)

const usageHeader = `Usage: gocaml [flags] [file]

  Compiler for GoCaml.
  When file is given as argument, compiler will compile it. Otherwise, compiler
  attempt to read from STDIN as source code to compile.

Flags:`

func usage() {
	fmt.Fprintln(os.Stderr, usageHeader)
	flag.PrintDefaults()
}

func getOptLevel() driver.OptLevel {
	switch *opt {
	case 0:
		return driver.O0
	case 1:
		return driver.O1
	case 2:
		return driver.O2
	case 3:
		return driver.O3
	default:
		if *llvm {
			return driver.O0
		}
		return driver.O2
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
		os.Exit(0)
	}

	if *showTargets {
		for _, t := range codegen.AllTargets() {
			tabs := (23 - (len(t.Name) + 1)) / 8
			if tabs <= 0 {
				tabs = 1
			}
			pad := strings.Repeat("\t", tabs)
			fmt.Printf("%s:%s%s\n", t.Name, pad, t.Description)
		}
		os.Exit(0)
	}

	var src *locerr.Source
	var err error

	if flag.NArg() == 0 {
		src, err = locerr.NewSourceFromStdin()
	} else {
		src, err = locerr.NewSourceFromFile(flag.Arg(0))
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error on opening source: %s\n", err.Error())
		os.Exit(4)
	}

	d := driver.Driver{
		Optimization: getOptLevel(),
		TargetTriple: *target,
		LinkFlags:    *ldflags,
		DebugInfo:    *debug,
	}

	switch {
	case *showTokens:
		d.PrintTokens(src)
	case *showAST:
	case *check:
		d.PrintAST(src)
		if _, _, err := d.SemanticAnalysis(src); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
	case *analyze:
		if err := d.DumpEnvToStdout(src); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
	case *showMIR:
		prog, env, err := d.EmitMIR(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
		prog.Println(os.Stdout, env)
	case *llvm:
		ir, err := d.EmitLLVMIR(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
		fmt.Println(ir)
	case *asm:
		asm, err := d.EmitAsm(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
		fmt.Println(asm)
	case *obj:
		if err := d.EmitObjFile(src); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
	default:
		if err := d.Compile(src); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
	}
}
