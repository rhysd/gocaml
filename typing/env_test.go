package typing

import (
	"fmt"
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestTypeCheckOK(t *testing.T) {
	testdir := filepath.FromSlash("../testdata/from-mincaml/")
	files, err := ioutil.ReadDir(testdir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		n := filepath.Join(testdir, f.Name())
		if !strings.HasSuffix(n, ".ml") {
			continue
		}

		t.Run(fmt.Sprintf("Infer types successfully: %s", n), func(t *testing.T) {
			s, err := token.NewSourceFromFile(n)
			if err != nil {
				panic(err)
			}

			l := lexer.NewLexer(s)
			go l.Lex()

			root, err := parser.Parse(l.Tokens)
			if err != nil {
				panic(root)
			}

			if err = alpha.Transform(root); err != nil {
				panic(err)
			}

			env := NewEnv()
			if err := env.ApplyTypeAnalysis(root); err != nil {
				t.Fatal(err)
			}
		})
	}
}
