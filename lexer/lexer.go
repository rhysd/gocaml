package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/rhysd/mincaml-parser/token"
	"io"
	"unicode"
	"unicode/utf8"
)

type stateFn func(*Lexer) stateFn

const eof = -1

type Lexer struct {
	state   stateFn
	start   token.Position
	current token.Position
	buf     bytes.Buffer
	src     *bufio.Reader
	tokens  chan token.Token
	top     rune
	eof     bool
	file    string
}

func NewLexer(file string, src io.Reader, tokens chan token.Token) *Lexer {
	start := token.Position{1, 1, 0}
	return &Lexer{
		state:   lex,
		start:   start,
		current: start,
		src:     bufio.NewReader(src),
		tokens:  tokens,
		file:    file,
	}
}

func (l *Lexer) Lex() {
	l.eat()
	for l.state != nil {
		l.state = l.state(l)
	}
}

func (l *Lexer) emit(kind token.TokenKind) {
	l.tokens <- token.Token{
		kind,
		l.buf.String(),
		l.start,
		l.current,
		l.file,
	}
	l.buf.Truncate(0)
	l.start = l.current
}

func (l *Lexer) emitIdent() {
	s := l.buf.String()
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

func (l *Lexer) emitIllegal(reason string) {
	l.tokens <- token.Token{
		token.ILLEGAL,
		reason,
		l.current,
		l.current,
		l.file,
	}
	l.buf.Truncate(0)
	l.start = l.current
}

func (l *Lexer) emitExpected(s string, a rune) {
	l.tokens <- token.Token{
		token.ILLEGAL,
		fmt.Sprintf("Expected %s but got '%c'", s, a),
		l.current,
		l.current,
		l.file,
	}
	l.buf.Truncate(0)
	l.start = l.current
}

func (l *Lexer) emitUnexpectedEOF(expected string) {
	if len(expected) == 1 {
		l.emitIllegal(fmt.Sprintf("Expected '%s' but got EOF", expected))
	} else {
		l.emitIllegal(fmt.Sprintf("Expected one of '%s' but got EOF", expected))
	}
}

func (l *Lexer) eat() {
	r, size, err := l.src.ReadRune()
	if err == io.EOF {
		l.top = 0
		l.eof = true
		return
	}

	if err != nil {
		panic(err)
	}

	l.current.Offset += size
	if r == '\n' {
		l.current.Line += 1
		l.current.Column = 1
	} else {
		l.current.Column += 1
	}
	l.buf.WriteRune(r)

	l.top = r
	l.eof = false
}

func lexComment(l *Lexer) stateFn {
	for {
		if l.eof {
			l.emitUnexpectedEOF("*")
			return nil
		}
		if l.top == '*' {
			l.eat()
			if l.eof {
				l.emitUnexpectedEOF(")")
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
		l.emitExpected("'.'", l.top)
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
			l.emitExpected("number", l.top)
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
		l.emitExpected("letter", l.top)
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
				l.eat()
			case unicode.IsDigit(l.top):
				return lexNumber
			default:
				return lexIdent
			}
		}
	}
}
