package syntax

import (
	"fmt"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/locerr"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestLexingOK(t *testing.T) {
	for _, testdir := range []string{
		"testdata",
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
				s, err := locerr.NewSourceFromFile(n)
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

// List literal can be lexed but parser should complain that it is not implemented yet.
// This behavior is implemented because array literal ressembles to list literal.
func TestLexingListLiteral(t *testing.T) {
	s := locerr.NewDummySource("[1; 2; 3]")
	l := NewLexer(s)
	go l.Lex()
lexing:
	for {
		select {
		case tok := <-l.Tokens:
			switch tok.Kind {
			case token.ILLEGAL:
				t.Fatal(tok.String())
			case token.EOF:
				break lexing
			}
		}
	}
}

func TestLexingIllegal(t *testing.T) {
	testdir := filepath.FromSlash("testdata/lexer/invalid")
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
			s, err := locerr.NewSourceFromFile(n)
			if err != nil {
				panic(err)
			}
			errorOccurred := false
			l := NewLexer(s)
			l.Error = func(_ string, _ locerr.Pos) {
				errorOccurred = true
			}
			go l.Lex()
			for {
				select {
				case tok := <-l.Tokens:
					switch tok.Kind {
					case token.ILLEGAL:
						if !errorOccurred {
							t.Fatalf("Illegal token was emitted but no error occurred")
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
