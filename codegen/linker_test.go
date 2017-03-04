package codegen

import (
	"strings"
	"testing"
)

func TestLinkFailed(t *testing.T) {
	l := newDefaultLinker("")
	err := l.link("dummy", []string{"not-exist.o"})
	if err == nil {
		t.Fatalf("Expected error not occurred")
	}
	if !strings.HasPrefix(err.Error(), "Linker command failed: ") {
		t.Fatalf("Unexpected error message '%s'", err.Error())
	}
}
