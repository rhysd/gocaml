GoCaml :camel:
==============
[![Linux and macOS Build Status][]][Travis CI]
[![Windows Build Status][]][Appveyor]
[![Coverage Status][]][Coveralls]

GoCaml is a [MinCaml][] implementation in Go using [LLVM][]. MinCaml is a minimal subset of OCaml for educational purpose. It is statically-typed and compiled into a binary. ([spec][MinCaml spec])

This project aims my practices for understanding type inference, closure transform and introducing own intermediate language (IL) to own language.

Example:

```ocaml
let rec gcd m n =
  if m = 0 then n else
  if m <= n then gcd m (n - m) else
  gcd n (m - n) in
print_int (gcd 21600 337500)
```

You can see [more examples][examples]. (e.g. [Brainfxxk interpreter][Brainfxxk interpreter example])

## Tasks

- [x] Lexer -> ([doc][lexer doc])
- [x] Parser with [goyacc][] -> ([doc][parser doc])
- [x] Alpha transform ([doc][alpha transform doc])
- [x] Type inference (Hindley Milner monomorphic type system) -> ([doc][typing doc])
- [x] GoCaml intermediate language (GCIL) ([doc][gcil doc])
- [x] K normalization from AST into GCIL ([doc][gcil doc])
- [x] Closure transform ([doc][closure doc])
- [x] Code generation (LLVM IR, assembly, object, executable) using [LLVM][] ([doc][codegen doc])
- [x] LLVM IR level optimization passes
- [x] Garbage collection with [Boehm GC][]
- [x] Debug information (DWARF) using LLVM's Debug Info builder

## Difference from original MinCaml

- MinCaml assumes external symbols' types are `int` when it can't be inferred. GoCaml does not have such an assumption.
  GoCaml assumes unknown return type of external functions as `()` (`void` in C), but in other cases, falls into compilation error.
  When you use nested external functions call, you need to clarify the return type of inner function call. For example, when `f` in
  `g (f ())` returns `int`, you need to show it like `g ((f ()) + 0)`. Note that this pitfall does not occur for built-in functions
  because a compiler knows their types.
- MinCaml allows `-` unary operator for float literal. So for example `-3.14` is valid but `-f` (where `f` is `float`) is not valid.
  GoCaml does not allow `-` unary operator for float values totally. You need to use `-.` unary operator instead (e.g. `-.3.14`).
- GoCaml adds more operators. `*` and `/` for integers, `&&` and `||` for booleans.
- GoCaml has string type. String value is immutable and used with slices.
- GoCaml does not have `Array.create`, which is an alias to `Array.make`. `Array.length` is available to obtain the size of array.
- Some useful built-in functions are added (described in below section).

## Language Spec

<details>
<summary>Program</summary>

Program is represented as one expression which MUST be evaluated as unit type. So `()` is the smallest program for GoCaml.

</details>

<details>
<summary>Sequence Expression</summary>

Sequenced program can be represented by joining multiple expressons with `;`.

```ml
e1; e2; e3; e4
```

In above program, expressions are evaluated in order of `e1 -> e2 -> e3 -> e4` and the sequenced expression is evaluated to the value of `e4`.
Program must be evaluated to unit type, so the `e4` expression must be evaluated to `()` (unit value).

</details>

<details>
<summary>Comments</summary>

There is a block comment syntax. It starts with `(*` and ends with `*)`. Any comment must be closed with `*)`, otherwise it falls into syntax error.

```ml
(*
   This is comment.
*)
```

</details>

<details>
<summary>Constants</summary>

There are unit, integer, boolean, float and string constants.

```ml
(* integer *)
42;

(* float *)
3.0;
3.14e+10;
3.14e-10;
1.;

(* boolean *)
true;
false;

(* string *)
"hello, world";
"contains\tescapes\n";

(* only one constant which is typed to unit *)
()
```

</details>

<details>
<summary>Show values</summary>

`print_*` and `println_*` built-in functions are available to output values to stdout.

```ml
print_int 42;
println_bool true
```

Please see 'Built-in functions' section below for more detail.

</details>

<details>
<summary>Unary operators</summary>

You can use some unary prefixed operators.

```ml
-42;

(* GoCaml distinguishes float and int in operators. -. is a float version of - *)
-.3.14;

not true;

()
```

</details>

<details>
<summary>Arithmetic binary operators</summary>

As mentioned above, GoCaml distinguishes int and float in operators. Operators for float values are suffixed by `.` (dot).

