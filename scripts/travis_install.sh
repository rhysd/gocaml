#! /bin/bash

set -e

export USE_SYSTEM_LLVM=true

if [[ "$TRAVIS_OS_NAME" == "osx" ]]; then
    brew update
    brew info llvm
    brew install bdw-gc llvm
else
    go get golang.org/x/tools/cmd/cover
    go get github.com/haya14busa/goverage
    go get github.com/mattn/goveralls
    export LLVM_CONFIG="llvm-config-6.0"
    export CC=gcc-7
    export CXX=g++-7
fi

make build
