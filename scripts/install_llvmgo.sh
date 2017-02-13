#!/bin/bash

set -e

if [[ "$GOPATH" == "" ]]; then
    echo '$GOPATH is empty' 1>&2
    exit 4
fi

LLVM_ORG_DIR="${GOPATH}/src/llvm.org"
LLVM_DIR="${LLVM_ORG_DIR}/llvm"
LLVM_GO_DIR="${LLVM_DIR}/bindings/go"
LLVM_ARCHIVE="$GOPATH/pkg/llvm.org/llvm/bindings/go/llvm.a"

if [[ -f "$LLVM_ARCHIVE" ]]; then
    echo "LLVM is already installed: ${LLVM_ARCHIVE}. Installation skipped."
    exit
fi

if [[ "$LLVM_BRANCH" == "" ]]; then
    LLVM_BRANCH="release_40"
fi

rm -rf "$LLVM_DIR"
mkdir -p "$LLVM_ORG_DIR"
cd "$LLVM_ORG_DIR"

echo "Cloning LLVM branch: ${LLVM_BRANCH}..."
git clone --depth 1 -b $LLVM_BRANCH --single-branch http://llvm.org/git/llvm.git
cd "$LLVM_GO_DIR"

if [[ "$USE_SYSTEM_LLVM" == "" ]]; then
    echo "Building LLVM locally: ${LLVM_DIR}"
    # -DCMAKE_BUILD_TYPE=Debug makes `go build` too slow because clang's linker is very slow with dwarf.
    ./build.sh -DCMAKE_BUILD_TYPE=Release

    go install -a ./llvm
    exit
fi

echo "Building go-llvm with system installed LLVM..."

if [[ "$LLVM_CONFIG" == "" ]]; then
    LLVM_CONFIG="llvm-config"
fi

if which llvm-config 2>&1 > /dev/null; then
    echo "llvm-config version: $($LLVM_CONFIG --version)"
else
    echo "llvm-config command not found: $LLVM_CONFIG" >&2
    exit 1
fi

cd ./llvm

export CGO_CPPFLAGS="$($LLVM_CONFIG --cppflags)"
export CGO_CXXFLAGS="$($LLVM_CONFIG --cxxflags)"
export CGO_LDFLAGS="$($LLVM_CONFIG --ldflags --libs --system-libs all | tr '\n' ' ')"
echo "CGO_CPPFLAGS=$CGO_CPPFLAGS"
echo "CGO_CXXFLAGS=$CGO_CXXFLAGS"
echo "CGO_LDFLAGS=$CGO_LDFLAGS"

go build -v -tags byollvm
go install
