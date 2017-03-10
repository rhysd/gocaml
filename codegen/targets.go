package codegen

import (
	"llvm.org/llvm/bindings/go/llvm"
)

type Target struct {
	Name        string
	Description string
}

func AllTargets() []Target {
	targets := []Target{}
	for t := llvm.FirstTarget(); t.C != nil; t = t.NextTarget() {
		targets = append(targets, Target{t.Name(), t.Description()})
	}
	return targets
}
