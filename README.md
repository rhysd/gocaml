GoCaml
======
[![Build Status][]][Travis CI]

GoCaml is a [MinCaml][] implementation in Go using [LLVM][]. MinCaml is a minimal subset of OCaml for educational purpose ([spec][MinCaml spec]).

This project aims my practices for understanding type inference and introducing own intermediate language (IL) to own language.

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
- [x] Type inference (Hindley Milner monomorphic type system) -> ([doc][typing doc])
- [ ] GoCaml intermediate language (GCIL)
- [ ] K nomarization from AST into GCIL
- [x] Alpha transform ([doc][alpha transform doc])
- [ ] Beta reduction
- [ ] Closure transform
- [ ] Optimizations
  - [ ] Inlining
  - [ ] Folding constants
  - [ ] Striping unused variables
- [ ] Code generation using [LLVM][]

## Difference from original MinCaml

- MinCaml assumes external symbols' types are `int` when it can't be inferred. GoCaml does not have such an assumption.
  GoCaml assumes unknown return type of external functions as `()` (`void` in C), but in other cases, falls into compilation error.
- MinCaml allows `-` unary operator for float literal. So for example `-3.14` is valid but `-f` (where `f` is `float`) is not valid.
  GoCaml does not allow `-` unary operator for float values totally.

## Installation

```
go get -u github.com/rhysd/gocaml.git
```

Or clone this repository and execute `make` in the directory.

## Usage

`gocaml` command is available to compile sources. Please refer `gocaml -help`.

[MinCaml]: https://github.com/esumii/min-caml
[goyacc]: https://github.com/cznic/goyacc
[LLVM]: http://llvm.org/
[Build Status]: https://travis-ci.org/rhysd/gocaml.svg?branch=master
[Travis CI]: https://travis-ci.org/rhysd/gocaml
[lexer doc]: https://godoc.org/github.com/rhysd/gocaml/lexer
[parser doc]: https://godoc.org/github.com/rhysd/gocaml/parser
[typing doc]: https://godoc.org/github.com/rhysd/gocaml/typing
[alpha transform doc]: https://godoc.org/github.com/rhysd/gocaml/alpha
[MinCaml spec]: http://esumii.github.io/min-caml/paper.pdf
