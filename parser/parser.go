package parser

import (
	"bytes"
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
)

type pseudoLexer struct {
	tokens       chan token.Token
	errorCount   int
	errorMessage bytes.Buffer
	result       ast.Expr
}

func newLexerWrapper(tokens chan token.Token) *pseudoLexer {
	return &pseudoLexer{
		tokens:     tokens,
		errorCount: 0,
	}
}

func (w *pseudoLexer) Lex(lval *yySymType) int {
	select {
	case t := <-w.tokens:
		lval.token = &t
		fmt.Printf("t: %d: %s\n", int(t.Kind), t.String())

		// XXX:
		// Converting token value into yacc's token.
		// This conversion requires that token order must the same as
		// yacc's token order. EOF is a first token. So we can use it
		// to make an offset between token value and yacc's token value.
		return int(t.Kind) + EOF
	}
	panic("Unreachable")
}

func (w *pseudoLexer) Error(msg string) {
	w.errorCount += 1
	w.errorMessage.WriteString(fmt.Sprintf("  * %s\n", msg))
}

func (w *pseudoLexer) getErrorMessage() error {
	return fmt.Errorf("%d errors while parsing\n%s", w.errorCount, w.errorMessage.String())
}

func Parse(tokens chan token.Token) (ast.Expr, error) {
	yyErrorVerbose = true
	yyDebug = 9999

	wrapper := &pseudoLexer{tokens: tokens}
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