```ml
(* integer calcuration *)
1 + 2;
1 - 2;
1 * 2;
1 / 2;

(* float calcuration *)
1.0 + 2.0;
1.0 - 2.0;
1.0 * 2.0;
1.0 / 2.0;

()
```

Integer operators must have integer values as their operands. And float operators must have float values as their operands.
There is no implicit conversion. You need to convert explicitly by using built-in functions (e.g. `3.14 +. (int_to_float 42)`).

</details>

<details>
<summary>Relational operators</summary>

Equal operator is `=` (NOT `==`), Not-equal operator is `<>`. Compare operators are the same as C (`<`, `<=`, `>` and `>=`).

```ml
42 = 42; (* => true *)
42 <> 42; (* => false *)

3.14 > 2.0;
1.0 < 3.0;
2.0 >= 2.0;
1.0 <= 3.0;

()
```

Tuples (described below) and strings can be compared with `=` or `<>`, but cannot be compared with `<`, `<=`, `>` and `>=`.
Arrays (described below) cannot be compared directly with any compare operators. You need to compare each element explicitly.

</details>

<details>
<summary>Logical operators</summary>

`&&` and `||` are available for boolean values.

```ml
println_bool (true || false && false || false)
```

</details>

<details>
<summary>Variable</summary>

`let` expression binds some value to a variable.

```ml
(* 42 is bound to a *)
let a = 42 in
(* You can use the variable as value in any expression *)
a + 4;

()
```

The syntax is `let {name} = {e1} in {e2}`. `e2` will be evaluated where `e1` is bound to `name`. By chain `let`, you can define multiple variables.

```ml
let pi = 3.14.1592 in
let r = 2 in
let area = r *. r *. pi in
print_float area
```

And you can redefine the same name variable as already-defined ones.

```ml
let a = 42 in
println_int a;
let a = true in
println_bool a
```

The first `a` and the second `a` are different variable. Second one just shadows first one.
So you can always redefine any variable names as any type. Shadowed variable can be no longer referred.

Functions are first-class object in GoCaml. So you can also bind functions to variable as value.

```ml
let rec hello _ = println_str "hello" in
let f = hello in

(* Shows "helllo" *)
f ();

(* Binds external function *)
let p = println_str in

(* Shows "hi" *)
p "hi"
```

</details>

<details>
<summary>Functions</summary>

`let rec` is a keyword to define a function. Syntax is `let rec name params... = e1 in e2` where function `name` is defined as `e1` and then `e2` will be evaluated.
`f a b c` is an expression to apply function `f` with argument `a`, `b` and `c`.
As long as the argument is simple, you don't need to use `()`.

Note that, if you use some complecated expression (for example, binary operators), you need to use `()` like
`f (a+b) c`. If you specify `f a + b c`, it would be parsed as `(f a) + (b c)`.

```ml
let rec f a b c = a + b + c in
let d = f 10 20 30 in

(* Output: 60 *)
println_int d
```

You can make a recursive function as below.

```ml
let rec fib x =
    if x <= 1 then 1 else
    fib (x - 1) + fib (x - 2)
in
println_int (fib 10)
```

Functions can be nested.

```ml
let rec sqrt x =
    let rec abs x = if x > 0.0 then x else -.x in
    let rec go z p =
        if abs (p -. z) <= 0.00001 then z else
        let (p, z) = z, z -. (z *. z -. x) /. (2.0 *. z) in
        go z p
    in
    go x 0.0
in
println_float (sqrt 10.0)

(* Error because of out of scope: go 10.0 0.0 *)
```

In above example, `abs` and `go` is nested in `sqrt`. Nested function is a hidden implementation of the outer function because inner scope is not visible from outside.

Functions can capture any outer variables (=environment). Functions which captured outer
environment are called 'closure'.  As many functional languages or modern languages,
GoCaml has closure functions.

```ml
(* Define variable *)
let pi = 3.14 in

(* Captures outer defined variable 'pi' into its body *)
let rec circle r = r *. r *. pi in

(* Invoke closure *)
println_float (circle 10.0)
```

Below is a bit more complecated example:

```ml
let rec make_adder x =
    let z = 1 in
    let rec f y = x + y + z in
    f
in
let add = make_adder 3 in

(* Output: 104 *)
println_int (add 100)
```

Here, inner function `f` captures hidden variable `special_value`. `make_special_value_adder` returns a closure which captured the variable.

</details>

<details>
<summary>Tuples</summary>

N-elements tuple can be created with comma-separated expression `e1, e2, ..., en`. Element of tuple can be extracted with `let` expression.

