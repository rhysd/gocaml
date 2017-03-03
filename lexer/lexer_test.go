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

			t.Run(fmt.Sprintf("Check lexing successfully: %s", n), func(t *testing.T) {
				s, err := token.NewSourceFromFile(n)
				if err != nil {
					panic(err)
				}
				l := NewLexer(s)
				go l.Lex()
				for {
					select {
					case tok := <-l.Tokens:
						switch tok.Kind {
						case token.ILLEGAL:
							t.Fatal(tok.String())
						case token.EOF:
							return
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
			l := NewLexer(s)
			l.Error = func(_ string, _ token.Position) {
				errorOccurred = true
			}
			go l.Lex()
			for {
				select {
				case tok := <-l.Tokens:
					switch tok.Kind {
					case token.ILLEGAL:
						if !errorOccurred {
							t.Fatalf("Illegal token was emitted but no error occured")
						}
						return
					case token.EOF:
						t.Fatalf("Lexing successfully done unexpectedly")
						return
					}
				}
			}
		})
	}
}
