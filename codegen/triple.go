package codegen

import (
	"llvm.org/llvm/bindings/go/llvm"
	"strings"
)

type Triple string

func NewTriple(specified string) Triple {
	if specified == "" {
		return Triple(llvm.DefaultTargetTriple())
	}
	return Triple(specified)
}

// Get archtecture from triple
// Conversion table was from llvm/Support/Triple.h
// http://llvm.org/docs/doxygen/html/Triple_8cpp_source.html
func (triple Triple) Arch() string {
	t := string(triple)
	name := t[:strings.IndexRune(t, '-')]

	switch name {
	case "i386", "i486", "i586", "i686", "i786", "i886", "i986":
		return "x86"
	case "amd64", "x86_64", "x86_64h":
		return "x86-64"
	case "powerpc", "ppc32":
		return "ppc"
	case "powerpc64", "ppu", "ppc64":
		return "ppc64"
	case "arm", "xscale":
		return "arm"
	case "xscaleeb":
		return "armeb"
	case "aarch64", "arm64": // arm64 is a alias for aarch64
		return "aarch64"
	case "aarch64_be":
		return "aarch64_be"
	case "armeb":
		return "armeb"
	case "thumb":
		return "thumb"
	case "thumbeb":
		return "thumbeb"
	case "avr":
		return "avr"
	case "msp430":
		return "msp430"
	case "mips", "mipseb", "mipsallegrex":
		return "mips"
	case "mipsel", "mipsallegrexel":
		return "mipsel"
	case "mips64", "mips64eb":
		return "mips64"
	case "mips64el":
		return "mips64el"
	case "s390", "systemz":
		return "systemz"
	case "sparcv9", "sparc64":
		return "sparcv9"
	case "bpf_be", "bpfeb":
		return "bpfeb"
	case "bpf_le", "bpfel":
		return "bpfel"
	case "r600", "hexagon", "sparc", "sparcel", "tce", "xcore", "nvptx", "nvptx64",
		"le32", "le64", "amdil", "amdil64", "hsail", "hsail64", "spir", "spir64",
		"shave", "wasm32", "wasm64":
		return name
	}

	// TODO:
	// If name starts with "arm", "thumb" or "aarch64", it should be ARM family.
	// In the case, we need to check endian to determine the architecture.

	// TODO:
	// If name is "bpf", we need to see endian to determine "bpfeb" or "bpfel"

	return "unknown"
}
