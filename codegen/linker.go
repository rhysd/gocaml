package codegen

import (
	"fmt"
	"os/exec"
	"strings"
)

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
	args := append(objFiles, "-o", executable, lnk.ldflags)
	_, err := exec.Command(lnk.linkerCmd, args...).Output()
	if exiterr, ok := err.(*exec.ExitError); ok {
		return lnk.makeError(args, string(exiterr.Stderr))
	}
	if err != nil {
		return lnk.makeError(args, err.Error())
	}
	return nil
}
