GoCaml
======

GoCaml is a [MinCaml][] implementation in Go using [LLVM][]. MinCaml is an educational compiler for a minimal subset of OCaml.

This project aims my practices for understanding type inference and introducing own intermediate language (IL) to own language.

```ml
let rec gcd m n =
  if m = 0 then n else
  if m <= n then gcd m (n - m) else
  gcd n (m - n) in
print_int (gcd 21600 337500)
```

- [x] Lexer -> ([./lexer](./lexer))
- [x] Parser with [goyacc][] -> ([./parser](./parser))
- [x] Type inference (Hindley Milner monomorphic type system) -> ([./typing](./typing))
- [ ] GoCaml intermediate language (GCIL)
- [ ] K nomarization from AST into GCIL
- [ ] Alpha transform
- [ ] Beta reduction
- [ ] Closure transform
- [ ] Optimizations
  - [ ] Inlining
  - [ ] Folding constants
  - [ ] Striping unused variables
- [ ] Code generation using [LLVM][]

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
