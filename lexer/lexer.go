package lexer

import (
	"bytes"
	"fmt"
	"github.com/rhysd/mincaml-parser/source"
	"github.com/rhysd/mincaml-parser/token"
	"io"
	"os"
	"unicode"
	"unicode/utf8"
)

type stateFn func(*Lexer) stateFn

const eof = -1

type Lexer struct {
	state   stateFn
	start   token.Position
	current token.Position
	src     *source.Source
	input   *bytes.Reader
	tokens  chan token.Token
	top     rune
	eof     bool
	Error   func(msg string, t token.Token)
}

func NewLexer(src *source.Source, tokens chan token.Token) *Lexer {
	start := token.Position{
		Offset: 0,
		Line:   1,
		Column: 1,
	}
	return &Lexer{
		state:   lex,
		start:   start,
		current: start,
		input:   bytes.NewReader(src.Code),
		src:     src,
		tokens:  tokens,
		Error: func(m string, t token.Token) {
			fmt.Fprintf(os.Stderr, "Error at %s: %s", t.String(), m)
		},
	}
}

func (l *Lexer) Lex() {
	// Set top to peek current rune
	l.consume()
	for l.state != nil {
		l.state = l.state(l)
	}
}

func (l *Lexer) emit(kind token.TokenKind) {
	l.tokens <- token.Token{
		kind,
		l.start,
		l.current,
		l.src,
	}
	l.start = l.current
}

func (l *Lexer) emitIdent() {
	s := string(l.src.Code[l.start.Offset:l.current.Offset])
	if len(s) == 1 {
		// Shortcut because no keyword is one character. It must be identifier
		l.emit(token.IDENT)
		return
	}

	switch s {
	case "true", "false":
		l.emit(token.BOOL)
	case "if":
		l.emit(token.IF)
	case "then":
		l.emit(token.THEN)
	case "else":
		l.emit(token.ELSE)
	case "let":
		l.emit(token.LET)
	case "in":
		l.emit(token.IN)
	case "rec":
		l.emit(token.REC)
	case "Array":
		l.emit(token.ARRAY)
	case "create":
		l.emit(token.CREATE)
	default:
		l.emit(token.IDENT)
	}
}

func (l *Lexer) emitIllegal() token.Token {
	t := token.Token{
		token.ILLEGAL,
		l.start,
		l.current,
		l.src,
	}
	l.tokens <- t
	l.start = l.current
	return t
}

func (l *Lexer) expected(s string, a rune) {
	t := l.emitIllegal()
	l.Error(fmt.Sprintf("Expected %s but got '%c'(%d)", s, a, a), t)
}

func (l *Lexer) unexpectedEOF(expected string) {
	t := l.emitIllegal()
	if len(expected) == 1 {
		l.Error(fmt.Sprintf("Expected '%s' but got EOF", expected), t)
	} else {
		l.Error(fmt.Sprintf("Expected one of '%s' but got EOF", expected), t)
	}
}

func (l *Lexer) consume() {
	r, _, err := l.input.ReadRune()
	if err == io.EOF {
		l.top = 0
		l.eof = true
		return
	}

	if err != nil {
		panic(err)
	}

	l.top = r
	l.eof = false
}

func (l *Lexer) eat() {
	size := utf8.RuneLen(l.top)
	l.current.Offset += size

	// TODO: Consider \n\r
	if l.top == '\n' {
		l.current.Line += 1
		l.current.Column = 1
	} else {
		l.current.Column += size
	}

	l.consume()
}

func (l *Lexer) skip() {
	if l.eof {
		return
	}
	l.eat()
	l.start = l.current
}

func lexComment(l *Lexer) stateFn {
	for {
		if l.eof {
			l.unexpectedEOF("*")
			return nil
		}
		if l.top == '*' {
			l.eat()
			if l.eof {
				l.unexpectedEOF(")")
				return nil
			}
			if l.top == ')' {
				l.eat()
				l.emit(token.COMMENT)
				return lex
			}
		}
		l.eat()
	}
}

