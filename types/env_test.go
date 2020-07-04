package types

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestDumpResult(t *testing.T) {
	env := NewEnv()
	env.DeclTable["test_ident"] = IntType
	env.DeclTable["test_ident2"] = BoolType
	env.DeclTable["external_ident"] = UnitType
	env.DeclTable["external_ident2"] = FloatType

	// TODO: Add dummy instantiations

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	env.Dump()

	ch := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		ch <- buf.String()
	}()
	w.Close()
	os.Stdout = old

	out := <-ch
	for _, s := range []string{
		"Variables:\n",
		"test_ident: int",
		"test_ident2: bool",
		"External Variables:\n",
		"external_ident: unit",
		"external_ident2: float",
	} {
		if !strings.Contains(out, s) {
			t.Fatalf("Output does not contain '%s': %s", s, out)
		}
	}
}

// TODO: TestDumpDebug

func TestEnvHasBuiltins(t *testing.T) {
	env := NewEnv()
	if len(env.Externals) == 0 {
		t.Fatal("Env must contain some external symbols by default because of builtin symbols")
	}
	if _, ok := env.Externals["print_int"]; !ok {
		t.Fatal("'print_int' is not found though it is builtin:", env.Externals)
	}
}
