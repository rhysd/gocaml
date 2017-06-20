package sema

import (
	"github.com/rhysd/gocaml/types"
)

func generalize(level int, t types.Type) types.Type {
	switch t := t.(type) {
	case *types.Var:
		if t.Ref != nil {
			return generalize(level, t.Ref)
		}
		if t.Level > level {
			// Bind free variable 'a' as 'forall a.a'
			t.AsGeneric()
		}
	case *types.Tuple:
		for i, e := range t.Elems {
			t.Elems[i] = generalize(level, e)
		}
	case *types.Array:
		t.Elem = generalize(level, t.Elem)
	case *types.Option:
		t.Elem = generalize(level, t.Elem)
	case *types.Fun:
		t.Ret = generalize(level, t.Ret)
		for i, p := range t.Params {
			t.Params[i] = generalize(level, p)
		}
	}
	return t
}

type instantiator struct {
	vars  map[*types.Var]*types.Var
	level int
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
		v, ok := inst.vars[t]
		if !ok {
			v = &types.Var{Level: inst.level}
			inst.vars[t] = v
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

type Instantiation struct {
	From    types.Type
	To      types.Type
	Mapping map[*types.Var]*types.Var
}

func instantiate(t types.Type, level int) *Instantiation {
	i := &instantiator{map[*types.Var]*types.Var{}, level}
	ret := i.apply(t)
	if len(i.vars) == 0 {
		// Should return the original type 't' here?
		// Even if no instantiation occurred, linked type variables may be dereferenced in instantiator.apply().
		return nil
	}
	return &Instantiation{
		From:    t,
		To:      ret,
		Mapping: i.vars,
	}
}
