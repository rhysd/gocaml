package sema

import (
	"github.com/rhysd/gocaml/syntax"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
	"strings"
	"testing"
)

func testInferCodeWithoutConv(code string) (types.Type, error) {
	src := locerr.NewDummySource(code)
	ast, err := syntax.Parse(src)
	if err != nil {
		return nil, err
	}
	env := types.NewEnv()
	if err := AlphaTransform(ast, env); err != nil {
		return nil, err
	}
	i := NewInferer(env)
	// nodeTypeConv is unnecessary because no type annotation is contained in test cases
	return i.infer(ast.Root, 0)
}

func TestInferAlgoWOK(t *testing.T) {
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
		"let rec f x = let y = x in y in f":                                                    "'a -> 'a",
		"let rec f x = let y = let z = x (fun x -> x) in z in y in f":                          "(('a -> 'a) -> 'b) -> 'b",
		"let rec f x = let rec g y = let x = x y in x y in g in f":                             "('a -> ('a -> 'b)) -> ('a -> 'b)",
		"let rec f x = let rec y z = x z in y in f":                                            "('a -> 'b) -> ('a -> 'b)",
		"let rec f x = let rec y z = x in y in f":                                              "'a -> ('b -> 'a)",
		"let rec f x = let rec g y = let x = x y in fun x -> y x in g in f":                    "(('a -> 'b) -> 'c) -> (('a -> 'b) -> ('a -> 'b))",
		"let rec f x = let rec y z = z in y y in f":                                            "'a -> ('b -> 'b)",
		"let rec a f = let rec x g y = let _ = g(y) in true in x in a":                         "'a -> (('b -> 'c) -> 'b -> bool)",
		"let rec const x = let rec f y = x in f in const":                                      "'a -> ('b -> 'a)",
		"let rec apply f x = f x in apply":                                                     "('a -> 'b) -> 'a -> 'b",
	}

	for code, want := range codes {
		ty, err := testInferCodeWithoutConv(code)
		if err != nil {
			t.Error(code, ": inference error:", err)
			continue
		}
		ty, _ = generalize(ty, -1)
		have := ty.String()
		if have != want {
			t.Error(code, ": unexpected type:", have, ", wanted:", want)
		}
	}
}

func TestInferAlgoWFail(t *testing.T) {
	codes := map[string]string{
		"let rec pair x y = x * y in let rec f g = pair (g 0) (g true) in f": "Type mismatch between 'int' and 'bool'",
		"let one = 1 in let rec add x y = x + y in add one true":             "Type mismatch between 'int' and 'bool'",
		"let one = 1 in let rec add x y = x + y in add one":                  "Number of parameters of function does not match: 2 vs 1",
		"fun x -> let y = x in y y":                                          "Cyclic dependency found",
		"fun x -> x x":                                                       "Cyclic dependency found",
		"let one = 1 in let rec id x = x in one id":                          "Cannot unify types",
	}

	for code, want := range codes {
		_, err := testInferCodeWithoutConv(code)
		if err == nil {
			t.Error(code, "did not make an error")
			continue
		}
		have := err.Error()
		if !strings.Contains(have, want) {
			t.Error(code, "made unexpected error:", have)
		}
	}
}

