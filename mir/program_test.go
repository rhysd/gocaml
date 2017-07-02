package mir

import (
	"bytes"
	"github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
	"strings"
	"testing"
)

func TestDump(t *testing.T) {
	prog := &Program{
		NewToplevel(),
		map[string][]string{},
		NewBlockFromArray("program", []*Insn{
			NewInsn("$k1", UnitVal, locerr.Pos{}),
		}),
	}

	env := types.NewEnv()
	env.DeclTable["$k1"] = types.UnitType

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
