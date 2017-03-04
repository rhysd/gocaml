package codegen

import (
	"github.com/rhysd/gocaml/alpha"
	"github.com/rhysd/gocaml/closure"
	"github.com/rhysd/gocaml/gcil"
	"github.com/rhysd/gocaml/lexer"
	"github.com/rhysd/gocaml/parser"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testCreateEmitter(optimize OptLevel, debug bool) (e *Emitter, err error) {
	s := token.NewDummySource("let rec f x = x + x in println_int (f 42)")
	l := lexer.NewLexer(s)
	go l.Lex()
	root, err := parser.Parse(l.Tokens)
	if err != nil {
		return
	}
	if err = alpha.Transform(root); err != nil {
		return
	}
	env := typing.NewEnv()
	if err = env.ApplyTypeAnalysis(root); err != nil {
		return
	}
	ir, err := gcil.FromAST(root, env)
	if err != nil {
		return
	}
	gcil.ElimRefs(ir, env)
	prog := closure.Transform(ir)
	opts := EmitOptions{optimize, "", "", debug}
	e, err = NewEmitter(prog, env, s, opts)
	if err != nil {
		return
	}
	e.RunOptimizationPasses()
	return
}

func TestEmitLLVMIR(t *testing.T) {
	e, err := testCreateEmitter(OptimizeDefault, false)
	if err != nil {
		t.Fatal(err)
	}
	ir := e.EmitLLVMIR()
	if !strings.Contains(ir, "ModuleID = 'dummy'") {
		t.Fatalf("Module ID is not contained: %s", ir)
	}
	if !strings.Contains(ir, "target datalayout = ") {
		t.Fatalf("Data layout is not contained: %s", ir)
	}
}

func TestEmitAssembly(t *testing.T) {
	e, err := testCreateEmitter(OptimizeDefault, false)
	if err != nil {
		t.Fatal(err)
	}
	asm, err := e.EmitAsm()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(asm, ".section") {
		t.Fatalf("Assembly was not emitted: %s", asm)
	}
}

func TestEmitObject(t *testing.T) {
	e, err := testCreateEmitter(OptimizeDefault, false)
	if err != nil {
		t.Fatal(err)
	}
	obj, err := e.EmitObject()
	if err != nil {
		t.Fatal(err)
	}
	if len(obj) == 0 {
		t.Fatalf("Emitted object file is empty")
	}
}

func TestEmitExecutable(t *testing.T) {
	e, err := testCreateEmitter(OptimizeDefault, false)
	if err != nil {
		t.Fatal(err)
	}
	outfile, err := filepath.Abs("__test_a.out")
	if err != nil {
		panic(err)
	}
	if err := e.EmitExecutable(outfile); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outfile)
	stats, err := os.Stat(outfile)
	if err != nil {
		t.Fatalf("Cannot stat emitted executable: %s", err.Error())
	}
	if stats.IsDir() {
		t.Fatalf("File was not emitted actually")
	}
	if stats.Size() == 0 {
		t.Errorf("Emitted executable is empty")
	}
}

func TestEmitUnoptimizedLLVMIR(t *testing.T) {
	e, err := testCreateEmitter(OptimizeNone, false)
	if err != nil {
		t.Fatal(err)
	}
	ir := e.EmitLLVMIR()
	if !strings.Contains(ir, `define private i64 @"f$t1"(i64 %"x$t2")`) {
		t.Fatalf("Function 'f' was inlined with OptimizeNone config: %s", ir)
	}
}

func TestEmitLLVMIRWithDebugInfo(t *testing.T) {
	e, err := testCreateEmitter(OptimizeNone, true)
	if err != nil {
		t.Fatal(err)
	}
	ir := e.EmitLLVMIR()
	if !strings.Contains(ir, "!llvm.dbg.cu = ") {
		t.Fatalf("Debug information is not contained: %s", ir)
	}
}