// Test cases from https://www.fos.kuis.kyoto-u.ac.jp/~igarashi/class/isle4-11w/testcases.html
func TestInferIgarashiOK(t *testing.T) {
	codes := map[string]string{
		"1 + 2":                                                                   "int",
		"-2 * 2":                                                                  "int",
		"1 < 2":                                                                   "bool",
		"fun x -> x":                                                              "'a -> 'a",
		"fun x -> fun y -> x":                                                     "'a -> ('b -> 'a)",
		"fun x -> fun y -> y":                                                     "'a -> ('b -> 'b)",
		"(fun x -> x + 1) 2 + (fun x -> x + -1) 3":                                "int",
		"fun f -> fun g -> fun x -> g (f x)":                                      "('a -> 'b) -> (('b -> 'c) -> ('a -> 'c))",
		"fun x -> fun y -> fun z -> x z (y z)":                                    "('a -> 'b -> 'c) -> (('a -> 'b) -> ('a -> 'c))",
		"fun x -> let y = x + 1 in x":                                             "int -> int",
		"fun x -> let y = x + 1 in y":                                             "int -> int",
		"fun b -> fun x -> if x b then x else (fun x -> b)":                       "bool -> ((bool -> bool) -> (bool -> bool))",
		"fun x -> if true then x else (if x then true else false)":                "bool -> bool",
		"fun x -> fun y -> if x then x else y":                                    "bool -> (bool -> bool)",
		"fun n -> (fun x -> x (fun y -> y)) (fun f -> f n)":                       "'a -> 'a",
		"fun x -> fun y -> x y":                                                   "('a -> 'b) -> ('a -> 'b)",
		"fun x -> fun y -> x (y x)":                                               "('a -> 'b) -> ((('a -> 'b) -> 'a) -> 'b)",
		"fun x -> fun y -> x (y x) (y x)":                                         "('a -> 'a -> 'b) -> ((('a -> 'a -> 'b) -> 'a) -> 'b)",
		"let id = fun x -> x in let f = fun y -> id (y id) in f":                  "(('a -> 'a) -> 'b) -> 'b",
		"let k = fun x -> fun y -> x in let k1 = fun x -> fun y -> k (x k) in k1": "(('a -> ('b -> 'a)) -> 'c) -> ('d -> ('e -> 'c))",
		"let g = fun h -> fun t -> fun f -> fun x -> f h (t f x) in g":            "'a -> ((('a -> 'b -> 'c) -> 'd -> 'b) -> (('a -> 'b -> 'c) -> ('d -> 'c)))",

		// Test cases needs let-polymerphism
		"let s = fun x -> fun y -> fun z -> x z (y z) in let s1 = fun x -> fun y -> fun z -> x s (z s) (y s (z s)) in s1":           "((('a -> 'b -> 'c) -> (('a -> 'b) -> ('a -> 'c))) -> 'd -> 'e -> 'f) -> (((('g -> 'h -> 'i) -> (('g -> 'h) -> ('g -> 'i))) -> 'd -> 'e) -> (((('j -> 'k -> 'l) -> (('j -> 'k) -> ('j -> 'l))) -> 'd) -> 'f))",
		"let f = fun x -> x in if f true then f 1 else f 2":                                                                         "int",
		"let f = fun x -> 3 in f true + f 4":                                                                                        "int",
		"fun b -> let f = fun x -> x in let g = fun y -> y in if b then f g else g f":                                               "bool -> ('a -> 'a)",
		"let s = fun x -> fun y -> fun z -> (x z) (y z) in let k = fun x -> fun y -> x in (s k) k":                                  "'a -> 'a",
		"let s = fun x -> fun y -> fun z -> (x z) (y z) in let k = fun x -> fun y -> x in let k_ = fun x -> fun y -> x in (s k) k_": "'a -> 'a",
		"let s = fun x -> fun y -> fun z -> (x z) (y z) in let k_ = fun x -> fun y -> y in (s k_) k_":                               "'a -> ('b -> 'b)",

		// Recursive functions
		"let rec f x = f x in f":                                                                 "'a -> 'b",
		"let rec f x = f (f x) in f":                                                             "'a -> 'a",
		"let rec fix_fun g = fun x -> g (fix_fun g) x in fix_fun":                                "(('a -> 'b) -> 'a -> 'b) -> ('a -> 'b)",
		"fun f -> let rec x z = f (x z) in x 666":                                                "('a -> 'a) -> 'a",
		"let rec f x y = if x < 0 then y else f (x + -1) y in f":                                 "int -> 'a -> 'a",
		"fun f -> fun g -> let rec h x = h (g (f x)) in h":                                       "('a -> 'b) -> (('b -> 'a) -> ('a -> 'c))",
		"let rec loop f = fun x -> (loop f) (f x) in loop":                                       "('a -> 'a) -> ('a -> 'b)",
		"let rec ind x = fun f -> fun n -> if n < 1 then x else f (((ind x) f) (n + -1)) in ind": "'a -> (('a -> 'a) -> (int -> 'a))",
	}
	for code, want := range codes {
		ty, err := testInferCodeWithoutConv(code)
		if err != nil {
			t.Error(code, "caused an error:", err)
			continue
		}
		ty, _ = generalize(ty, -1)
		have := ty.String()
		if have != want {
			t.Error(code, ": unexpected type:", have, ", wanted:", want)
		}
	}
}

func TestInferIgarashiFail(t *testing.T) {
	codes := []string{
		"1 + true",
		"2 + (fun x -> x)",
		"-2 * false",
		"fun x -> x x;;",
		"let f = fun x -> fun g -> g ((x x) g) in f f",
		"let g = fun f -> fun x -> f x (f x) in g",
		"let g = fun f -> fun x -> f x (x f) in g",
		"fun x -> fun y -> x y + y x",
		"fun x -> fun y -> x y + x",
		"fun x -> fun y -> if x y then x else y",
		"fun x -> fun y -> if x y then (fun z -> if y z then z else x) else (fun x -> x)",
		"fun x -> fun y -> fun z -> let b = (x y) z in if b then z y else z x",
		"fun x -> fun y -> fun z -> if x y then z x else y z",
		"fun x -> if x then 1 else x",
		"(fun x -> x + 1) true",
		"fun x -> fun y -> y (x (y x))",
		"(fun f -> fun x -> f (f x)) (fun x -> fun y -> x)",
		"fun x -> fun y -> y (x (fun z1 -> fun z2 -> z1)) (x (fun z -> z))",
		"fun b -> fun f -> let g1 = fun x -> f x in let g2 = fun x -> f x in if b then g1 g2 else g2 g1",

		// Recursive function errors
		"let rec f = fun x -> f in f",
		"et rec looq = fun f -> fun x -> (looq f) (x f) in looq",
		"let rec f = fun x -> f (x f) in f",
		"let rec f = fun z -> f z (fun g -> fun h -> h (g h)) in f",
	}

	for _, code := range codes {
		_, err := testInferCodeWithoutConv(code)
		if err == nil {
			t.Error(code, "should raise an error")
		}
	}
}
