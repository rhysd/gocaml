// Package compiler provides a compiler function for GoCaml codes.
package compiler

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"os"
)

// Compiler instance to compile GoCaml code into other representations.
type Compiler struct {
	// Compiler options (e.g. optimization level) go here.
}

func (c *Compiler) Compile(source *token.Source) error {
	// TODO
	return nil
}

// PrintTokens returns the lexed tokens for a source code.
func (c *Compiler) Lex(src *token.Source) chan token.Token {
	l := lexer.NewLexer(src)
	l.Error = func(msg string, pos token.Position) {
		fmt.Fprintf(os.Stderr, "%s at (line:%d, column:%d)\n", msg, pos.Line, pos.Column)
	}
	go l.Lex()
	return l.Tokens
}

// PrintTokens show list of tokens lexed.
func (c *Compiler) PrintTokens(src *token.Source) {
	tokens := c.Lex(src)
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
func (c *Compiler) Parse(src *token.Source) (*ast.AST, error) {
	tokens := c.Lex(src)
	root, err := parser.Parse(tokens)

	if err != nil {
		return nil, err
	}

	ast := &ast.AST{
		File: src,
		Root: root,
	}

	return ast, nil
}

// PrintAST outputs AST structure to stdout.
func (c *Compiler) PrintAST(src *token.Source) {
	a, err := c.Parse(src)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	ast.Println(a)
}

// SemanticAnalysis checks types and symbol duplicates.
// It returns the result of type analysis or an error.
func (c *Compiler) SemanticAnalysis(a *ast.AST) (*typing.Env, error) {
	if err := alpha.Transform(a.Root); err != nil {
		return nil, errors.Wrapf(err, "While semantic analysis (alpha transform) for %s\n", a.File.Name)
	}
	env := typing.NewEnv()
	if err := env.ApplyTypeAnalysis(a.Root); err != nil {
		return nil, errors.Wrapf(err, "While semantic analysis (type inferernce) for %s\n", a.File.Name)
	}
	return env, nil
}

// EmitGCIL emits GCIL tree representation.
func (c *Compiler) EmitGCIL(src *token.Source) (*gcil.Block, *typing.Env, error) {
	ast, err := c.Parse(src)
	if err != nil {
		return nil, nil, err
	}
	env, err := c.SemanticAnalysis(ast)
	if err != nil {
		return nil, nil, err
	}
	ir := gcil.EmitIR(ast.Root, env)
	gcil.ElimRefs(ir, env)
	return ir, env, nil
}
