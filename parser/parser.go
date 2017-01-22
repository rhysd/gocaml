package parser

import (
	"bytes"
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
)

type lexerWrapper struct {
	tokens       chan token.Token
	errorCount   int
	errorMessage bytes.Buffer
	result       ast.Expr
}

func newLexerWrapper(tokens chan token.Token) *lexerWrapper {
	return &lexerWrapper{
		tokens:     tokens,
		errorCount: 0,
	}
}

func (w *lexerWrapper) Lex(lval *yySymType) int {
	select {
	case t := <-w.tokens:
		lval.token = &t
		return int(t.Kind) // XXX: Is this correct?
	}
	panic("Unreachable")
}

func (w *lexerWrapper) Error(msg string) {
	w.errorCount += 1
	w.errorMessage.WriteString(fmt.Sprintf("  * %s\n", msg))
}

func (w *lexerWrapper) getErrorMessage() error {
	return fmt.Errorf("%d errors while parsing\n%s", w.errorCount, w.errorMessage.String())
}

func Parse(tokens chan token.Token) (ast.Expr, error) {
	yyErrorVerbose = true

	wrapper := &lexerWrapper{tokens: tokens}
	ret := yyParse(wrapper)

	if ret != 0 {
		return nil, wrapper.getErrorMessage()
	}

	root := wrapper.result
	if root == nil {
		panic("FATAL: Parsing was successfully done but result was not set")
	}

	return root, nil
}
