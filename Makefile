SRCS := \
	ast/node.go \
	ast/printer.go \
	ast/visitor.go \
	lexer/lexer.go \
	parser/grammar.go \
	parser/parser.go \
	token/token.go \
	token/source.go \

all: gobuild goyacc

gobuild: $(SRCS)
	go build

goyacc: parser/grammar.go.y
	goyacc -o parser/grammar.go parser/grammar.go.y

test:
	go test ./...

clean:
	rm -f gocaml y.output parser/grammar.go

.PHONY: all clean test
