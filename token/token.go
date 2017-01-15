package token

import (
	"fmt"
)

type Pos int
type TokenKind int

const (
	ILLEGAL TokenKind = iota
	EOF
	COMMENT
	NUMBER
	LPAREN
	RPAREN
	IDENT
)

var TokenStrings = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",
	NUMBER:  "NUMBER",
	LPAREN:  "(",
	RPAREN:  ")",
	IDENT:   "IDENT",
}

type Token struct {
	Kind   TokenKind
	Line   int
	Column int
	Offset int
	Value  string
}

func (tok *Token) String() string {
	return fmt.Sprintf("<%s:%d:%d:%d:%s>", TokenStrings[tok.Kind], tok.Line, tok.Column, tok.Offset, tok.Value)
}
