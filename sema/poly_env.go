package sema

import (
	"fmt"
	"github.com/rhysd/gocaml/types"
	"strings"
)

type varVariants map[types.VarID][]types.Type

type polyEnv struct {
	parent *polyEnv
	vars   varVariants
}

func (poly *polyEnv) variants(id types.VarID) (found []types.Type) {
	if poly.parent != nil {
		found = poly.variants(id)
	}
	if ts, ok := poly.vars[id]; ok {
		found = append(found, ts...)
	}
	return
}

func (poly *polyEnv) variantExists(id types.VarID, v types.Type) bool {
	if ts, ok := poly.vars[id]; ok {
		for _, t := range ts {
			if types.Equals(t, v) {
				return true
			}
		}
	}
	if poly.parent != nil {
		return poly.parent.variantExists(id, v)
	}
	return false
}

func (poly *polyEnv) addPoly(id types.VarID, t types.Type) {
	if !poly.variantExists(id, t) {
		poly.vars[id] = append(poly.vars[id], t)
	}
}

func (poly *polyEnv) nest() *polyEnv {
	return &polyEnv{
		poly,
		varVariants{},
	}
}

func (poly *polyEnv) merge(other, hint *polyEnv) *polyEnv {
	child := poly.nest()
	for other != nil && other != hint {
		for id, ts := range other.vars {
			for _, t := range ts {
				child.addPoly(id, t)
			}
		}
		other = other.parent
	}
	return child
}

func (poly *polyEnv) varsString() string {
	ss := make([]string, 0, len(poly.vars)+1)
	if poly.parent != nil {
		ss = append(ss, poly.parent.varsString())
	}
	for id, ts := range poly.vars {
		v := make([]string, 0, len(ts))
		for _, t := range ts {
			v = append(v, types.Debug(t))
		}
		ss = append(ss, fmt.Sprintf("  '%d => %s", id, strings.Join(v, " | ")))
	}
	return strings.Join(ss, "\n")
}

func (poly *polyEnv) String() string {
	return "polyEnv {\n" + poly.varsString() + "\n}"
}

func (poly *polyEnv) flatten() (flat varVariants) {
	if poly.parent == nil {
		flat = varVariants{}
	} else {
		flat = poly.parent.flatten()
	}
	for id, ts := range poly.vars {
		flat[id] = append(flat[id], ts...)
	}
	return
}

func (variants varVariants) String() string {
	vars := make([]string, 0, len(variants))
	for id, ts := range variants {
		vs := make([]string, 0, len(ts))
		for _, t := range ts {
			vs = append(vs, types.Debug(t))
		}
		vars = append(vars, fmt.Sprintf("\n  '%d => %s", id, strings.Join(vs, " | ")))
	}
	return "variants {" + strings.Join(vars, "") + "\n}"
}
