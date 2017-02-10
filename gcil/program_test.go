package gcil

import (
	"bytes"
	"github.com/rhysd/gocaml/typing"
	"strings"
	"testing"
)

func TestDump(t *testing.T) {
	prog := &Program{
		map[string]*Fun{},
		map[string][]string{},
		NewBlockFromArray("program", []*Insn{
			NewInsn("$k1", UnitVal),
		}),
	}

	env := typing.NewEnv()
	env.Table["$k1"] = typing.UnitType

	var buf bytes.Buffer
	prog.Dump(&buf, env)
	out := buf.String()
	if !strings.Contains(out, "[TOPLEVELS (0)]") {
		t.Fatalf("Toplevel section not found")
	}
	if !strings.Contains(out, "[CLOSURES (0)]") {
		t.Fatalf("Closures section not found")
	}
	if !strings.Contains(out, "[ENTRY]") {
		t.Fatalf("Entry section not found")
	}
}
