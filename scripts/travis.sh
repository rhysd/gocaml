#! /bin/bash

set -e

export USE_SYSTEM_LLVM=true

if [[ "$TRAVIS_OS_NAME" == "osx" ]]; then
    brew update
    brew install bdw-gc
    brew install llvm --with-libffi
    export LLVM_CONFIG="/usr/local/Cellar/llvm/4.0.0/bin/llvm-config"
    make build
    go test -v ./...
else
    go get golang.org/x/tools/cmd/cover
    go get github.com/haya14busa/goverage
    go get github.com/mattn/goveralls
    export LLVM_CONFIG="llvm-config-4.0"
    make build
    go test -v ./...
    make cover.out
    go tool cover -func cover.out
    goveralls -coverprofile cover.out -service=travis-ci
fi

