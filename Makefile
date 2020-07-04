SRCS := \
	main.go \
	ast/node.go \
	ast/printer.go \
	ast/visitor.go \
	driver/driver.go \
	syntax/lexer.go \
	syntax/grammar.go \
	syntax/parser.go \
	token/token.go \
	types/builtins.go \
	types/env.go \
	types/type.go \
	types/visitor.go \
	types/equals.go \
	sema/unify.go \
	sema/generic.go \
	sema/deref.go \
	sema/infer.go \
	sema/node_to_type.go \
	sema/semantics_check.go \
	sema/to_mir.go \
	sema/alpha_transform.go \
	sema/scope.go \
	mir/val.go \
	mir/block.go \
	mir/printer.go \
	mir/program.go \
	closure/transform.go \
	closure/freevars.go \
	closure/fix_apps.go \
	mono/monomorphize.go \
	codegen/emitter.go \
	codegen/module_builder.go \
	codegen/type_builder.go \
	codegen/block_builder.go \
	codegen/debug_info_builder.go \
	codegen/linker.go \
	codegen/targets.go \
	common/ordinal.go \

TESTS := \
	ast/example_test.go \
	ast/visitor_test.go \
	ast/printer_test.go \
	closure/example_test.go \
	closure/transform_test.go \
	driver/example_test.go \
	syntax/lexer_test.go \
	syntax/example_test.go \
	syntax/parser_test.go \
	token/token_test.go \
	types/env_test.go \
	types/type_test.go \
	types/visitor_test.go \
	sema/example_test.go \
	sema/infer_test.go \
	sema/deref_test.go \
	sema/node_to_type_test.go \
	sema/to_mir_test.go \
	sema/semantics_check_test.go \
	sema/scope_test.go \
	sema/alpha_transform_test.go \
	sema/algorithm_w_test.go \
	mir/block_test.go \
	mir/program_test.go \
	codegen/example_test.go \
	codegen/executable_test.go \
	codegen/linker_test.go \
	codegen/targets_test.go \
	common/ordinal_test.go \

all: build test

build: gocaml runtime/gocamlrt.a

gocaml: $(SRCS)
	./scripts/install_llvmgo.sh
	go get -t -d ./...
	if which time > /dev/null; then\
		CGO_LDFLAGS_ALLOW='-Wl,(-search_paths_first|-headerpad_max_install_names)' time go build;\
	else\
		CGO_LDFLAGS_ALLOW='-Wl,(-search_paths_first|-headerpad_max_install_names)' go build;\
	fi

syntax/grammar.go: syntax/grammar.go.y
	go get golang.org/x/tools/cmd/goyacc
	goyacc -o syntax/grammar.go syntax/grammar.go.y

runtime/gocamlrt.o: runtime/gocamlrt.c runtime/gocaml.h
	$(CC) -Wall -Wextra -std=c99 -I/usr/local/include -I./runtime $(CFLAGS) -c runtime/gocamlrt.c -o runtime/gocamlrt.o
runtime/gocamlrt.a: runtime/gocamlrt.o
	ar -r runtime/gocamlrt.a runtime/gocamlrt.o

test: $(TESTS)
ifdef VERBOSE
	CGO_LDFLAGS_ALLOW='-Wl,(-search_paths_first|-headerpad_max_install_names)' go test -v ./...
else
	CGO_LDFLAGS_ALLOW='-Wl,(-search_paths_first|-headerpad_max_install_names)' go test ./...
endif

cover.out: $(TESTS)
	go get github.com/haya14busa/goverage
	CGO_LDFLAGS_ALLOW='-Wl,(-search_paths_first|-headerpad_max_install_names)' goverage -coverprofile=cover.out -covermode=count ./ast ./mir ./closure ./syntax ./token ./sema ./codegen ./common ./mono

cov: cover.out
	go get golang.org/x/tools/cmd/cover
	go tool cover -html=cover.out

cpu.prof codegen.test: $(SRCS) codegen/executable_test.go
	CGO_LDFLAGS_ALLOW='-Wl,(-search_paths_first|-headerpad_max_install_names)' go test -cpuprofile cpu.prof -bench . -run '^$$' ./codegen

prof: cpu.prof codegen.test
	go tool pprof codegen.test cpu.prof

prof.png: cpu.prof codegen.test
	go tool pprof -png codegen.test cpu.prof > prof.png

gocaml-darwin-x86_64.zip: gocaml runtime/gocamlrt.a
	rm -rf gocaml-darwin-x86_64 gocaml-darwin-x86_64.zip
	mkdir -p gocaml-darwin-x86_64/runtime
	mkdir -p gocaml-darwin-x86_64/include
	cp gocaml gocaml-darwin-x86_64/
	cp runtime/gocamlrt.a gocaml-darwin-x86_64/runtime/
	cp runtime/gocaml.h gocaml-darwin-x86_64/include/
	cp README.md LICENSE gocaml-darwin-x86_64/
	zip gocaml-darwin-x86_64.zip -r gocaml-darwin-x86_64
	rm -rf gocaml-darwin-x86_64

release: gocaml-darwin-x86_64.zip

clean:
	rm -f gocaml y.output syntax/grammar.go runtime/gocamlrt.o runtime/gocamlrt.a cover.out cpu.prof codegen.test prof.png gocaml-darwin-x86_64.zip

.PHONY: all build clean test cov prof release
