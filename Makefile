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
	gcil/from_ast.go \
	gcil/printer.go \
	gcil/elim_ref.go \
	gcil/program.go \
	closure/transform.go \
	closure/freevars.go \
	closure/post_process.go \
	codegen/emitter.go \
	codegen/module_builder.go \
	codegen/type_builder.go \
	codegen/block_builder.go \
	codegen/debug_info_builder.go \
	codegen/linker.go \

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
	typing/deref_test.go \
	gcil/example_test.go \
	gcil/block_test.go \
	gcil/from_ast_test.go \
	gcil/elim_ref_test.go \
	gcil/program_test.go \
	codegen/example_test.go \

all: build test

build: gocaml runtime/gocamlrt.a

gocaml: $(SRCS)
	./scripts/install_llvmgo.sh
	go get -t -d ./...
	if which time > /dev/null; then time go build; else go build; fi

parser/grammar.go: parser/grammar.go.y
	go get golang.org/x/tools/cmd/goyacc
	goyacc -o parser/grammar.go parser/grammar.go.y

runtime/gocamlrt.o: runtime/gocamlrt.c runtime/gocaml.h
	$(CC) -Wall -Wextra -pedantic -I./runtime -c runtime/gocamlrt.c -o runtime/gocamlrt.o
runtime/gocamlrt.a: runtime/gocamlrt.o
	ar -r runtime/gocamlrt.a runtime/gocamlrt.o

test: $(TESTS)
	go test ./...

cover.out: $(TESTS)
	go get github.com/haya14busa/goverage
	goverage -coverprofile cover.out ./alpha ./ast ./gcil ./closure ./lexer ./parser ./token ./typing ./codegen

cov: cover.out
	go get golang.org/x/tools/cmd/cover
	go tool cover -html=cover.out

clean:
	rm -f gocaml y.output parser/grammar.go runtime/gocamlrt.o runtime/gocamlrt.a cover.out cpu.out

.PHONY: all build clean test cov
