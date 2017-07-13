package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/types"
	"strings"
)

type polyVariants map[types.Type][]*types.Instantiation

type polyVariantsContext struct {
	parent *polyVariantsContext
	insts  polyVariants
}

func (ctx *polyVariantsContext) exists(inst *types.Instantiation) bool {
	if insts, ok := ctx.insts[inst.From]; ok {
		for _, i := range insts {
			if types.Equals(inst.To, i.To) {
				return true
			}
		}
	}
	if ctx.parent != nil {
		return ctx.parent.exists(inst)
	}
	return false
}

func (ctx *polyVariantsContext) add(inst *types.Instantiation) {
	if ctx.exists(inst) {
		return
	}
	ctx.insts[inst.From] = append(ctx.insts[inst.From], inst)
}

func (ctx *polyVariantsContext) allVariants() (variants polyVariants) {
	if ctx.parent == nil {
		variants = polyVariants{}
	} else {
		variants = ctx.parent.allVariants()
	}
	for from, insts := range ctx.insts {
		variants[from] = append(variants[from], insts...)
	}
	return
}

func (ctx *polyVariantsContext) variantsOf(t types.Type) (insts []*types.Instantiation) {
	if ctx.parent == nil {
		insts = make([]*types.Instantiation, 0, 3)
	} else {
		insts = ctx.parent.variantsOf(t)
	}
	if i, ok := ctx.insts[t]; ok {
		insts = append(insts, i...)
	}
	return
}

func (ctx *polyVariantsContext) nest() *polyVariantsContext {
	return &polyVariantsContext{ctx, polyVariants{}}
}

func (ctx *polyVariantsContext) merge(others ...*polyVariantsContext) *polyVariantsContext {
	current := ctx
	for _, other := range others {
		current = current.nest()
		for other != nil && other != ctx {
			for _, insts := range other.insts {
				for _, i := range insts {
					current.add(i)
				}
			}
			other = other.parent
		}
	}
	return current
}

func (poly polyVariants) String() string {
	lines := make([]string, 0, len(poly))
	for from, insts := range poly {
		ss := make([]string, 0, len(insts))
		for _, i := range insts {
			ss = append(ss, types.Debug(i.To))
		}
		lines = append(lines, fmt.Sprintf("\n  %s => %s", types.Debug(from), strings.Join(ss, " | ")))
	}
	return "polyVariants {" + strings.Join(lines, "") + "\n}"
}
