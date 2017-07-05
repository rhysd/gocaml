package monomorphize

import (
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
)

type codeDup struct {
	env *types.Env
}

func (dup *codeDup) insn(i *mir.Insn) *mir.Insn {
	return i
}

func (dup *codeDup) block(block *mir.Block) *mir.Block {
	begin, _ := block.WholeRange()
	insn := dup.insn(begin)
	// TODO: Do not consume O(n) to get end of block
	return mir.NewBlock(block.Name, insn, insn.Last())
}

func CodeDup(from *mir.Block, env *types.Env) *mir.Block {
	dup := &codeDup{env}
	return dup.block(from)
}
