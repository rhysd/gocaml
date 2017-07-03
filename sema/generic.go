package sema

import (
	"github.com/rhysd/gocaml/types"
)

type boundVarIDs map[types.VarID]struct{}

func (ids boundVarIDs) add(id types.VarID) {
	ids[id] = struct{}{}
}

func (ids boundVarIDs) contains(id types.VarID) bool {
	_, ok := ids[id]
	return ok
}

type generalizer struct {
	bounds boundVarIDs
	level  int
}

func (gen *generalizer) apply(t types.Type) types.Type {
	switch t := t.(type) {
	case *types.Var:
		if t.Ref != nil {
			return gen.apply(t.Ref)
		}
		if t.Level > gen.level {
			gen.bounds.add(t.ID)
			t.SetGeneric()
		}
		return t
	case *types.Tuple:
		elems := make([]types.Type, 0, len(t.Elems))
		for _, e := range t.Elems {
			elems = append(elems, gen.apply(e))
		}
		return &types.Tuple{elems}
	case *types.Array:
		return &types.Array{gen.apply(t.Elem)}
	case *types.Option:
		return &types.Option{gen.apply(t.Elem)}
	case *types.Fun:
		params := make([]types.Type, 0, len(t.Params))
		for _, p := range t.Params {
			params = append(params, gen.apply(p))
		}
		return &types.Fun{gen.apply(t.Ret), params}
	default:
		return t
	}
}

// Generalize given type variable. It means binding proper free type variables in the type. It returns
// generalized type and IDs of bound type variables in given type.
func generalize(t types.Type, level int) (types.Type, boundVarIDs) {
	gen := &generalizer{boundVarIDs{}, level}
	t = gen.apply(t)
	return t, gen.bounds
}

type instantiator struct {
	freeVars []*types.VarMapping
	level    int
}

func (inst *instantiator) apply(t types.Type) types.Type {
	switch t := t.(type) {
	case *types.Var:
		if t.Ref != nil {
			return inst.apply(t.Ref)
		}
		if !t.IsGeneric() {
			return t
		}
		for _, m := range inst.freeVars {
			if t.ID == m.ID {
				return m.Type
			}
		}
		v := types.NewVar(nil, inst.level)
		inst.freeVars = append(inst.freeVars, &types.VarMapping{t.ID, v})
		return v
	case *types.Tuple:
		ts := make([]types.Type, 0, len(t.Elems))
		for _, e := range t.Elems {
			ts = append(ts, inst.apply(e))
		}
		return &types.Tuple{ts}
	case *types.Array:
		return &types.Array{inst.apply(t.Elem)}
	case *types.Option:
		return &types.Option{inst.apply(t.Elem)}
	case *types.Fun:
		ts := make([]types.Type, 0, len(t.Params))
		for _, p := range t.Params {
			ts = append(ts, inst.apply(p))
		}
		return &types.Fun{inst.apply(t.Ret), ts}
	default:
		return t
	}
}

func instantiate(t types.Type, level int) *types.Instantiation {
	i := &instantiator{[]*types.VarMapping{}, level}
	ret := i.apply(t)
	if len(i.freeVars) == 0 {
		// Should return the original type 't' here?
		// Even if no instantiation occurred, linked type variables may be dereferenced in instantiator.apply().
		return nil
	}
	return &types.Instantiation{
		From:    t,
		To:      ret,
		Mapping: i.freeVars,
	}
}
