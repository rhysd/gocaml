package codegen

import (
	"os"
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

func TestMultiGOPATH(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	defer os.Setenv("GOPATH", gopath)
	os.Setenv("GOPATH", "unknown-path:"+gopath)

	l := newDefaultLinker("")
	err := l.link("dummy", []string{"not-exist.o"})
	if !strings.HasPrefix(err.Error(), "Linker command failed: ") {
		t.Fatalf("Unexpected error message '%s'", err.Error())
	}
}

func TestRuntimeNotFound(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	defer os.Setenv("GOPATH", gopath)
	os.Setenv("GOPATH", "")

	l := newDefaultLinker("")
	err := l.link("dummy", []string{"not-exist.o"})
	if !strings.HasPrefix(err.Error(), "Runtime library (gocamlrt.a) was not found") {
		t.Fatalf("Unexpected error message '%s'", err.Error())
	}
}
