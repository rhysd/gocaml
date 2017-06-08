package parser

import (
	"fmt"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/locerr"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseOK(t *testing.T) {
	for _, testdir := range []string{
		"../testdata/syntax",
		"../testdata/from-mincaml/",
	} {
		files, err := ioutil.ReadDir(filepath.FromSlash(testdir))
		if err != nil {
			panic(err)
		}

		for _, f := range files {
			n := filepath.Join(testdir, f.Name())
			if !strings.HasSuffix(n, ".ml") {
				continue
			}

			t.Run(fmt.Sprintf("Check parsing successfully: %s", n), func(t *testing.T) {
				s, err := locerr.NewSourceFromFile(n)
				if err != nil {
					panic(err)
				}

				l := lexer.NewLexer(s)
				go l.Lex()

				root, err := Parse(l.Tokens)
				if err != nil {
					t.Fatalf("Error on parsing %s: %s", f.Name(), err.Error())
				}
				if root == nil {
					t.Fatalf("Parsed node should not be nil")
				}
			})
		}
	}
}

func TestErrorOnListLiteral(t *testing.T) {
	for _, code := range []string{"[]", "[1; 2]", "[true; false;]"} {
		s := locerr.NewDummySource(code)
		l := lexer.NewLexer(s)
		go l.Lex()
		_, err := Parse(l.Tokens)
		if err == nil {
			t.Fatal("List literal must cause parse error:", code)
		}
		msg := err.Error()
		if !strings.Contains(msg, "List literal is not implemented yet.") {
			t.Fatal("Unexpected error message:", msg)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	src := locerr.NewDummySource("")
	tokens := []token.Token{
		token.Token{
			Kind:  token.IF,
			Start: locerr.Pos{0, 1, 1, src},
			End:   locerr.Pos{2, 1, 3, src},
		},
		token.Token{
			Kind:  token.IF,
			Start: locerr.Pos{3, 1, 4, src},
			End:   locerr.Pos{5, 1, 6, src},
		},
		token.Token{
			Kind:  token.EOF,
			Start: locerr.Pos{2, 1, 6, src},
			End:   locerr.Pos{2, 1, 6, src},
		},
	}
	c := make(chan token.Token)
	go func() {
		for _, t := range tokens {
			c <- t
		}
	}()
	r, err := Parse(c)
	if err == nil {
		t.Fatalf("Illegal token must raise an error but got %v", r)
	}
}

func TestTooLargeIntLiteral(t *testing.T) {
	src := locerr.NewDummySource("123456789123456789123456789123456789123456789")
	tokens := []token.Token{
		token.Token{
			Kind:  token.INT,
			Start: locerr.Pos{0, 1, 1, src},
			End:   locerr.Pos{45, 1, 45, src},
			File:  src,
		},
		token.Token{
			Kind:  token.EOF,
			Start: locerr.Pos{45, 1, 45, src},
			End:   locerr.Pos{45, 1, 45, src},
			File:  src,
		},
	}
	c := make(chan token.Token)
	go func() {
		for _, t := range tokens {
			c <- t
		}
	}()
	r, err := Parse(c)
	if err == nil {
		t.Fatalf("Invalid int literal must raise an error but got %v", r)
	}
	if !strings.Contains(err.Error(), "value out of range") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestInvalidStringLiteral(t *testing.T) {
	src := locerr.NewDummySource("\"a\nb\"\n")
	tokens := []token.Token{
		token.Token{
			Kind:  token.STRING_LITERAL,
			Start: locerr.Pos{1, 1, 0, src},
			End:   locerr.Pos{2, 3, 5, src},
			File:  src,
		},
		token.Token{
			Kind:  token.EOF,
			Start: locerr.Pos{3, 1, 6, src},
			End:   locerr.Pos{3, 1, 6, src},
			File:  src,
		},
	}
	c := make(chan token.Token)
	go func() {
		for _, t := range tokens {
			c <- t
		}
	}()
	r, err := Parse(c)
	if err == nil {
		t.Fatalf("Invalid string literal must raise an error but got %v", r)
	}
}

func TestTooLargeFloatLiteral(t *testing.T) {
	src := locerr.NewDummySource("1.7976931348623159e308")
	tokens := []token.Token{
		token.Token{
			Kind:  token.FLOAT,
			Start: locerr.Pos{0, 1, 1, src},
			End:   locerr.Pos{22, 1, 22, src},
			File:  src,
		},
		token.Token{
			Kind:  token.EOF,
			Start: locerr.Pos{22, 1, 22, src},
			End:   locerr.Pos{22, 1, 22, src},
			File:  src,
		},
	}
	c := make(chan token.Token)
	go func() {
		for _, t := range tokens {
			c <- t
		}
	}()
	r, err := Parse(c)
	if err == nil {
		t.Fatalf("Invalid int literal must raise an error but got %v", r)
	}
	if !strings.Contains(err.Error(), "value out of range") {
		t.Fatal("Unexpected error:", err)
	}
}
