package sema

import (
	"github.com/rhysd/gocaml/types"
	"unsafe"
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

type instantiation struct {
	vars  map[types.GenericId]*types.Var
	level int
}

func (inst *instantiation) apply(t types.Type) types.Type {
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

func instantiate(t types.Type, level int) types.Type {
	i := &instantiation{map[types.GenericId]*types.Var{}, level}
	return i.apply(t)
}
