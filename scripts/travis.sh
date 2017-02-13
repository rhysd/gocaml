#! /bin/bash

set -e

PACKAGES="./alpha ./ast ./gcil ./closure ./lexer ./parser ./token ./typing"

if [[ "$TRAVIS_OS_NAME" == "osx" ]]; then
    # Avoid building LLVM
    go get -v -t -d $PACKAGES
    go get golang.org/x/tools/cmd/goyacc
    go build -v $PACKAGES
    go test -v $PACKAGES
else
    go get golang.org/x/tools/cmd/cover
    go get github.com/haya14busa/goverage
    go get github.com/mattn/goveralls
    make build
    go test -v ./...
    make cover.out
    go tool cover -func cover.out
    goveralls -coverprofile cover.out -service=travis-ci -repotoken $COVERALLS_TOKEN
fi

