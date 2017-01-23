SRCS := \
	ast/node.go \
	ast/printer.go \
	ast/visitor.go \
	lexer/lexer.go \
	parser/grammar.go \
	parser/parser.go \
	token/token.go \
	token/source.go \

TESTS := \
	lexer/lexer_test.go \

all: build test

build: gocaml

gocaml: $(SRCS)
	go build

parser/grammar.go: parser/grammar.go.y
	go get golang.org/x/tools/cmd/goyacc
	go tool yacc -o parser/grammar.go parser/grammar.go.y

test: $(TESTS)
	go test -v ./...

clean:
	rm -f gocaml y.output parser/grammar.go

.PHONY: all build clean test
