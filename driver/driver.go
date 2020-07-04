// Package driver is amediator to glue all packages for GoCaqml. provides a compiler function for GoCaml codes.
// It provides compiler functinalities for GoCaml.
package driver

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/closure"
	"github.com/rhysd/gocaml/codegen"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/mono"
	"github.com/rhysd/gocaml/sema"
	"github.com/rhysd/gocaml/syntax"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
	"io/ioutil"
	"os"
	"path/filepath"
)

type OptLevel int

const (
	O0 OptLevel = iota
	O1
	O2
	O3
)

// Driver instance to compile GoCaml code into other representations.
type Driver struct {
	Optimization OptLevel
	LinkFlags    string
	TargetTriple string
	DebugInfo    bool
}

// PrintTokens returns the lexed tokens for a source code.
func (d *Driver) Lex(src *locerr.Source) chan token.Token {
	l := syntax.NewLexer(src)
	l.Error = func(msg string, pos locerr.Pos) {
		err := locerr.ErrorAt(pos, msg)
		err.PrintToFile(os.Stderr)
		fmt.Fprintln(os.Stderr)
	}
	go l.Lex()
	return l.Tokens
}

// PrintTokens show list of tokens lexed.
func (d *Driver) PrintTokens(src *locerr.Source) {
	tokens := d.Lex(src)
	for {
		select {
		case t := <-tokens:
			fmt.Println(t.String())
			switch t.Kind {
			case token.EOF, token.ILLEGAL:
				return
			}
		}
	}
}

// Parse parses the source and returns the parsed AST.
func (d *Driver) Parse(src *locerr.Source) (*ast.AST, error) {
	return syntax.Parse(src)
}

// PrintAST outputs AST structure to stdout.
func (d *Driver) PrintAST(src *locerr.Source) {
	a, err := d.Parse(src)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	ast.Println(a)
}

// SemanticAnalysis checks symbol duplicates, infers types and so on. It returns analyzed type
// environment and inferred types of AST node.
func (d *Driver) SemanticAnalysis(src *locerr.Source) (*types.Env, sema.InferredTypes, error) {
	a, err := syntax.Parse(src)
	if err != nil {
		return nil, nil, err
	}

	return sema.Analyze(a)
}

func (d *Driver) DumpEnvToStdout(src *locerr.Source) error {
	env, inferred, err := d.SemanticAnalysis(src)
	if err != nil {
		return err
	}
	env.Dump()
	fmt.Println("\nType Information:\n")
	for expr, ty := range inferred {
		fmt.Printf("  %s: %s (%s-%s)\n", expr.Name(), ty.String(), expr.Pos().String(), expr.End().String())
	}
	return nil
}

// EmitMIR emits MIR tree representation.
func (d *Driver) EmitMIR(src *locerr.Source) (*mir.Program, *types.Env, error) {
	parsed, err := d.Parse(src)
	if err != nil {
		return nil, nil, err
	}
	env, ir, err := sema.SemanticsCheck(parsed)
	if err != nil {
		return nil, nil, err
	}
	prog := closure.Transform(ir)
	prog = mono.Monomorphize(prog, env)
	return prog, env, nil
}

func (d *Driver) emitterFromSource(src *locerr.Source) (*codegen.Emitter, error) {
	prog, env, err := d.EmitMIR(src)
	if err != nil {
		return nil, err
	}

	level := codegen.OptimizeDefault
	switch d.Optimization {
	case O0:
		level = codegen.OptimizeNone
	case O1:
		level = codegen.OptimizeLess
	case O3:
		level = codegen.OptimizeAggressive
	}
	opts := codegen.EmitOptions{level, d.TargetTriple, d.LinkFlags, d.DebugInfo}

	return codegen.NewEmitter(prog, env, src, opts)
}

func (d *Driver) EmitObjFile(src *locerr.Source) error {
	emitter, err := d.emitterFromSource(src)
	if err != nil {
		return err
	}
	defer emitter.Dispose()
	emitter.RunOptimizationPasses()
	obj, err := emitter.EmitObject()
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%s.o", src.BaseName())
	return ioutil.WriteFile(filename, obj, 0666)
}

func (d *Driver) EmitLLVMIR(src *locerr.Source) (string, error) {
	emitter, err := d.emitterFromSource(src)
	if err != nil {
		return "", err
	}
	defer emitter.Dispose()
	emitter.RunOptimizationPasses()

	return emitter.EmitLLVMIR(), nil
}

func (d *Driver) EmitAsm(src *locerr.Source) (string, error) {
	emitter, err := d.emitterFromSource(src)
	if err != nil {
		return "", err
	}
	defer emitter.Dispose()
	emitter.RunOptimizationPasses()

	return emitter.EmitAsm()
}

func (d *Driver) Compile(source *locerr.Source) error {
	emitter, err := d.emitterFromSource(source)
	if err != nil {
		return err
	}
	defer emitter.Dispose()
	emitter.RunOptimizationPasses()
	var executable string
	if source.Exists {
		executable = source.BaseName()
	} else {
		executable, err = filepath.Abs("a.out")
		if err != nil {
			return err
		}
	}
	return emitter.EmitExecutable(executable)
}
