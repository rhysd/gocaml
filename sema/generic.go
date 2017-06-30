package sema

import (
	"github.com/rhysd/gocaml/types"
)

type boundIDs map[types.VarID]struct{}

func (ids boundIDs) add(id types.VarID) {
	ids[id] = struct{}{}
}

func (ids boundIDs) contains(id types.VarID) bool {
	_, ok := ids[id]
	return ok
}

type generalizer struct {
	bounds boundIDs
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
func generalize(t types.Type, level int) (types.Type, boundIDs) {
	gen := &generalizer{boundIDs{}, level}
	t = gen.apply(t)
	return t, gen.bounds
}

type instantiator struct {
	freeVars map[types.VarID]types.Type
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
		v, ok := inst.freeVars[t.ID]
		if !ok {
			v = types.NewVar(nil, inst.level)
			inst.freeVars[t.ID] = v
		}
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
	i := &instantiator{map[types.VarID]types.Type{}, level}
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
