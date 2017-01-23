package lexer

import (
	"fmt"
	"github.com/rhysd/gocaml/token"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestLexingOK(t *testing.T) {
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

			t.Run(fmt.Sprintf("Check lexing successfully: %s", n), func(t *testing.T) {
				s, err := token.NewSourceFromFile(n)
				if err != nil {
					panic(err)
				}
				tokens := make(chan token.Token)
				l := NewLexer(s, tokens)
				go l.Lex()
				for {
					select {
					case tok := <-tokens:
						switch tok.Kind {
						case token.ILLEGAL:
							t.Fatal(tok)
						case token.EOF:
							return
						default:
							break
						}
					}
				}
			})
		}
	}
}

func TestLexingIllegal(t *testing.T) {
	testdir := filepath.FromSlash("../testdata/lexer/invalid")
	files, err := ioutil.ReadDir(testdir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		n := filepath.Join(testdir, f.Name())
		if !strings.HasSuffix(n, ".ml") {
			continue
		}

		t.Run(fmt.Sprintf("Check lexing illegal input: %s", f.Name()), func(t *testing.T) {
			s, err := token.NewSourceFromFile(n)
			if err != nil {
				panic(err)
			}
			errorOccurred := false
			tokens := make(chan token.Token)
			l := NewLexer(s, tokens)
			l.Error = func(_ string, _ token.Position) {
				errorOccurred = true
			}
			go l.Lex()
			for {
				select {
				case tok := <-tokens:
					switch tok.Kind {
					case token.ILLEGAL:
						if !errorOccurred {
							t.Fatalf("Illegal token was emitted but no error occured")
						}
						return
					case token.EOF:
						t.Fatalf("Lexing successfully done unexpectedly")
						return
					default:
						break
					}
				}
			}
		})
	}
}
