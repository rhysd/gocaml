package codegen

import (
	"testing"
)

func TestAllTargets(t *testing.T) {
	targets := AllTargets()
	if len(targets) == 0 {
		t.Fatalf("No target was found")
	}
}
