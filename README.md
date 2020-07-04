GoCaml :camel:
==============
[![Linux and macOS Build Status][]][Travis CI]
[![Windows Build Status][]][Appveyor]
[![Coverage Status][]][Codecov]

GoCaml is subset of OCaml in Go based on [MinCaml][] using [LLVM][]. GoCaml adds many features to original MinCaml.
MinCaml is a minimal subset of OCaml for educational purpose. It is statically-typed and compiled into a binary.

This project aims incremental compiler development for my own programming language.
Type inference, closure transform, mid-level IR are implemented.

- [Japanese presentation (Go Conference Tokyo 2017 Spring)](https://speakerdeck.com/rhysd/go-detukurufan-yong-yan-yu-chu-li-xi-shi-zhuang-zhan-lue)
- [Japanese presentation (builderscon 2017)](https://speakerdeck.com/rhysd/xiao-sakushi-meteyu-terukonpaira)

Example:

```ml
let rec gcd m n =
  if m = 0 then n else
  if m <= n then gcd m (n - m) else
  gcd n (m - n) in
print_int (gcd 21600 337500)
```

You can see [more examples][examples]. (e.g. [Brainfxxk interpreter][Brainfxxk interpreter example], [N-Queens puzzle][N-Queens puzzle example])

## Tasks

- [x] Lexer -> ([doc][syntax doc])
- [x] Parser with [goyacc][] -> ([doc][syntax doc])
- [x] Alpha transform ([doc][sema doc])
- [x] Type inference (Hindley Milner monomorphic type system) -> ([doc][sema doc])
- [x] mid-level intermediate representation (MIR) ([doc][mir doc])
- [x] K normalization from AST into MIR ([doc][mir doc])
- [x] Closure transform ([doc][closure doc])
- [x] Code generation (LLVM IR, assembly, object, executable) using [LLVM][] ([doc][codegen doc])
- [x] LLVM IR level optimization passes
- [x] Garbage collection with [Boehm GC][]
- [x] Debug information (DWARF) using LLVM's Debug Info builder

## Difference from Original MinCaml

- MinCaml assumes external symbols' types are `int` when it can't be inferred. GoCaml requires explicit declaration for external
  symbols with `external` syntax like OCaml. All external symbols (except for builtins) must be declared. Otherwise, undeclared
  symbols are reported undefined symbols as compilation error.
- MinCaml allows `-` unary operator for float literal. So for example `-3.14` is valid but `-f` (where `f` is `float`) is not valid.
  GoCaml does not allow `-` unary operator for float values totally. You need to use `-.` unary operator instead (e.g. `-.3.14`).
- GoCaml adds more operators. `*` and `/` for integers, `&&` and `||` for booleans.
- GoCaml has string type. String value is immutable and used with slices.
- GoCaml does not have `Array.create`, which is an alias to `Array.make`. `Array.length` is available to obtain the size of array
  and `[| ... |]` literal is available to create an array with specific elements.
- Some useful built-in functions are added (described in below section).
- [Option type][] is implemented in GoCaml. Please see below 'Option Type' section or [test cases][option type test cases].
- GoCaml has `fun` syntax to make an anonymous funcion or closure like `fun x y -> x + y`.
- GoCaml has type annotations syntax. Users can specify types explicitly.
- Symbols named `_` are ignored.
- Type alias using `type` keyword.

## Language Spec

### Program

Program is represented as one expression which MUST be evaluated as unit type. So `()` is the
smallest program for GoCaml.

### Sequence Expression

Sequenced program can be represented by joining multiple expressons with `;`.

```ml
e1; e2; e3; e4
```

In above program, expressions are evaluated in order of `e1 -> e2 -> e3 -> e4` and the sequenced
expression is evaluated to the value of `e4`.
Program must be evaluated to unit type, so the `e4` expression must be evaluated to `()` (unit value).

### Comments

There is a block comment syntax. It starts with `(*` and ends with `*)`. Any comment must be closed
with `*)`, otherwise it falls into syntax error.

```ml
(*
   This is comment.
*)
```

### Constants

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

### Show values

`print_*` and `println_*` built-in functions are available to output values to stdout.

```ml
print_int 42;
println_bool true
```

Please see 'Built-in functions' section below for more detail.

### Unary operators

You can use some unary prefixed operators.

```ml
-42;

(* GoCaml distinguishes float and int in operators. -. is a float version of - *)
-.3.14;

not true;

()
```

### Arithmetic binary operators

As mentioned above, GoCaml distinguishes int and float in operators. Operators for float values are
suffixed by `.` (dot).

```ml
(* integer calculation *)
1 + 2;
1 - 2;
1 * 2;
1 / 2;

(* float calculation *)
1.0 +. 2.0;
1.0 -. 2.0;
1.0 *. 2.0;
1.0 /. 2.0;

()
```

Integer operators must have integer values as their operands. And float operators must have float
values as their operands. There is no implicit conversion. You need to convert explicitly by using
built-in functions (e.g. `3.14 +. (int_to_float 42)`).

Note that strings don't have any operators for concatenating two strings or slicing sub string.
They can be done with `str_concat` and `str_sub` built-in functions (See 'Built-in Functions' section).

### Relational operators

Equal operator is `=` (NOT `==`), Not-equal operator is `<>`. Compare operators are the same as C
(`<`, `<=`, `>` and `>=`).

```ml
42 = 42; (* => true *)
42 <> 42; (* => false *)

3.14 > 2.0;
1.0 < 3.0;
2.0 >= 2.0;
1.0 <= 3.0;

()
```

Tuples (described below) and strings can be compared with `=` or `<>`, but cannot be compared with
`<`, `<=`, `>` and `>=`. Arrays (described below) cannot be compared directly with any compare
operators. You need to compare each element explicitly.

### Logical operators

`&&` and `||` are available for boolean values.

```ml
println_bool (true || false && false || false)
```

### Variable

`let` expression binds some value to a variable.

```ml
(* 42 is bound to a *)
let a = 42 in
(* You can use the variable as value in any expression *)
a + 4;

()
```

The syntax is `let {name} = {e1} in {e2}`. `e2` will be evaluated where `e1` is bound to `name`.
By chain `let`, you can define multiple variables.

```ml
let pi = 3.141592 in
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

(* Shows "hello" *)
f ();

(* Binds external function *)
let p = println_str in

(* Shows "hi" *)
p "hi"
```

### Functions

`let rec` is a keyword to define a function. Syntax is `let rec name params... = e1 in e2` where
function `name` is defined as `e1` and then `e2` will be evaluated.
`f a b c` is an expression to apply function `f` with argument `a`, `b` and `c`.
As long as the argument is simple, you don't need to use `()`.

Note that, if you use some complicated expression (for example, binary operators), you need to use
`()` like `f (a+b) c`. If you specify `f a + b c`, it would be parsed as `(f a) + (b c)`.

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

In above example, `abs` and `go` is nested in `sqrt`. Nested function is a hidden implementation of
the outer function because inner scope is not visible from outside.

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

Below is a bit more complicated example:

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

Here, inner function `f` captures hidden variable `special_value`. `make_special_value_adder`
returns a closure which captured the variable.

### Lambda

Functions can be made without names using `fun` syntax.

```ml
(* Make a lambda and bind it to foo *)
let add = fun x y -> x + y in

(* Above is like below, but the function is anonymous *)
let rec add2 x y in x + y in

println_int (add 1 10);
println_int (add2 1 10);
```

It's useful when passing a function without considering its name.

```ml
let rec quick_sort xs pred =
    (* ...snip *)
in
let a = Array.make 10 0 in
let sorted = quick_sort a (fun l r -> l < r) in
()
```

Lambda does not have its name, so it cannot be called recursively.

Using lambda, above `make_adder` can be implemented as following:

```ml
let rec make_adder x =
    let z = 1 in
    fun x y -> x + y + z
in
...
```

### Type Annotation

Type can be specified explicitly at any expression, parameter and return type of function with `:`

Types can be written in the same syntax as other ML languages.

- Primitive: `int`, `float`, `bool`, `string`
- Any type: `_`
- Tuple: `t1 * t2 * ... * tn` (e.g. `int * bool`)
- Function: `a -> b -> ... -> r` (e.g. if `f` takes `int` and `bool` and returns `string`, then `f: int -> bool -> string`)
- Array: `t array` (e.g. `int array`, `int array array`)
- Option: `t option` (e.g. `int option` `(int -> bool) option`)

Types can be specified in code as following. Compiler will look and check them in type inference.

```ml
(* Type of variable *)
let v: int = 42 in

(* Type of parameters *)
let rec f (x:int) = x + 10 in
let f = fun (x:int) -> x + 10 in

(* Type of return value *)
let rec f x: string = int_to_str x in
let f = fun x: string -> int_to_str x in

(* Type of parameter and return value *)
let rec f (x:int): string = int_to_str x in
let f = fun (x:int): string -> int_to_str x in

(* Type of tuple at `let` *)
let (a, b): int * bool = 42, bool in

(* Array type *)
let a: bool array = Array.make 3 true in
let a: int array array = Array.make 3 (Array.make 3 42) in

(* Option type *)
let o: int option = None in
let o: (int array * (int -> bool)) option = None in

(* '_' means 'any'. Specify type partially *)
let (a, b): _ * _ = 42, bool in
let f: _ -> _ = fun x -> x in
let a: _ array = Array.make 3 true in

()
```

### Type Alias

`type {name} = {type};` syntax declares type alias. It can be declared on toplevel. It means that
all type aliases must be put before any expression.

```ml
type board = int array array;
type point = int * int;
let p: point = 1, 2 in
let b: board = Array.make 4 (Array.make 4 0) in
()
```

In above example, `board` is an alias of `int array array`. It can be used the same as `int array array`.
Note that `type` does not make another type here. Just make an alias.

### Tuples

N-elements tuple can be created with comma-separated expression `e1, e2, ..., en`. Element of tuple
can be extracted with `let` expression.

```ml
(* (int, bool, string) is bound to t *)
let t = 1, true, "aaa" in

(* Destructuring tuple with `let` expression *)
let i, b, s = t in

let rec fst pair = let x, _ = pair in x in

(* Show '42' *)
println_int (fst (42, true))
```

### Arrays

Array can be created with `Array.make size elem` where created array is allocated with `size` elements
and all elements are initialized as `elem`. And array literal `[| e1; e2; ... |]` is also supported which
allocates an array with specified (`e1`, `e2`...) elements.

`arr.(idx)` accesses to the element of array where `arr` is an array and `idx` is an integer.
And `arr.(idx) <- val` updates the `idx`th element to `val`.

```ml
(* Make boolean array whose size is 42 *)
let arr = Array.make 42 true in

(* Output: true *)
println_bool arr.(8);

(* Update element *)
arr.(8) <- false;

(* Output: false *)
println_bool arr.(8);

(* Make an array with 1, 2 and 3 elements *)
let lit = [| 1; 2; 3 |] in

(* Output: 2 *)
println_int lit.(1);

(* Output: 3 *)
println_int (Array.length lit)
```

Note that arrays are NOT immutable because of performance (GoCaml doesn't have persistentarray).
`e1.(e2) <- e3` is always evaluated to `()` and updates the element destructively.
Accessing to out of bounds of arrays causes undefined behavior.

And note that list literal (`[e1; e2; ...]`) is not supported yet. Please do not be confused.

### Option Type

Option type represents some value or none.

```ml
let rec print o =
    match o with
      | Some i -> println_int o
      | None   -> println_str "none"
in
print None;
print (Some 42)
```

First `|` can be omitted so you can write it in one line.

```ml
if match o with Some i -> true | None -> false then
  println_str "some!"
else
  println_str "none..."
```

Option values can be compared with `=` or `<>` directly.

```ml
let rec is_some x = x <> None in
let rec is_none x = (x = None) in

println_bool (is_some (Some 42));
println_bool (is_some None);
println_bool (is_none (Some 42));
println_bool (is_none None)
```

Currently `match with` expression is only for option type because GoCaml doesn't have variant types.

### Ignored Symbol `_`

Variables named `_` are ignored. It's useful if the variable is never used.

```ml
(* Ignored variable *)
let _ = 42 in

(* Ignored parameter *)
let rec first x _ _ = x in

(* Ignored element of tuple *)
let (_, second, _) = 1, "foo", true in

(* Error! Cannot refer ignored variable *)
print_int _
```

If a type of an ignored variable is never determined, compiler regards its type as `()` and
compilation will pass.

```
let f _ = 42 in
println_int 42
```

In above program, the type of ignored variable `_` will never be determined because function `f` is
never used. In this case, compiler regards type of `f` as `unit -> int` and compilation will continue.

### External Symbols

All external symbol must be declared with `external` syntax.

```ml
external name: type = "c_name";
```

The `name` is a symbol name of the external symbol. And the `"c_name"` is a symbol name linked in
C level. The `type` cannot contain any generic type variable and `_`.
For example, when you define `gocaml_int foo(gocaml_int i)` function in C, then you need to declare
`"foo"` external C name with type `int -> int` to use it from GoCaml.

```ml
external foo: int -> int = "foo";
foo 42
```

Or when you define a global variable `gocaml_int x` in C, you need to declare external `"x"` C name
to use the global variable value from GoCaml.

```ml
external x: int = "x";
println_int x
```

If C name does not exist in link phase, compiler will cause a linker error at compiling the source.

Like `type` syntax, all `external` declarations should be written before any expression.

## Prerequisites

- Go 1.7+
- GNU make
- Clang or GCC (for building small runtime)
- cmake (for building LLVM)
- Git

## Installation

### Linux or macOS

For Linux or macOS, below commands build `gocaml` binary at root of the repository.
[libgc][] is needed as dependency.

```console
# On Debian-family Linux
$ sudo apt-get install libgc-dev

# On macOS
$ brew install go cmake bdw-gc coreutils

$ mkdir -p $GOPATH/src/github.com/rhysd && cd $GOPATH/src/github.com/rhysd
$ git clone https://github.com/rhysd/gocaml.git
$ cd gocaml

# Full-installation with building LLVM locally
$ make
```

The `make` command will do all. First, it clones LLVM into `$GOPATH/src/llvm.org/llvm/` and builds
it for LLVM Go binding. Second, it builds `gocaml` binary and `gocamlrt.a` runtime. Finally, it
runs all tests for validation.
Note that `go get -d` is not available because `llvm.org/*` dependency is not go-gettable for now.

Above is the easiest way to install gocaml, but if you want to use system-installed LLVM instead of
building `$GOPATH/src/llvm.org/llvm`, please follow build instruction.

`USE_SYSTEM_LLVM=true` will build `gocaml` binary with system-installed LLVM libraries.
Note that it still clones LLVM repository because `$GOPATH/src/llvm.org/llvm/bindings/go/*` is
necessary for building gocaml.

To use `USE_SYSTEM_LLVM`, you need to install LLVM 4.0.0 or later (5.0.0 is recommended) with system's
package manager in advance.

If you use Debian-family Linux, use [LLVM apt repository][] or download [LLVM official binary][].

```console
$ sudo apt-get install libllvm5.0 llvm-5.0-dev
$ export LLVM_CONFIG=llvm-config-5.0
```

If you use macOS, use [Homebrew][]. GoCaml's installation script will automatically detect LLVM
installed with Homebrew.

```console
$ brew install llvm
```

Now you can build gocaml with `USE_SYSTEM_LLVM` flag.

```console
$ USE_SYSTEM_LLVM=true make
```

### Windows

Currently Windows is not well-supported. You need to clone LLVM repository to `$GOPATH/src/llvm.org/`
and build Go bindings of llvm-c following [the instruction][Go binding building instruction].
It needs `cmake` command and C++ toolchain.

It also needs to build [libgc][] static library and put it to library path.

After installing [goyacc][], generate a parser code with it.

```
$ goyacc -o parser/grammar.go parser/grammar.go.y
```

Finally you can build `gocaml` binary with `go build`.

## Usage

`gocaml` command is available to compile sources. Please refer `gocaml -help`.

```
Usage: gocaml [flags] [file]

  Compiler for GoCaml.
  When file is given as argument, compiler will compile it. Otherwise, compiler
  attempt to read from STDIN as source code to compile.

Flags:
  -analyze
    	Analyze code and report errors if exist
  -asm
    	Emit assembler code to stdout
  -ast
    	Show AST for input
  -dump-env
    	Dump analyzed symbols and types information to stdout
  -g	Compile with debug information
  -help
    	Show this help
  -ldflags string
    	Flags passed to underlying linker
  -llvm
    	Emit LLVM IR to stdout
  -mir
    	Emit GoCaml Intermediate Language representation to stdout
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

Compiled code will be linked to [small runtime][]. In runtime, some functions are defined to print
values and it includes `<stdlib.h>` and `<stdio.h>`. So you can use them from GoCaml codes.

`gocaml` uses `clang` for linking objects by default. If you want to use other linker, set
`$GOCAML_LINKER_CMD` environment variable to your favorite linker command.

## Program Arguments

You can access to program arguments via special global variable `argv`. `argv` is always defined
before program starts.

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

- `str_sub : string -> int -> int -> string`

Returns substring of first argument. Second argument is an index to start and Third argument is an
index to end.
Returns string slice `[start, end)` so it does not cause any allocation.

- `get_line : () -> string`
- `get_char : () -> string`

Get user input by line or character and return it as string.

- `to_char_code : string -> int`
- `from_char_code : int -> string`

Covert between a character and integer. First character of string is converted into integer and
integer is converted into one character string.


- `do_garbage_collection : () -> ()`
- `enable_garbage_collection : () -> ()`
- `disable_garbage_collection : () -> ()`

These functions control behavior of GC. `do_garbage_collection` runs GC with stopping the world.
`enable_garbage_collection`/`disable_garbage_collection` starts/stops GC. (GC is enabled by default)

- `bit_and : int -> int -> int`
- `bit_or : int -> int -> int`
- `bit_xor : int -> int -> int`
- `bit_rsft : int -> int -> int`
- `bit_lsft : int -> int -> int`
- `bit_inv : int -> int`

Built-in functions instead of bitwise operators.

- `time_now : () -> int`

Returns epoch time in seconds.

- `read_file : string -> string option`

First argument is a file name. It returns the content of the file. If failed, it returns `None`.

- `write_file : string -> string -> bool`

It takes file name as first argument and its content as second argument.
It returns wether it could write the content to the file.


- `ceil : float -> float`
- `floor : float -> float`
- `exp : float -> float`
- `log : float -> float`
- `log10 : float -> float`
- `log1p : float -> float`
- `sqrt : float -> float`
- `sin : float -> float`
- `cos : float -> float`
- `tan : float -> float`
- `asin : float -> float`
- `acos : float -> float`
- `atan : float -> float`
- `atan2 : float -> float`
- `sinh : float -> float`
- `cosh : float -> float`
- `tanh : float -> float`
- `asinh : float -> float`
- `acosh : float -> float`
- `atanh : float -> float`
- `hypot : float -> float -> float`
- `mod_float : float -> float -> float`
- `modf : float -> float * float`
- `frexp : float -> float * int`
- `ldexp : float -> int -> float`

Basic math functions. This is the same functions as defined in [OCaml's `Pervasives` module][OCaml Pervasives module].

## Built-in Constants

- `infinity : float`
- `nan : float`

Floating point values represent initinity and NaN. It's the same values as defined in
[OCaml's `Pervasives` module][OCaml Pervasives module].

## How to Work with C

All symbols not defined in source are treated as external symbols. So you can define it in C source
and link it to compiled GoCaml code after.

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

Finally comile the GoCaml code and the object file together with `gocaml` compiler. You need to link
`.o` file after compiling GoCaml code by passing the object file to `-ldflags`.

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
[Travis CI]: https://travis-ci.org/rhysd/gocaml/branches
[syntax doc]: https://godoc.org/github.com/rhysd/gocaml/syntax
[sema doc]: https://godoc.org/github.com/rhysd/gocaml/sema
[mir doc]: https://godoc.org/github.com/rhysd/gocaml/mir
[closure doc]: https://godoc.org/github.com/rhysd/gocaml/closure
[codegen doc]: https://godoc.org/github.com/rhysd/gocaml/codegen
[Boehm GC]: https://github.com/ivmai/bdwgc
[Coverage Status]: https://codecov.io/gh/rhysd/gocaml/branch/master/graph/badge.svg
[Codecov]: https://codecov.io/gh/rhysd/gocaml
[Windows Build Status]: https://ci.appveyor.com/api/projects/status/7lfewhhjg57nek2v/branch/master?svg=true
[Appveyor]: https://ci.appveyor.com/project/rhysd/gocaml/branch/master
[small runtime]: ./runtime/gocamlrt.c
[LLVM apt repository]: http://apt.llvm.org/
[Homebrew]: https://brew.sh/index.html
[libgc]: https://www.hboehm.info/gc/
[target triple]: https://clang.llvm.org/docs/CrossCompilation.html#target-triple
[examples]: ./examples
[Brainfxxk interpreter example]: ./examples/brainfxxk.ml
[N-Queens puzzle example]: ./examples/n-queens.ml
[LLVM official binary]: http://releases.llvm.org/download.html#5.0.0
[Go binding building instruction]: https://github.com/llvm-mirror/llvm/blob/master/bindings/go/README.txt
[goyacc]: https://godoc.org/golang.org/x/tools/cmd/goyacc
[Option type]: https://en.wikipedia.org/wiki/Option_type
[option type test cases]: ./codegen/testdata/option_values.ml
[OCaml Pervasives module]: https://caml.inria.fr/pub/docs/manual-ocaml/libref/Pervasives.html
