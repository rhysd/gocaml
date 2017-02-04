package gcil

import (
	"fmt"
	"github.com/rhysd/gocaml/typing"
	"io"
	"strings"
)

type printer struct {
	types  *typing.Env
	out    io.Writer
	indent string
}

func (p *printer) getTypeNameOf(insn *Insn) string {
	t, ok := p.types.Table[insn.Ident]
	if !ok {
		panic(fmt.Sprintf("Type for identifier '%s' not found", insn.Ident))
	}
	if t == nil {
		if strings.HasPrefix(insn.Ident, "$unused") {
			return "unknown (unused)"
		}
		panic("ERROR!: " + insn.Ident)
	}
	return t.String()
}

func (p *printer) printlnInsn(insn *Insn) {
	fmt.Fprintf(p.out, "%s%s = ", p.indent, insn.Ident)
	insn.Val.Print(p.out)
	fmt.Fprintf(p.out, " ; type=%s\n", p.getTypeNameOf(insn))
	switch i := insn.Val.(type) {
	case *If:
		indented := printer{p.types, p.out, p.indent + "  "}
		indented.printlnBlock(i.Then)
		indented.printlnBlock(i.Else)
	case *Fun:
		indented := printer{p.types, p.out, p.indent + "  "}
		indented.printlnBlock(i.Body)
	}
}

func (p *printer) printlnBlock(b *Block) {
	fmt.Fprintf(p.out, "%sBEGIN: %s\n", p.indent, b.Name)
	for i := b.Top.Next; i.Next != nil; i = i.Next {
		p.printlnInsn(i)
	}
	fmt.Fprintf(p.out, "%sEND: %s\n", p.indent, b.Name)
}

func (b *Block) Println(out io.Writer, types *typing.Env) {
	p := printer{
		types,
		out,
		"",
	}
	p.printlnBlock(b)
}
