SRCS := \
	main.go \
	ast/node.go \
	ast/printer.go \
	ast/visitor.go \
	compiler/compiler.go \
	lexer/lexer.go \
	parser/grammar.go \
	parser/parser.go \
	token/source.go \
	token/token.go \
	typing/env.go \
	typing/unify.go \
	typing/deref.go \
	typing/infer.go \
	typing/type.go \
	alpha/transform.go \
	alpha/mapping.go \

TESTS := \
	lexer/lexer_test.go \
	parser/parser_test.go \
	ast/visitor_test.go \
	token/source_test.go \
	token/token_test.go \
	typing/infer_test.go \

all: build test

build: gocaml

gocaml: $(SRCS)
	go build

parser/grammar.go: parser/grammar.go.y
	go get golang.org/x/tools/cmd/goyacc
	go tool yacc -o parser/grammar.go parser/grammar.go.y

test: $(TESTS)
	go test ./...

clean:
	rm -f gocaml y.output parser/grammar.go

.PHONY: all build clean test
