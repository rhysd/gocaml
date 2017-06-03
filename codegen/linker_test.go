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
		t.Fatalf("No error occurred")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Linker command failed: ") {
		t.Fatalf("Unexpected error message '%s'", msg)
	}
}

func TestMultiGOPATH(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	defer os.Setenv("GOPATH", gopath)
	os.Setenv("GOPATH", "unknown-path:"+gopath)

	l := newDefaultLinker("")
	err := l.link("dummy", []string{"not-exist.o"})
	if !strings.Contains(err.Error(), "Linker command failed: ") {
		t.Fatalf("Unexpected error message '%s'", err.Error())
	}
}

func TestRuntimeNotFound(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	defer os.Setenv("GOPATH", gopath)
	os.Setenv("GOPATH", "/unknown/path/to/somewhere")

	l := newDefaultLinker("")
	err := l.link("dummy", []string{"not-exist.o"})
	if !strings.Contains(err.Error(), "Runtime library (gocamlrt.a) was not found") {
		t.Fatalf("Unexpected error message '%s'", err.Error())
	}
}

func TestCustomizeLinkerCommand(t *testing.T) {
	saved := os.Getenv("GOCAML_LINKER_CMD")
	defer os.Setenv("GOCAML_LINKER_CMD", saved)
	os.Setenv("GOCAML_LINKER_CMD", "linker-command-for-test")
	l := newDefaultLinker("")
	if l.linkerCmd != "linker-command-for-test" {
		t.Fatalf("Wanted 'linker-command-for-test' as linker command but had '%s'", l.linkerCmd)
	}
}