```
(* (int, bool, string) is bound to t *)
let t = 1, true, "aaa" in

(* Destructuring tuple with `let` expression *)
let i, b, s = t in

let rec fst pair = let x, _ = pair in x in

(* Show '42' *)
println_int (fst (42, true))
```

</details>

<details>
<summary>Arrays</summary>

Array can be created with `Array.make size elem` where created array is allocated with `size` elemens
and all elements are initialized as `elem`.

`arr.(idx)` accesses to the element of array where `arr` is an array and `idx` is an integer.
And `arr.(idx) <- val` updates the `idx`th element to `val`.

```ml
(* Make boolean array whose size is 42 *)
let arr = Array.make 42 true in

(* Output: true *)
println_bool arr.(8)

(* Update element *)
arr.(8) <- false;

(* Output: false *)
println_bool arr.(8)
```

Note that arrays are NOT immutable because of performance (GoCaml doesn't have persistentarray).
`e1.(e2) <- e3` is always evaluated to `()` and updates the element destructively.
Accessing to out of bounds of arrays causes undefined behavior.

</details>

<details>
<summary>External symbols</summary>

All symbols which are not defined but used are treated as external symbols.
External symbol means `extern` names in C. So you have responsibility to define the symbols
in other object which will be linked to executable. Please see below 'How to Work with C'
section to know how to do that.

Note that all external symbols' types MUST be determined by type inference. Unknown type symbol causes compilation error.

```ml
(* This causes compile error because type of 'x' is unknown *)
x;

(* Type of 'y' is known as 'int' because it's passed to argument of `println_int` *)
println_int y;

(* When return type of function is (), it is treated as void in C *)
some_func ();

(* Below 'pow' function is inferred as float -> float -> float *)
print_float (pow 3.0 1.0 2.0)
```

</details>

## Prerequisities

- Go 1.2+ (Go 1.7+ is recommended)
- GNU make
- Clang
- cmake (for building LLVM)
- Git

## Installation

```sh
$ go get -d github.com/rhysd/gocaml
$ cd $GOPATH/src/github.com/rhysd/gocaml

# Full-installation with building LLVM locally
$ make

# Use system-installed LLVM. You need to install LLVM in advance (see below)
$ USE_SYSTEM_LLVM=true make
```

If you want to use `USE_SYSTEM_LLVM`, you need to install LLVM 4.0.0 in advance.

If you use Debian-family Linux, use [LLVM apt repository][] or download [LLVM official binary][].

```sh
$ sudo apt-get install libllvm4.0 llvm-4.0-dev
$ export LLVM_CONFIG=llvm-config-4.0
```

If you use macOS, use [Homebrew][]. GoCaml's installation script will automatically detect LLVM
installed with Homebrew.

```sh
$ brew install llvm
```

And you need to install [libgc][] as dependency.

```sh
# On Debian-family Linux
$ sudo apt-get install libgc-dev

# On macOS
$ brew install bdw-gc
```

## Usage

`gocaml` command is available to compile sources. Please refer `gocaml -help`.

```
Usage: gocaml [flags] [file]

  Compiler for GoCaml.
  When file is given as argument, compiler will compile it. Otherwise, compiler
  attempt to read from STDIN as source code to compile.

Flags:
  -asm
    	Emit assembler code to stdout
  -ast
    	Show AST for input
  -externals
    	Display external symbols
  -g	Compile with debug information
  -gcil
    	Emit GoCaml Intermediate Language representation to stdout
  -help
    	Show this help
  -ldflags string
    	Flags passed to underlying linker
  -llvm
    	Emit LLVM IR to stdout
  -obj
    	Compile to object file
  -opt int
    	Optimization level (0~3). 0: none, 1: less, 2: default, 3: aggressive (default -1)
  -show-targets
    	Show all available targets
  -target string
    	Target architecture triple
  -tokens
    	Show tokens for input
```

Compiled code will be linked to [small runtime][]. In runtime, some functions are defined to print values and it includes
`<stdlib.h>` and `<stdio.h>`. So you can use them from GoCaml codes.

## Program Arguments

You can access to program arguments via special global variable `argv`. `argv` is always defined before program starts.

```ml
print_str "argc: "; println_int (Array.length argv);
print_str "prog: "; println_str (argv.(0))
```

## Built-in Functions

Built-in functions are defined as external symbols.

- `print_int : int -> ()`
- `print_bool : bool -> ()`
- `print_float : float -> ()`
- `print_str : string -> ()`

Output the value to stdout.

- `println_int : int -> ()`
- `println_bool : bool -> ()`
- `println_float : float -> ()`
- `println_str : string -> ()`

Output the value to stdout with newline.

- `float_to_int : float -> int`
- `int_to_float : int -> float`
- `int_to_str : int -> string`
- `str_to_int : string -> int`
- `float_to_str : float -> string`
- `str_to_float : string -> float`

Convert between float and int, string and int, float and int.

- `str_length : string -> int`

Return the size of string.

- `str_concat : string -> string -> string`

Concat two strings as a new allocated string because strings are immutable in GoCaml.

- `substr : string -> int -> int -> string`

Returns substring of first argument. Second argument is an index to start and Third argument is an index to end.
Returns string slice `[start, end)` so it does not cause any allocation.

- `get_line : () -> string`
- `get_char : () -> string`

Get user input by line or character and return it as string.

- `to_char_code : string -> int`
- `from_char_code : int -> string`

Covert between a character and integer. First character of string is converted into integer and
integer is converted into one character string.

## How to Work with C

All symbols not defined in source are treated as external symbols. So you can define it in C source and link it to compiled GoCaml
code after.

Let's say to write C code.

```c
// gocaml.h is put in runtime/ directory. Please add it to include directory path.
#include "gocaml.h"

gocaml_int plus100(gocaml_int const i)
{
    return i + 100;
}
```

Then compile it to an object file:

```
$ clang -Wall -c plus100.c -o plus100.o
```

Then you can refer the function from GoCaml code:

```ml
println_int (plus100 10)
```

`println_int` is a function defined in runtime. So you don't need to care about it.

Finally comile the GoCaml code and the object file together with `gocaml` compiler. You need to link `.o` file after compiling
GoCaml code by passing the object file to `-ldflags`.
```
$ gocaml -ldflags plus100.o test.ml
```

After the command, you can find `test` executable. Executing by `./test` will show `110`.

## Cross Compilation

For example, let's say to want to make an `x86` binary on `x86_64` Ubuntu.

```
$ cd /path/to/gocaml
$ make clean
# Install gcc-multilib
$ sudo apt-get install gcc-4.8-multilib
```

Then you need to know [target triple][] string for the architecture compiler will compile into.
The format is `{name}-{vendor}-{sys}-{abi}`. (ABI might be omitted)

You can know all the supported targets by below command:

```
$ ./gocaml -show-targets
```

Then you can compile source into object file for the target.

```
# Create object file for specified target
$ ./gocaml -obj -target i686-linux-gnu source.ml

# Compile runtime for the target
CC=gcc CFLAGS=-m32 make ./runtime/gocamlrt.a
```

Finally link object files into one executable binary by hand.

```
$ gcc -m32 -lgc source.o ./runtime/gocamlrt.a
```

[MinCaml]: https://github.com/esumii/min-caml
[goyacc]: https://github.com/cznic/goyacc
[LLVM]: http://llvm.org/
[Linux and macOS Build Status]: https://travis-ci.org/rhysd/gocaml.svg?branch=master
[Travis CI]: https://travis-ci.org/rhysd/gocaml
[lexer doc]: https://godoc.org/github.com/rhysd/gocaml/lexer
[parser doc]: https://godoc.org/github.com/rhysd/gocaml/parser
[typing doc]: https://godoc.org/github.com/rhysd/gocaml/typing
[alpha transform doc]: https://godoc.org/github.com/rhysd/gocaml/alpha
[gcil doc]: https://godoc.org/github.com/rhysd/gocaml/gcil
[closure doc]: https://godoc.org/github.com/rhysd/gocaml/closure
[codegen doc]: https://godoc.org/github.com/rhysd/gocaml/codegen
[MinCaml spec]: http://esumii.github.io/min-caml/paper.pdf
[Boehm GC]: https://github.com/ivmai/bdwgc
[Coverage Status]: https://coveralls.io/repos/github/rhysd/gocaml/badge.svg
[Coveralls]: https://coveralls.io/github/rhysd/gocaml
[Windows Build Status]: https://ci.appveyor.com/api/projects/status/7lfewhhjg57nek2v/branch/master?svg=true
[Appveyor]: https://ci.appveyor.com/project/rhysd/gocaml/branch/master
[small runtime]: ./runtime/gocamlrt.c
[LLVM apt repository]: http://apt.llvm.org/
[Homebrew]: https://brew.sh/index.html
[libgc]: https://www.hboehm.info/gc/
[target triple]: https://clang.llvm.org/docs/CrossCompilation.html#target-triple
[examples]: ./examples
[Brainfxxk interpreter example]: ./examples/brainfxxk.ml
[LLVM official binary]: http://releases.llvm.org/download.html#4.0.0
