package parser

import (
	"fmt"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/token"
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
				s, err := token.NewSourceFromFile(n)
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

func TestParseInvalid(t *testing.T) {
	tokens := []token.Token{
		token.Token{
			Kind:  token.IF,
			Start: token.Position{0, 1, 1},
			End:   token.Position{2, 1, 3},
		},
		token.Token{
			Kind:  token.IF,
			Start: token.Position{3, 1, 4},
			End:   token.Position{5, 1, 6},
		},
		token.Token{
			Kind:  token.EOF,
			Start: token.Position{2, 1, 6},
			End:   token.Position{2, 1, 6},
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
	src := token.NewDummySource("123456789123456789123456789123456789123456789")
	tokens := []token.Token{
		token.Token{
			Kind:  token.INT,
			Start: token.Position{0, 1, 1},
			End:   token.Position{45, 1, 45},
			File:  src,
		},
		token.Token{
			Kind:  token.EOF,
			Start: token.Position{45, 1, 45},
			End:   token.Position{45, 1, 45},
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
	src := token.NewDummySource("\"a\nb\"\n")
	tokens := []token.Token{
		token.Token{
			Kind:  token.STRING_LITERAL,
			Start: token.Position{1, 1, 0},
			End:   token.Position{2, 3, 5},
			File:  src,
		},
		token.Token{
			Kind:  token.EOF,
			Start: token.Position{3, 1, 6},
			End:   token.Position{3, 1, 6},
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
	src := token.NewDummySource("1.7976931348623159e308")
	tokens := []token.Token{
		token.Token{
			Kind:  token.FLOAT,
			Start: token.Position{0, 1, 1},
			End:   token.Position{22, 1, 22},
			File:  src,
		},
		token.Token{
			Kind:  token.EOF,
			Start: token.Position{22, 1, 22},
			End:   token.Position{22, 1, 22},
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
