#! /bin/bash

set -e

if [[ "$TRAVIS_OS_NAME" == "osx" ]]; then
    brew update
    brew upgrade go
    go get -t -d -v ./...
    make
    go test -v ./...
else
    go get golang.org/x/tools/cmd/cover
    go get github.com/haya14busa/goverage
    go get -t -d -v ./...
    make
    go test -v ./...
    if [[ $? == 0 ]]; then
        goverage -coverprofile cover.out ./...
        go tool cover -func cover.out
    fi
fi

