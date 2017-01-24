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
