package codegen

import (
	"fmt"
	"llvm.org/llvm/tools/llgo/third_party/gofrontend/libgo/go/path/filepath"
	"os"
	"os/exec"
	"strings"
)

func detectRuntimePath() (string, error) {
	// XXX:
	// Need to investigate solid way to get runtime library path

	fromBuildDir, err := filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), "runtime/gocamlrt.a"))
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(fromBuildDir); err == nil {
		return fromBuildDir, nil
	}

	candidates := []string{fromBuildDir}

	gopaths := strings.Split(os.Getenv("GOPATH"), ":")
	for _, gopath := range gopaths {
		fromGopath := filepath.Join(gopath, "src/github.com/rhysd/gocaml/runtime/gocamlrt.a")
		if _, err := os.Stat(fromGopath); err == nil {
			return fromGopath, nil
		}
		candidates = append(candidates, fromGopath)
	}

	return "", fmt.Errorf("Runtime library (gocamlrt.a) was not found\nCandidates:\n%s", strings.Join(candidates, "\n"))
}

type linker struct {
	linkerCmd string
	ldflags   string
}

func newDefaultLinker(ldflags string) *linker {
	return &linker{"clang", ldflags}
}

func (lnk *linker) makeError(args []string, msg string) error {
	return fmt.Errorf("Linker command failed: %s %s:\n%s", lnk.linkerCmd, strings.Join(args, " "), msg)
}

func (lnk *linker) link(executable string, objFiles []string) error {
	// TODO: Consider Windows environment

	runtimePath, err := detectRuntimePath()
	if err != nil {
		return err
	}

	args := append(objFiles, "-o", executable, runtimePath, lnk.ldflags)

	if _, err := exec.Command(lnk.linkerCmd, args...).Output(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return lnk.makeError(args, string(exiterr.Stderr))
		}
		return lnk.makeError(args, err.Error())
	}

	return nil
}
