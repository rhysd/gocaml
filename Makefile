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
	gcil/val.go \
	gcil/block.go \
	gcil/emit_ir.go \
	gcil/visitor.go \
	gcil/printer.go \
	gcil/elim_ref.go \
	gcil/program.go \
	closure/transform.go \
	closure/freevars.go \
	codegen/emitter.go \
	codegen/module_builder.go \

TESTS := \
	alpha/example_test.go \
	alpha/mapping_test.go \
	alpha/transform_test.go \
	ast/example_test.go \
	ast/visitor_test.go \
	ast/printer_test.go \
	closure/example_test.go \
	closure/transform_test.go \
	compiler/example_test.go \
	lexer/example_test.go \
	lexer/lexer_test.go \
	parser/example_test.go \
	parser/parser_test.go \
	token/source_test.go \
	token/token_test.go \
	typing/env_test.go \
	typing/example_test.go \
	typing/infer_test.go \
	gcil/example_test.go \
	gcil/block_test.go \
	gcil/emit_ir_test.go \
	gcil/visitor_test.go \
	gcil/elim_ref_test.go \
	gcil/program_test.go \

all: build test

build: gocaml

gocaml: $(SRCS)
	./install_llvmgo.sh
	time go build

parser/grammar.go: parser/grammar.go.y
	go get golang.org/x/tools/cmd/goyacc
	go tool yacc -o parser/grammar.go parser/grammar.go.y

test: $(TESTS)
	go test ./...

cover.out: $(TESTS)
	go get github.com/haya14busa/goverage
	goverage -coverprofile cover.out ./alpha ./ast ./gcil ./closure ./lexer ./parser ./token ./typing

cov: cover.out
	go get golang.org/x/tools/cmd/cover
	go tool cover -html=cover.out

clean:
	rm -f gocaml y.output parser/grammar.go

.PHONY: all build clean test cov
