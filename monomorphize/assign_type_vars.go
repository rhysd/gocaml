package monomorphize

import (
	"github.com/rhysd/gocaml/types"
)

type assignments map[types.VarID]types.Type

type assignTypeVars struct {
	mapping assignments
}

func (assign *assignTypeVars) assignToTypes(ts []types.Type) ([]types.Type, bool) {
	assigned := make([]types.Type, 0, len(ts))
	changed := false
	for _, t := range ts {
		t, c := assign.assign(t)
		changed = changed || c
		assigned = append(assigned, t)
	}
	return assigned, changed
}

func (assign *assignTypeVars) assignToVar(v *types.Var) types.Type {
	if v.Ref != nil {
		t, _ := assign.assign(v.Ref)
		return t
	}
	t, ok := assign.mapping[v.ID]
	if !ok {
		panic("FATAL: Unknown type variable")
	}
	ret, _ := assign.assign(t)
	return ret
}

func (assign *assignTypeVars) assign(t types.Type) (types.Type, bool) {
	switch t := t.(type) {
	case *types.Fun:
		ret, changed := assign.assign(t.Ret)
		params, changed2 := assign.assignToTypes(t.Params)
		if changed || changed2 {
			return &types.Fun{ret, params}, true
		}
	case *types.Tuple:
		elems, changed := assign.assignToTypes(t.Elems)
		if changed {
			return &types.Tuple{elems}, true
		}
	case *types.Array:
		elem, changed := assign.assign(t.Elem)
		if changed {
			return &types.Array{elem}, true
		}
	case *types.Option:
		elem, changed := assign.assign(t.Elem)
		if changed {
			return &types.Option{elem}, true
		}
	case *types.Var:
		return assign.assignToVar(t), true
	}
	return t, false
}
