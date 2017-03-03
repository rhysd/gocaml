#! /bin/bash

set -e

export USE_SYSTEM_LLVM=true

if [[ "$TRAVIS_OS_NAME" == "osx" ]]; then
    brew install bdw-gc
    # brew update
    # brew install llvm
    #
    # Fallback until LLVM 4.0 comes to Homebrew
    mkdir -p llvm-4.0.0-rc2-workaround
    if [ ! -d llvm-4.0.0-rc2-workaround/bin ]; then
        wget -O clang+llvm-4.0.0-rc2.tar.xz http://www.llvm.org/pre-releases/4.0.0/rc2/clang+llvm-4.0.0-rc2-x86_64-apple-darwin.tar.xz
        tar -xvf clang+llvm-4.0.0-rc2.tar.xz --strip 1 -C llvm-4.0.0-rc2-workaround
    fi
    export LLVM_CONFIG="$(pwd)/llvm-4.0.0-rc2-workaround/bin/llvm-config"

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

