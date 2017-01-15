package lexer

import (
	"bytes"
	"fmt"
	"github.com/rhysd/mincaml-parser/token"
	"io"
	"text/scanner"
)

type stateFn func(*Lexer) stateFn

type LexError struct {
	ErrorCount int
	Filename   string
}

func (e *LexError) Error() string {
	return fmt.Sprintf("%d errors found on lexing '%s'", e.ErrorCount, e.Filename)
}

type Lexer struct {
	state  stateFn
	scan   scanner.Scanner
	pos    scanner.Position
	tokens chan token.Token
	// errors chan error
}

func LexAll(file string, src io.Reader) ([]token.Token, error) {
	c := make(chan token.Token)
	ts := []token.Token{}
	l := NewLexer(file, src, c)
	go l.Lex()
	for {
		select {
		case t := <-c:
			switch t.Kind {
			case token.EOF:
				return ts, nil
			case token.ILLEGAL:
				return nil, fmt.Errorf("Error at %s", t.String())
			default:
				ts = append(ts, t)
			}
		}
	}
}

func NewLexer(file string, src io.Reader, tokens chan token.Token) *Lexer {
	var s scanner.Scanner
	s.Filename = file
	s.Init(src)
	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats
	return &Lexer{
		scan:   s,
		tokens: tokens,
	}
}

func (l *Lexer) Lex() error {
	l.state = lex

	for l.state != nil {
		l.state = l.state(l)
	}

	if l.scan.ErrorCount > 0 {
		return &LexError{
			l.scan.ErrorCount,
			l.scan.Filename,
		}
	}

	return nil
}

func (l *Lexer) emit(kind token.TokenKind, value string) {
	l.tokens <- token.Token{
		kind,
		l.pos.Line,
		l.pos.Column,
		l.pos.Offset,
		value,
	}
}

// Assume '(' of '(*' was consumed
func lexComment(l *Lexer) stateFn {
	var b bytes.Buffer

	// Eat '*' of '(*'
	l.scan.Next()

	// Temporarily be aware of whitespaces in comments
	l.scan.Whitespace = uint64(0)

	for {
		r := l.scan.Scan()
		fmt.Printf("'%c'\n", r)
		switch l.scan.Scan() {
		case '*':
			if l.scan.Peek() == ')' {
				l.emit(token.COMMENT, b.String())
				l.scan.Next()
				l.scan.Whitespace = scanner.GoWhitespace
				return lex
			}
		case scanner.EOF:
			l.scan.Error(&l.scan, "Unclosed comment")
			l.emit(token.ILLEGAL, "")
			l.scan.Whitespace = scanner.GoWhitespace
			return lex
		case scanner.Int, scanner.Float, scanner.Ident:
			b.WriteString(l.scan.TokenText())
		default:
			b.WriteRune(r)
		}
	}
}

// TODO: Emit identifier or keywords
// func (l* Lexer) emitIdentifier(val string) {

func lex(l *Lexer) stateFn {
Loop:
	for {
		r := l.scan.Scan()
		switch r {
		case '(':
			if l.scan.Peek() == '*' {
				return lexComment
			} else {
				l.emit(token.LPAREN, token.TokenStrings[token.LPAREN])
			}
		case ')':
			l.emit(token.RPAREN, token.TokenStrings[token.RPAREN])
		case scanner.Int, scanner.Float:
			l.emit(token.NUMBER, l.scan.TokenText())
		case scanner.Ident:
			l.emit(token.IDENT, l.scan.TokenText())
		case scanner.EOF:
			break Loop
		default:
			l.scan.Error(&l.scan, fmt.Sprintf("Unexpected character '%c'", r))
			l.emit(token.ILLEGAL, "")
			break Loop
		}
	}
	l.emit(token.EOF, "")
	return nil
}
