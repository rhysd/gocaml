package sema

import (
	"github.com/rhysd/gocaml/syntax"
	"github.com/rhysd/locerr"
	"strings"
	"testing"
)

func TestInferOK(t *testing.T) {
	codes := map[string]string{
		"let rec id x = x in id":                                                               "'a -> 'a",
		"let one = 1 in one":                                                                   "int",
		"let rec id x = x in let x = id in x":                                                  "'a -> 'a",
		"let x = fun y -> y in x":                                                              "'a -> 'a",
		"fun x -> x":                                                                           "'a -> 'a",
		"let rec pair x y = x, y in pair":                                                      "'a -> 'b -> ('a * 'b)",
		"fun x -> let y = fun z -> z in y":                                                     "'a -> ('b -> 'b)",
		"let rec f x = x in let rec pair x y = x, y in (pair (f 1) (f true))":                  "int * bool",
		"let rec f x y = let a = x = y in x = y in f":                                          "'a -> 'a -> bool",
		"let rec id x = x in (id id)":                                                          "'a -> 'a",
		"let rec choose a b = if true then a else b in (choose (fun x y -> x) (fun x y -> y))": "'a -> 'a -> 'a",
		"let rec id a = a in let x = id in let y = let z = x id in z in y":                     "'a -> 'a",
		"let rec f x = Array.make 3 x in let rec id x = x in f id":                             "('a -> 'a) array",
		// "let rec inc (x:int):int = x + 1 in let rec id x = x in let a = Array.make 3 id in a.(1) <- inc; a": "(int -> int) array",
		"let rec f x = let y = x in y in f":                                 "'a -> 'a",
		"let rec f x = let y = let z = x (fun x -> x) in z in y in f":       "(('a -> 'a) -> 'b) -> 'b",
		"let rec f x = let rec g y = let x = x y in x y in g in f":          "('a -> ('a -> 'b)) -> ('a -> 'b)",
		"let rec f x = let rec y z = x z in y in f":                         "('a -> 'b) -> ('a -> 'b)",
		"let rec f x = let rec y z = x in y in f":                           "'a -> ('b -> 'a)",
		"let rec f x = let rec g y = let x = x y in fun x -> y x in g in f": "(('a -> 'b) -> 'c) -> (('a -> 'b) -> ('a -> 'b))",
		"let rec f x = let rec y z = z in y y in f":                         "'a -> ('b -> 'b)",
		"let rec a f = let rec x g y = let _ = g(y) in true in x in a":      "'a -> (('b -> 'c) -> 'b -> bool)",
		"let rec const x = let rec f y = x in f in const":                   "'a -> ('b -> 'a)",
		"let rec apply f x = f x in apply":                                  "('a -> 'b) -> 'a -> 'b",
	}

	for code, want := range codes {
		src := locerr.NewDummySource(code)
		ast, err := syntax.Parse(src)
		if err != nil {
			t.Fatal(code, ": parse error:", err)
		}
		if err := AlphaTransform(ast); err != nil {
			t.Fatal(code, ": alpha trans error:", err)
		}
		i := NewInferer()
		// nodeTypeConv is unnecessary because no type annotation is contained in test cases
		ty, err := i.infer(ast.Root, 0)
		if err != nil {
			t.Fatal(code, ": inference error:", err)
		}
		ty, _ = generalize(ty, -1)
		have := ty.String()
		if have != want {
			t.Fatal(code, ": unexpected type:", have, ", wanted:", want)
		}
	}
}

func TestInferFail(t *testing.T) {
	codes := map[string]string{
		"let rec pair x y = x * y in let rec f g = pair (g 0) (g true) in f": "Type mismatch between 'int' and 'bool'",
		"let one = 1 in let rec add x y = x + y in add one true":             "Type mismatch between 'int' and 'bool'",
		"let one = 1 in let rec add x y = x + y in add one":                  "Number of parameters of function does not match: 2 vs 1",
		"fun x -> let y = x in y y":                                          "Cyclic dependency found",
		"fun x -> x x":                                                       "Cyclic dependency found",
		"let one = 1 in let rec id x = x in one id":                          "Cannot unify types",
	}

	for code, want := range codes {
		src := locerr.NewDummySource(code)
		ast, err := syntax.Parse(src)
		if err != nil {
			t.Fatal(code, ": parse error:", err)
		}
		if err := AlphaTransform(ast); err != nil {
			t.Fatal(code, ": alpha trans error:", err)
		}
		i := NewInferer()
		// nodeTypeConv is unnecessary because no type annotation is contained in test cases
		_, err = i.infer(ast.Root, 0)
		if err == nil {
			t.Fatal(code, "did not make an error")
		}
		have := err.Error()
		if !strings.Contains(have, want) {
			t.Error(code, "made unexpected error:", have)
		}
	}
}
