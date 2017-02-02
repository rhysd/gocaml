package ast

import (
	"bytes"
	"github.com/rhysd/gocaml/token"
	"io"
	"os"
	"testing"
)

func TestPrintAST(t *testing.T) {
	s := token.NewDummySource("")
	tok := &token.Token{
		Kind:  token.ILLEGAL,
		Start: token.Position{0, 0, 0},
		End:   token.Position{0, 0, 0},
		File:  s,
	}

	root := &Let{
		tok,
		NewSymbol("foo"),
		&Add{
			&Sub{
				&FSub{
					&Less{
						&Unit{tok, tok},
						&Not{
							tok,
							&Bool{tok, true},
						},
					},
					&Neg{
						tok,
						&Int{tok, 42},
					},
				},
				&Eq{
					&FNeg{
						tok,
						&Float{tok, 3.14},
					},
					&VarRef{tok, NewSymbol("variable")},
				},
			},
			&FAdd{
				&FDiv{
					&Tuple{
						[]Expr{
							&Int{tok, 42},
							&Float{tok, 3.14},
						},
					},
					&ArrayCreate{
						tok,
						&Int{tok, 42},
						&Bool{tok, false},
					},
				},
				&FMul{
					&Get{
						&ArrayCreate{
							tok,
							&Int{tok, 42},
							&Bool{tok, false},
						},
						&Int{tok, 1},
					},
					&Put{
						&ArrayCreate{
							tok,
							&Int{tok, 42},
							&Bool{tok, false},
						},
						&Int{tok, 1},
						&Bool{tok, true},
					},
				},
			},
		},
		&LetTuple{
			tok,
			[]*Symbol{
				NewSymbol("a"),
				NewSymbol("b"),
			},
			&Tuple{
				[]Expr{
					&Int{tok, 42},
					&Float{tok, 3.14},
				},
			},
			&LetRec{
				tok,
				&FuncDef{
					NewSymbol("f"),
					[]*Symbol{
						NewSymbol("a"),
					},
					&VarRef{tok, NewSymbol("a")},
				},
				&If{
					tok,
					&Bool{tok, true},
					&Apply{
						&VarRef{tok, NewSymbol("f")},
						[]Expr{
							&Int{tok, 42},
						},
					},
					&Int{tok, 0},
				},
			},
		},
	}

	ast := &AST{
		Root: root,
		File: s,
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Println(ast)

	ch := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		ch <- buf.String()
	}()
	w.Close()
	os.Stdout = old

	expected := `AST for dummy:
-   Let (foo) (0:0-0:0)
-   -   Add (0:0-0:0)
-   -   -   Sub (0:0-0:0)
-   -   -   -   FSub (0:0-0:0)
-   -   -   -   -   Less (0:0-0:0)
-   -   -   -   -   -   Unit (0:0-0:0)
-   -   -   -   -   -   Not (0:0-0:0)
-   -   -   -   -   -   -   Bool (0:0-0:0)
-   -   -   -   -   Neg (0:0-0:0)
-   -   -   -   -   -   Int (0:0-0:0)
-   -   -   -   Eq (0:0-0:0)
-   -   -   -   -   FNeg (0:0-0:0)
-   -   -   -   -   -   Float (0:0-0:0)
-   -   -   -   -   VarRef (variable) (0:0-0:0)
-   -   -   FAdd (0:0-0:0)
-   -   -   -   FDiv (0:0-0:0)
-   -   -   -   -   Tuple (0:0-0:0)
-   -   -   -   -   -   Int (0:0-0:0)
-   -   -   -   -   -   Float (0:0-0:0)
-   -   -   -   -   ArrayCreate (0:0-0:0)
-   -   -   -   -   -   Int (0:0-0:0)
-   -   -   -   -   -   Bool (0:0-0:0)
-   -   -   -   FMul (0:0-0:0)
-   -   -   -   -   Get (0:0-0:0)
-   -   -   -   -   -   ArrayCreate (0:0-0:0)
-   -   -   -   -   -   -   Int (0:0-0:0)
-   -   -   -   -   -   -   Bool (0:0-0:0)
-   -   -   -   -   -   Int (0:0-0:0)
-   -   -   -   -   Put (0:0-0:0)
-   -   -   -   -   -   ArrayCreate (0:0-0:0)
-   -   -   -   -   -   -   Int (0:0-0:0)
-   -   -   -   -   -   -   Bool (0:0-0:0)
-   -   -   -   -   -   Int (0:0-0:0)
-   -   -   -   -   -   Bool (0:0-0:0)
-   -   LetTuple (a, b) (0:0-0:0)
-   -   -   Tuple (0:0-0:0)
-   -   -   -   Int (0:0-0:0)
-   -   -   -   Float (0:0-0:0)
-   -   -   LetRec (fun f a) (0:0-0:0)
-   -   -   -   VarRef (a) (0:0-0:0)
-   -   -   -   If (0:0-0:0)
-   -   -   -   -   Bool (0:0-0:0)
-   -   -   -   -   Apply (0:0-0:0)
-   -   -   -   -   -   VarRef (f) (0:0-0:0)
-   -   -   -   -   -   Int (0:0-0:0)
-   -   -   -   -   Int (0:0-0:0)
`
	actual := <-ch
	if expected != actual {
		t.Fatalf("Unexpected output from Println():\n%s", actual)
	}
}
