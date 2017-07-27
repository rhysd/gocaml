package closure

import (
	"bytes"
	"fmt"
	"github.com/rhysd/gocaml/sema"
	"github.com/rhysd/gocaml/syntax"
	"github.com/rhysd/locerr"
	"strings"
	"testing"
)

func TestClosureTransform(t *testing.T) {
	empty := map[string][]string{}
	cases := []struct {
		what     string
		code     string
		closures map[string][]string
		toplevel []string
		entry    []string
	}{
		{
			what:     "no function",
			code:     "let x = 42 in x",
			closures: empty,
			toplevel: []string{},
			entry: []string{
				"x$t1 = int 42 ; type=int",
			},
		},
		{
			what:     "no function with nested block",
			code:     "if true then 0 else 2",
			closures: empty,
			toplevel: []string{},
			entry: []string{
				"if $k1 ; type=int",
			},
		},
		{
			what:     "simple normal function",
			code:     "let rec f x = x in f 42",
			closures: empty,
			toplevel: []string{
				"f$t1 = fun x$t2 ; type=int -> int",
			},
			entry: []string{
				"app f$t1 $k3 ; type=int",
			},
		},
		{
			what:     "non-captured variable",
			code:     "let y = 0 in let rec f x = x in f y",
			closures: empty,
			toplevel: []string{
				"f$t2 = fun x$t3 ; type=int -> int",
			},
			entry: []string{
				"app f$t2 y$t1 ; type=int",
			},
		},
		{
			what: "simple closure",
			code: "let x = 42 in let rec f a = x + a in f 1",
			closures: map[string][]string{
				"f$t2": []string{"x$t1"},
			},
			toplevel: []string{
				"f$t2 = fun a$t3 ; type=int -> int",
			},
			entry: []string{
				"f$t2 = makecls (x$t1) f$t2 ; type=int -> int",
				"appcls f$t2 $k6 ; type=int",
			},
		},
		{
			what:     "normal function in nested block",
			code:     "(if true then 0 else let rec f a = a in f 10)",
			closures: empty,
			toplevel: []string{
				"f$t1 = fun a$t2 ; type=int -> int",
			},
			entry: []string{
				"BEGIN: else",
				"app f$t1 $k5 ; type=int",
			},
		},
		{
			what: "closure in nested block",
			code: "let x = 42 in (if true then 0 else let rec f a = a + x in f 10)",
			closures: map[string][]string{
				"f$t2": []string{"x$t1"},
			},
			toplevel: []string{
				"f$t2 = fun a$t3 ; type=int -> int",
			},
			entry: []string{
				"BEGIN: else",
				"f$t2 = makecls (x$t1) f$t2 ; type=int -> int",
				"appcls f$t2 $k8 ; type=int",
			},
		},
		{
			what: "capture var in block of fun",
			code: "let x = 10 in let rec f a = let rec g a = a in g x in f 0",
			closures: map[string][]string{
				"f$t2": []string{"x$t1"},
			},
			toplevel: []string{
				"f$t2 = fun a$t3 ; type=int -> int",
				"g$t4 = fun a$t5 ; type=int -> int",
				"app g$t4 x$t1 ; type=int",
			},
			entry: []string{
				"makecls (x$t1) f$t2 ; type=int -> int",
				"appcls f$t2 $k7 ; type=int",
			},
		},
		{
			what:     "nested functions",
			code:     "let rec f x = let rec g x = x in g x in f 42",
			closures: empty,
			toplevel: []string{
				"f$t1 = fun x$t2 ; type=int -> int",
				"g$t3 = fun x$t4 ; type=int -> int",
				"app g$t3 x$t2 ; type=int",
			},
			entry: []string{
				"app f$t1 $k6",
			},
		},
		{
			what: "nested closures",
			code: "let a = 1 in let b = 1 in let rec f x = let rec g x = x + a in g  x + b in f 42",
			closures: map[string][]string{
				"f$t3": []string{"a$t1", "b$t2"},
				"g$t5": []string{"a$t1"},
			},
			toplevel: []string{
				"f$t3 = fun x$t4 ; type=int -> int",
				"g$t5 = fun x$t6 ; type=int -> int",
				"makecls (a$t1) g$t5 ; type=int -> int",
				"appcls g$t5 x$t4 ; type=int",
			},
			entry: []string{
				"makecls (a$t1,b$t2) f$t3 ; type=int -> int",
				"appcls f$t3 $k12 ; type=int",
			},
		},
		{
			what:     "recursive function",
			code:     "let rec f x = if x < 0 then 1 else x + f (x-1) in f 10",
			closures: empty,
			toplevel: []string{
				"f$t1 = recfun x$t2 ; type=int -> int",
				"app f$t1",
			},
			entry: []string{
				"app f$t1",
			},
		},
		{
			what: "recursive closure",
			code: "let a = 42 in let rec f x = if x < 0 then 1 else a + x + f (x-1) in f 10",
			closures: map[string][]string{
				"f$t2": []string{"a$t1"},
			},
			toplevel: []string{
				"f$t2 = recfun x$t3 ; type=int -> int",
				"appcls f$t2",
			},
			entry: []string{
				"makecls (a$t1) f$t2",
				"appcls f$t2",
			},
		},
		{
			what: "multiple closures",
			code: "let a = 42 in let b = 42 in let rec f x = a + x in let rec g x = b + x in f (g a)",
			closures: map[string][]string{
				"f$t3": []string{"a$t1"},
				"g$t5": []string{"b$t2"},
			},
			toplevel: []string{
				"f$t3 = fun x$t4 ; type=int -> int",
				"g$t5 = fun x$t6 ; type=int -> int",
			},
			entry: []string{
				"makecls (a$t1) f$t3",
				"makecls (b$t2) g$t5",
				"appcls g$t5",
				"appcls f$t3",
			},
		},
		{
			what: "multiple closures",
			code: "let a = 42 in let b = 42 in let rec f x = a + x in let rec g x = b + x in f (g a)",
			closures: map[string][]string{
				"f$t3": []string{"a$t1"},
				"g$t5": []string{"b$t2"},
			},
			toplevel: []string{
				"f$t3 = fun x$t4 ; type=int -> int",
				"g$t5 = fun x$t6 ; type=int -> int",
			},
			entry: []string{
				"makecls (a$t1) f$t3",
				"makecls (b$t2) g$t5",
				"appcls g$t5",
				"appcls f$t3",
			},
		},
		{
			what: "returned function as variable",
			code: "let rec f x = x in let rec g x = f in (g ()) 42",
			closures: map[string][]string{
				// Need to be closure because returned value from `g ()` can't be determined function or closure.
				"f$t1": []string{},
				"g$t3": []string{"f$t1"},
			},
			toplevel: []string{
				"f$t1 = fun x$t2",
				"g$t3 = fun x$t4",
				"ref f$t1 ; type=int -> int",
			},
			entry: []string{
				"makecls () f$t1",
				"makecls (f$t1) g$t3",
				"appcls g$t3",
				"appcls $k",
			},
		},
		{
			what: "returned function is also referred in function body",
			code: "let rec f x = x in let rec g x = (f 10; f) in (g ()) 42",
			closures: map[string][]string{
				// Need to be closure because returned value from `g ()` can't be determined function or closure.
				"f$t1": []string{},
				"g$t3": []string{"f$t1"},
			},
			toplevel: []string{
				"f$t1 = fun x$t2",
				"g$t3 = fun x$t4",
				"appcls f$t1",
				"ref f$t1 ; type=int -> int",
			},
			entry: []string{
				"makecls () f$t1",
				"makecls (f$t1) g$t3",
				"appcls g$t3",
				"appcls $k",
			},
		},
		{
			what: "returned closure as variable",
			code: "let a = 10 in let rec f x = a + x in let rec g x = f in (g ()) 42",
			closures: map[string][]string{
				// Need to be closure because returned value from `g ()` can't be determined function or closure.
				"f$t2": []string{"a$t1"},
				"g$t4": []string{"f$t2"},
			},
			toplevel: []string{
				"f$t2 = fun x$t3",
				"g$t4 = fun x$t5",
				"ref f$t2 ; type=int -> int",
			},
			entry: []string{
				"makecls (a$t1) f$t2",
				"makecls (f$t2) g$t4",
				"appcls g$t4",
				"appcls $k",
			},
		},
		{
			what:     "external function call",
			code:     "print_int 42",
			closures: empty,
			toplevel: []string{},
			entry: []string{
				"appx print_int $k2",
			},
		},
		{
			what: "external function call in closure body",
			code: "let x = 42 in let rec f a = (print_int (a + x); ()) in f 10",
			closures: map[string][]string{
				"f$t2": []string{"x$t1"},
			},
			toplevel: []string{
				"f$t2 = fun a$t3",
				"appx print_int $k",
			},
			entry: []string{
				"makecls (x$t1) f$t2",
				"appcls f$t2 $k",
			},
		},
		{
			// This is a test for deep-copy of transform visitor instance
			what: "sibling closures",
			code: "let a = 42 in let rec f x = let rec g y = y + a in g x in let rec h p = a - p in f (h 1)",
			closures: map[string][]string{
				"f$t2": []string{"a$t1"},
				"g$t4": []string{"a$t1"},
				"h$t6": []string{"a$t1"},
			},
		},
		{
			what: "option values",
			code: "let o = None in let rec f x = match o with Some i -> i | None -> 42 in f (Some 13); f o",
			closures: map[string][]string{
				"f$t2": []string{"o$t1"},
			},
			toplevel: []string{},
			entry: []string{
				"none ; type=int option",
			},
		},
		{
			what: "capture match var",
			code: "(match Some 42 with Some i -> let rec f x = x + i in f | None -> let rec f x = x * 2 in f)",
			closures: map[string][]string{
				"f$t2": []string{"i$t1"},
				"f$t4": []string{},
			},
			toplevel: []string{},
			entry: []string{
				"some $k1 ; type=int option",
				"issome $k2 ; type=bool",
				"derefsome $k2 ; type=int",
			},
		},
		{
			what: "capture in array literal",
			code: "let a = 42 in let rec f x = [| a; x |] in f 3",
			closures: map[string][]string{
				"f$t2": []string{"a$t1"},
			},
			toplevel: []string{},
			entry: []string{
				"int 42 ; type=int",
				"makecls (a$t1) f$t2 ; type=int -> int array",
				"int 3 ; type=int",
				"appcls f$t2 $k6 ; type=int array",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.what, func(t *testing.T) {
			s := locerr.NewDummySource(fmt.Sprintf("%s; ()", tc.code))
			ast, err := syntax.Parse(s)
			if err != nil {
				t.Fatal(err)
			}
			env, ir, err := sema.SemanticsCheck(ast)
			if err != nil {
				t.Fatal(err)
			}
			prog := Transform(ir)

			if len(tc.closures) != len(prog.Closures) {
				t.Fatalf("Expected %d closures but %d closures found: %v", len(tc.closures), len(prog.Closures), prog.Closures)
			}
			for f, expected := range tc.closures {
				actual, ok := prog.Closures[f]
				if !ok {
					t.Errorf("Function '%s' was expected to be captured: %v", f, prog.Closures)
					continue
				}
				if len(actual) != len(expected) {
					t.Fatalf("Function '%s' should have %d captures but actually have %d captures", f, len(expected), len(actual))
				}
				for i, e := range expected {
					a := actual[i]
					if e != a {
						t.Errorf("Expected %dth capture of function '%s' as '%s' but actually '%s'", i, f, e, a)
					}
				}
			}
			for f, expected := range prog.Closures {
				fv, ok := prog.Closures[f]
				if !ok {
					t.Errorf("Function '%s' must be in closures but not found", f)
					continue
				}
				if len(expected) != len(fv) {
					t.Fatalf("%d free variables are expected but %d ones found", len(expected), len(fv))
					continue
				}
				for i, v := range fv {
					if expected[i] != v {
						t.Errorf("%dth free variable mismatched: %s v.s. %s", i, expected, v)
					}
				}
			}

			var buf bytes.Buffer
			prog.PrintToplevels(&buf, env)
			out := buf.String()
			for _, expected := range tc.toplevel {
				if !strings.Contains(out, expected) {
					t.Errorf("Expected '%s' to be contained in toplevel '%s'", expected, out)
				}
			}

			buf.Reset()
			prog.Entry.Println(&buf, env)
			out = buf.String()
			for _, expected := range tc.entry {
				if !strings.Contains(out, expected) {
					t.Errorf("Expected '%s' to be contained in entry '%s'", expected, out)
				}
			}
		})
	}
}

func TestClosureCaptureInInsn(t *testing.T) {
	code := `
	let a = 0 in
	let b = 1 in
	let c = 1.0 in
	let d = 3.14 in
	let e = false in
	let f = 3 in
	let g = 4 in
	let h = 1 in
	let i = 0 in
	let j = 42 in
	let k = 3.14 in
	let l = true in
	let m = 0 in
	let n = -1 in
	let o = 1.0 in
	let p = 3.1 in
	let q = true in
	let r = false in
	let rec func x =
		a + b;
		-a;
		-.c;
		c *. d;
		m < n;
		o <> p;
		q = r;
		(if e then
			let arr = Array.make f g in
			Array.length arr;
			arr.(h);
			arr.(i) <- j
		else
			let tpl = (k, l) in
			let (y, z) = tpl in
			k +. y /. x;
			()
		)
	in
		func 3.14
	`

	expected := map[string]struct{}{}
	for _, c := range []string{
		"a$t1",
		"b$t2",
		"c$t3",
		"d$t4",
		"e$t5",
		"f$t6",
		"g$t7",
		"h$t8",
		"i$t9",
		"j$t10",
		"k$t11",
		"l$t12",
		"m$t13",
		"n$t14",
		"o$t15",
		"p$t16",
		"q$t17",
		"r$t18",
	} {
		expected[c] = struct{}{}
	}

	s := locerr.NewDummySource(code)
	ast, err := syntax.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	_, ir, err := sema.SemanticsCheck(ast)
	if err != nil {
		t.Fatal(err)
	}
	prog := Transform(ir)

	c, ok := prog.Closures["func$t19"]
	if !ok {
		t.Fatalf("No capture found for function 'f': %v", prog.Closures)
	}
	if len(c) != len(expected) {
		t.Errorf("%d captures was expected but actually %d found", len(expected), len(c))
	}
	for _, n := range c {
		if _, ok := expected[n]; !ok {
			t.Errorf("Variable '%s' was expected to be captured: %v", n, c)
		}
	}
}
