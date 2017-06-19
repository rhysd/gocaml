package sema

import (
	"unsafe"

	"github.com/rhysd/gocaml/types"
)

func Generalize(level int, t types.Type) types.Type {
	switch t := t.(type) {
	case *types.Var:
		if t.Ref != nil {
			return Generalize(level, t.Ref)
		}
		if t.Level > level {
			// Bind free variable 'a' as 'forall a.a'
			return &types.Generic{types.GenericId(unsafe.Pointer(t))}
		}
	case *types.Tuple:
		for i, e := range t.Elems {
			t.Elems[i] = Generalize(level, e)
		}
	case *types.Array:
		t.Elem = Generalize(level, t.Elem)
	case *types.Option:
		t.Elem = Generalize(level, t.Elem)
	case *types.Fun:
		t.Ret = Generalize(level, t.Ret)
		for i, p := range t.Params {
			t.Params[i] = Generalize(level, p)
		}
	}
	return t
}

type instantiator struct {
	vars  map[types.GenericId]*types.Var
	level int
}

func (inst *instantiator) apply(t types.Type) types.Type {
	switch t := t.(type) {
	case *types.Var:
		if t.Ref != nil {
			return inst.apply(t.Ref)
		}
		return t
	case *types.Generic:
		v, ok := inst.vars[t.Id]
		if !ok {
			v = &types.Var{Level: inst.level}
			inst.vars[t.Id] = v
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
	Mapping map[types.GenericId]*types.Var
}

func instantiate(t types.Type, level int) *Instantiation {
	i := &instantiator{map[types.GenericId]*types.Var{}, level}
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
