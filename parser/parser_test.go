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
		"../testdata/basic",
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