func lexLeftParen(l *Lexer) stateFn {
	l.eat()
	if l.top == '*' {
		l.eat()
		return lexComment
	} else {
		l.emit(token.LPAREN)
		return lex
	}
}

func lexAdditiveOp(l *Lexer) stateFn {
	dot, op := token.PLUS_DOT, token.PLUS
	if l.top == '-' {
		dot, op = token.MINUS_DOT, token.MINUS
	}
	l.eat()

	if l.top == '.' {
		l.eat()
		l.emit(dot)
	} else {
		l.emit(op)
	}
	return lex
}

func lexMultOp(l *Lexer) stateFn {
	op := token.STAR_DOT
	if l.top == '/' {
		op = token.SLASH_DOT
	}
	l.eat()

	if l.top != '.' {
		l.expected("'.'", l.top)
		return nil
	}

	l.eat()
	l.emit(op)
	return lex
}

func lexLess(l *Lexer) stateFn {
	l.eat()
	switch l.top {
	case '>':
		l.eat()
		l.emit(token.LESS_GREATER)
	case '=':
		l.eat()
		l.emit(token.LESS_EQUAL)
	case '-':
		l.eat()
		l.emit(token.LESS_MINUS)
	default:
		l.emit(token.LESS)
	}
	return lex
}

func lexGreater(l *Lexer) stateFn {
	l.eat()
	switch l.top {
	case '=':
		l.eat()
		l.emit(token.GREATER_EQUAL)
	default:
		l.emit(token.GREATER)
	}
	return lex
}

// e.g. 123.45e10
func lexNumber(l *Lexer) stateFn {
	tok := token.INT

	// Eat first digit. It's known as digit in lex()
	l.eat()
	for unicode.IsDigit(l.top) {
		l.eat()
	}

	// Note: Allow 1. as 1.0
	if l.top == '.' {
		tok = token.FLOAT
		l.eat()
		for unicode.IsDigit(l.top) {
			l.eat()
		}
	}

	if l.top == 'e' || l.top == 'E' {
		tok = token.FLOAT
		l.eat()
		if l.top == '+' || l.top == '-' {
			l.eat()
		}
		if !unicode.IsDigit(l.top) {
			l.expected("number", l.top)
			return nil
		}
		for unicode.IsDigit(l.top) {
			l.eat()
		}
	}

	l.emit(tok)
	return lex
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z' ||
		r == '_' ||
		r >= utf8.RuneSelf && unicode.IsLetter(r)
}

func lexIdent(l *Lexer) stateFn {
	if !isLetter(l.top) {
		l.expected("letter", l.top)
		return nil
	}
	l.eat()

	for isLetter(l.top) || unicode.IsDigit(l.top) {
		l.eat()
	}

	l.emitIdent()
	return lex
}

func lex(l *Lexer) stateFn {
	for {
		if l.eof {
			l.emit(token.EOF)
			return nil
		}
		switch l.top {
		case '(':
			return lexLeftParen
		case ')':
			l.eat()
			l.emit(token.RPAREN)
		case '+':
			return lexAdditiveOp
		case '-':
			return lexAdditiveOp
		case '*':
			return lexMultOp
		case '/':
			return lexMultOp
		case '=':
			l.eat()
			l.emit(token.EQUAL)
		case '<':
			return lexLess
		case '>':
			return lexGreater
		case ',':
			l.eat()
			l.emit(token.COMMA)
		case '.':
			l.eat()
			l.emit(token.DOT)
		case ';':
			l.eat()
			l.emit(token.SEMICOLON)
		default:
			switch {
			case unicode.IsSpace(l.top):
				l.skip()
			case unicode.IsDigit(l.top):
				return lexNumber
			default:
				return lexIdent
			}
		}
	}
}
