package mono

import (
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
)

// Monomorphization makes polymorphic instructions into monomorphic ones.
// Polymorphic functions are duplicated for each instantiations. And polymorphic (not function) variables
// remain as-is.
//
// - Before:
//
// ```
// id$t1 = fun x$t2     ; type='a -> 'a
//   BEGIN body (id$t1)
//   $k1 = ref x$t2     ; type='a
//   END body (id$t1)
// ```
//
// - After:
//
// ```
// id$t1$int = fun x$t2$int     ; type=int -> int
//   BEGIN body (id$t1$int)
//   $k1 = ref x$t2$int     ; type=int
//   END body (id$t1$int)
//
// id$t1$bool = fun x$t2$bool     ; type=bool -> bool
//   BEGIN body (id$t1$bool)
//   $k1 = ref x$t2$bool     ; type=bool
//   END body (id$t1$bool)
//
// id$t1$unit = fun x$t2$unit     ; type=unit -> unit
//   BEGIN body (id$t1$unit)
//   $k1 = ref x$t2$unit     ; type=unit
//   END body (id$t1$unit)
// ```
//
// How polymorphic functions were instantiated is recorded as env.PolyVariants in *types.Env.
//
// Monomorphization is performed in following order:
// 1. Monomorphize the expression of entry point of program
// 2. Monomorphize non-closure functions. Each function is duplicated for each instantiations. For
//    non-polymorphic functions, they don't need to be duplicated.
// 3. Closure functions are monomorphized at makecls instruction because captures may contain polymorphic
//    type values.
//
// The reason of 3. is that captures of a closure function is passed as a hidden parameter of function.
// Monomorphization must consider the case where the hidden parameter is polymorphic.
// For example of 3., The function `f` is a polymorphic function 'a -> (). And `g` is also polymorphic
// 'b -> 'a * 'b. In this case `g` captures `x`, which is typed as 'a. So `g` need to consider the captured
// type 'a on monomorphization. `g` should be duplicated for each `x` types (int and bool). It means that
//
// ```
// let rec f x =
//   let rec g y = x, y in
//   g 3.14; g "foo"; ()
// in
// f 10; f true
// ```
//
// `g` should be duplicated into 4 functions `g$int$float`, `g$int$string`, `g$bool$float` and `g$bool$string`.
//
// ```
// let rec f$int x$int =
//   let rec g$int$float y$float = x$int, y$float in
//   let rec g$int$string y$string = x$int, y$string in
//   g$int$float 3.14; g$int$string "foo"; ()
// in
// let rec f$bool x$bool =
//   let rec g$bool$float y$float = x$bool, y$float in
//   let rec g$bool$string y$string = x$bool, y$string in
//   g$bool$float 3.14; g$bool$string "foo"; ()
// in
// f$int 10; f$bool true
// ```
//
// However, all the type variables are not instantiated in function definitions. There are some cases
// where variables are polymorphic; Type is undetermined at variable definition at `let` expression.
// This case can be split into 2 cases:
//
// 1. not function
// 2. function
//
// For 1., we leave the type variable as-is because it is not important which type it is actually
// instantiated.
//
// ```
// let o = None in Some 42 = o; Some true = o; ()
// ```
//
// This program is valid. `o` in `Some 42 = o` is instantiated as `int option` and `Some true = o`
// is instantiated as `bool option`. However, `o` in `let` expression is `'a option`. But this is
// not important because it's `None`. Although this is a case of option type, other types are the same.
//
// For 2., actual function pointer is determined at the function call. So we create a table of instantiations
// and set the pointer to the table instead of the function pointer when it is polymorphic function
// variable.
//
// ```
// let rec id x = id in
// let f = id in
// f 42;
// f ()
// ```
//
// In above example, we create a list of instantiations of `id`; `let id_ptr = [id$int; id$unit]`.
// Then the pointer to the table is embedded in closure object instead of function pointer because
// `id` is a polymorphic function. Function variable `f` can be assigned correctly.
// At invocation of `f`, we have list of instantiations statically so it's possible to get the instantiated
// function pointer from the list in the closure object by index. Finally we can call it with captures.
//
// All instructions in polymorphic functions' bodies are duplicated. We need to consider the duplications
// don't break alpha transformation. We introduce another ID counter to solve this.
// All created instructions due to monomorphization will have a new name with counter. If the counter
// value is `42`, the instruction monomorphized from `foo$t3` will be named as `foo$t3$42`.

func Monomorphize(prog *mir.Program, env *types.Env) {
	// TODO
}
