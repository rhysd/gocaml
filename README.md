GoCaml
======
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
- [ ] Optimizations
  - [ ] Beta reduction
  - [ ] Inlining
  - [ ] Folding constants
  - [ ] Striping unused variables
- [ ] Code generation using [LLVM][]
- [ ] Garbage collection with [Boehm GC][]

## Difference from original MinCaml

- MinCaml assumes external symbols' types are `int` when it can't be inferred. GoCaml does not have such an assumption.
  GoCaml assumes unknown return type of external functions as `()` (`void` in C), but in other cases, falls into compilation error.
  When you use nested external functions call, you need to clarify the return type of inner function call. For example, when `f` in
  `g (f ())` returns `int`, you need to show it like `g ((f ()) + 0)`.
- MinCaml allows `-` unary operator for float literal. So for example `-3.14` is valid but `-f` (where `f` is `float`) is not valid.
  GoCaml does not allow `-` unary operator for float values totally.

## Installation

```
go get -u github.com/rhysd/gocaml
```

Or clone this repository into your `$GOPATH/src/github.com/rhysd/gocaml` and execute `make` in the directory.

## Usage

`gocaml` command is available to compile sources. Please refer `gocaml -help`.

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
[MinCaml spec]: http://esumii.github.io/min-caml/paper.pdf
[Boehm GC]: https://github.com/ivmai/bdwgc
[Coverage Status]: https://coveralls.io/repos/github/rhysd/gocaml/badge.svg
[Coveralls]: https://coveralls.io/github/rhysd/gocaml
[Windows Build Status]: https://ci.appveyor.com/api/projects/status/7lfewhhjg57nek2v/branch/master?svg=true
[Appveyor]: https://ci.appveyor.com/project/rhysd/gocaml/branch/master
