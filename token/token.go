package token

import (
	"fmt"
)

type Position struct {
	Offset int
	Line   int
	Column int
}

type TokenKind int

const (
	ILLEGAL TokenKind = iota
	EOF
	COMMENT
	LPAREN
	RPAREN
	IDENT
	BOOL
	NOT
	INT
	FLOAT
	MINUS
	PLUS
	MINUS_DOT
	PLUS_DOT
	STAR_DOT
	SLASH_DOT
	EQUAL
	LESS_GREATER
	LESS_EQUAL
	LESS
	GREATER
	GREATER_EQUAL
	IF
	THEN
	ELSE
	LET
	IN
	REC
	COMMA
	ARRAY
	CREATE
	DOT
	LESS_MINUS
	SEMICOLON
)

var TokenStrings = [...]string{
	ILLEGAL:       "ILLEGAL",
	EOF:           "EOF",
	COMMENT:       "COMMENT",
	LPAREN:        "(",
	RPAREN:        ")",
	IDENT:         "IDENT",
	BOOL:          "BOOL",
	NOT:           "not",
	INT:           "INT",
	FLOAT:         "FLOAT",
	MINUS:         "-",
	PLUS:          "+",
	MINUS_DOT:     "-.",
	PLUS_DOT:      "+.",
	STAR_DOT:      "*.",
	SLASH_DOT:     "/.",
	EQUAL:         "=",
	LESS_GREATER:  "<>",
	LESS_EQUAL:    "<=",
	LESS:          "<",
	GREATER:       ">",
	GREATER_EQUAL: ">=",
	IF:            "if",
	THEN:          "then",
	ELSE:          "else",
	LET:           "let",
	IN:            "in",
	REC:           "rec",
	COMMA:         ",",
	ARRAY:         "Array",
	CREATE:        "create",
	DOT:           ".",
	LESS_MINUS:    "<-",
	SEMICOLON:     ";",
}

type Token struct {
	Kind  TokenKind
	Value string
	Start Position
	End   Position
	File  string
}

func (tok *Token) String() string {
	return fmt.Sprintf(
		"<%s:%s>(%d:%d:%d-%d:%d:%d)",
		TokenStrings[tok.Kind],
		tok.Value,
		tok.Start.Line, tok.Start.Column, tok.Start.Offset,
		tok.End.Line, tok.End.Column, tok.End.Offset)
}