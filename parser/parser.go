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

func (l *pseudoLexer) Lex(lval *yySymType) int {
	for {
		select {
		case t := <-l.tokens:
			lval.token = &t

			switch t.Kind {
			case token.EOF:
				// Zero means input ends
				// (see golang.org/x/tools/cmd/goyacc/testdata/expr/expr.y)
				return 0
			case token.COMMENT:
				continue
			}

			// XXX:
			// Converting token value into yacc's token.
			// This conversion requires that token order must the same as
			// yacc's token order. EOF is a first token. So we can use it
			// to make an offset between token value and yacc's token value.
			return int(t.Kind) + ILLEGAL
		}
	}
	panic("Unreachable")
}

func (l *pseudoLexer) Error(msg string) {
	l.errorCount += 1
	l.errorMessage.WriteString(fmt.Sprintf("  * %s\n", msg))
}

func (l *pseudoLexer) getErrorMessage() error {
	return fmt.Errorf("%d errors while parsing\n%s", l.errorCount, l.errorMessage.String())
}

func Parse(tokens chan token.Token) (ast.Expr, error) {
	yyErrorVerbose = true

	l := &pseudoLexer{tokens: tokens}
	ret := yyParse(l)

	if ret != 0 {
		return nil, l.getErrorMessage()
	}

	root := l.result
	if root == nil {
		return nil, fmt.Errorf("")
		panic("FATAL: Parsing was successfully done but result was not set")
	}

	return root, nil
}
