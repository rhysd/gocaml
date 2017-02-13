#!/bin/bash

set -e

if [[ "$GOPATH" == "" ]]; then
    echo '$GOPATH is empty' 1>&2
    exit 4
fi

LLVM_ORG_DIR="${GOPATH}/src/llvm.org"
LLVM_DIR="${LLVM_ORG_DIR}/llvm"
LLVM_GO_DIR="${LLVM_DIR}/bindings/go"
LLVM_ARCHIVE="$GOPATH/pkg/llvm.org/llvm/bindings/go/llvm/llvm.a"

if [[ -f "$LLVM_ARCHIVE" ]]; then
    echo "LLVM is already installed: ${LLVM_ARCHIVE}. Installation skipped."
    exit
fi

rm -rf "$LLVM_DIR"
mkdir -p "$LLVM_ORG_DIR"
cd "$LLVM_ORG_DIR"

git clone --depth 1 -b release_40 --single-branch http://llvm.org/git/llvm.git
cd "$LLVM_GO_DIR"

# -DCMAKE_BUILD_TYPE=Debug makes `go build` too slow because clang's linker is very slow with dwarf.
./build.sh -DCMAKE_BUILD_TYPE=Release

go install -a ./llvm
