GoCaml :camel:
==============
[![Linux and macOS Build Status][]][Travis CI]
[![Windows Build Status][]][Appveyor]
[![Coverage Status][]][Coveralls]

GoCaml is a [MinCaml][] implementation in Go using [LLVM][]. MinCaml is a minimal subset of OCaml for educational purpose ([spec][MinCaml spec]).

This project aims my practices for understanding type inference, closure transform and introducing own intermediate language (IL) to own language.

Example:

```ocaml
let rec gcd m n =
  if m = 0 then n else
  if m <= n then gcd m (n - m) else
  gcd n (m - n) in
print_int (gcd 21600 337500)
```

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
- GoCaml does not have `Array.create`, which is an alias to `Array.make`.

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

If you use Debian-family Linux, use [LLVM apt repository][]

```sh
$ sudo apt-get install libllvm4.0 llvm-4.0-dev
```

If you use macOS, use [Homebrew][]. GoCaml's installation script will automatically detect LLVM
installed with Homebrew.

*Note:* LLVM 4.0 is now on an RC stage. So it doesn't come to Homebrew yet.

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

## Builtin Functions

Built-in functions are defined as external symbols.

- `print_int :: (int) -> ()`
- `print_bool :: (bool) -> ()`
- `print_float :: (float) -> ()`
- `print_str :: (string) -> ()`

Output the value to stdout.

- `println_int :: (int) -> ()`
- `println_bool :: (bool) -> ()`
- `println_float :: (float) -> ()`
- `println_str :: (string) -> ()`

Output the value to stdout with newline.

- `float_to_int :: (float) -> int`
- `int_to_float :: (int) -> float`
- `int_to_str :: (int) -> string`
- `str_to_int :: (string) -> int`
- `float_to_str :: (float) -> string`
- `str_to_float :: (string) -> float`

Convert between float and int, string and int, float and int.

- `str_size :: (string) -> int`

Return the size of string.

- `str_concat :: (string, string) -> string`

Concat two strings as a new allocated string because strings are immutable in GoCaml.

- `substr :: (string, int, int) -> string`

Returns substring of first argument. Second argument is an index to start and Third argument is an index to end. Returns string slice `[start, end)` so it does not cause any allocation.

- `get_line :: (()) -> string`

Get user input by line and return it as string.

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
println_int ((plus100 10) + 0)
```

`println_int` is a function defined in runtime. So you don't need to care about it.
The `+ 0` is necessary to tell a compiler that the type of returned value of `plus100` is `int`. A compiler can know the type
via type inference.

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
