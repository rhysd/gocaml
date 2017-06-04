// Package parser provides a parsing function for GoCaml.
package parser

import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/locerr"
)

type pseudoLexer struct {
	lastToken    *token.Token
	tokens       chan token.Token
	errorCount   int
	errorMessage string
	result       *ast.AST
}

func (l *pseudoLexer) Lex(lval *yySymType) int {
	for {
		select {
		case t := <-l.tokens:
			lval.token = &t

			switch t.Kind {
			case token.EOF, token.ILLEGAL:
				// Zero means input ends
				// (see golang.org/x/tools/cmd/goyacc/testdata/expr/expr.y)
				return 0
			case token.COMMENT:
				continue
			}

			l.lastToken = &t

			// XXX:
			// Converting token value into yacc's token.
			// This conversion requires that token order must the same as
			// yacc's token order. EOF is a first token. So we can use it
			// to make an offset between token value and yacc's token value.
			return int(t.Kind) + ILLEGAL
		}
	}
}

func (l *pseudoLexer) Error(msg string) {
	l.errorCount++
	l.errorMessage = msg
}

func (l *pseudoLexer) getError() error {
	msg := fmt.Sprintf("Error occurred while parsing. %s", l.errorMessage)
	if l.errorCount > 1 {
		msg = fmt.Sprintf("%s (%d error(s) remain)", msg, l.errorCount-1)
	}
	if l.lastToken != nil {
		return locerr.ErrorfAt(l.lastToken.Start, msg)
	}
	return locerr.NewError(msg)
}

// Parse parses given tokens and returns parsed AST.
// Tokens are passed via channel.
func Parse(tokens chan token.Token) (*ast.AST, error) {
	yyErrorVerbose = true

	l := &pseudoLexer{tokens: tokens}
	ret := yyParse(l)

	if ret != 0 || l.errorCount != 0 {
		return nil, l.getError()
	}

	root := l.result
	if root == nil {
		return nil, locerr.NewError("Parsing failed")
	}

	return root, nil
}
