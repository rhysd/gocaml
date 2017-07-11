// Package token defines tokens of GoCaml source codes.
package token

import (
	"fmt"
	"github.com/rhysd/locerr"
)

type Kind int

const (
	ILLEGAL Kind = iota
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
	ARRAY_MAKE
	DOT
	LESS_MINUS
	SEMICOLON
	STAR
	SLASH
	BAR_BAR
	AND_AND
	ARRAY_LENGTH
	STRING_LITERAL
	PERCENT
	MATCH
	WITH
	BAR
	SOME
	NONE
	MINUS_GREATER
	FUN
	COLON
	TYPE
	LBRACKET_BAR
	BAR_RBRACKET
	LBRACKET
	RBRACKET
	EXTERNAL
	EOF
)

var tokenTable = [...]string{
	ILLEGAL:        "ILLEGAL",
	EOF:            "EOF",
	COMMENT:        "COMMENT",
	LPAREN:         "(",
	RPAREN:         ")",
	IDENT:          "IDENT",
	BOOL:           "BOOL",
	NOT:            "NOT",
	INT:            "INT",
	FLOAT:          "FLOAT",
	MINUS:          "-",
	PLUS:           "+",
	MINUS_DOT:      "-.",
	PLUS_DOT:       "+.",
	STAR_DOT:       "*.",
	SLASH_DOT:      "/.",
	EQUAL:          "=",
	LESS_GREATER:   "<>",
	LESS_EQUAL:     "<=",
	LESS:           "<",
	GREATER:        ">",
	GREATER_EQUAL:  ">=",
	IF:             "if",
	THEN:           "then",
	ELSE:           "else",
	LET:            "let",
	IN:             "in",
	REC:            "rec",
	COMMA:          ",",
	ARRAY_MAKE:     "Array.make",
	DOT:            ".",
	LESS_MINUS:     "<-",
	SEMICOLON:      ";",
	STAR:           "*",
	SLASH:          "/",
	BAR_BAR:        "||",
	AND_AND:        "&&",
	ARRAY_LENGTH:   "Array.length",
	STRING_LITERAL: "STRING_LITERAL",
	PERCENT:        "%",
	MATCH:          "match",
	WITH:           "with",
	BAR:            "|",
	SOME:           "Some",
	NONE:           "None",
	MINUS_GREATER:  "->",
	FUN:            "fun",
	COLON:          ":",
	TYPE:           "type",
	LBRACKET_BAR:   "[|",
	BAR_RBRACKET:   "|]",
	LBRACKET:       "[",
	RBRACKET:       "]",
	EXTERNAL:       "external",
}

// Token instance for GoCaml.
// It contains its location information and kind.
type Token struct {
	Kind  Kind
	Start locerr.Pos
	End   locerr.Pos
	File  *locerr.Source
}

// String returns an information of token. This method is used mainly for
// debug purpose.
func (tok *Token) String() string {
	return fmt.Sprintf(
		"<%s:%s>(%d:%d:%d-%d:%d:%d)",
		tokenTable[tok.Kind],
		tok.Value(),
		tok.Start.Line, tok.Start.Column, tok.Start.Offset,
		tok.End.Line, tok.End.Column, tok.End.Offset)
}

// Value returns the corresponding a string part of code.
func (tok *Token) Value() string {
	return string(tok.File.Code[tok.Start.Offset:tok.End.Offset])
}
