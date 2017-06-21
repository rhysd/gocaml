package sema

import (
	"github.com/rhysd/gocaml/types"
)

// Generalization cannot modify given type destructively. It sometimes breaks existing other type variable.
// For example, in `let rec f x = f true; x in f 42`, first function call `f true` is inferred as `bool -> 'a`
// after type inference because `f true;` is equivalent to `let $unused = f true in`. 'let' expression causes
// generalization and $unused is inferred as 'a. It changes type of the `f` from `bool -> ?` to `bool -> 'a`
// if generalization changed a given type destructively.
func generalize(level int, t types.Type) types.Type {
	switch t := t.(type) {
	case *types.Var:
		if t.Ref != nil {
			return generalize(level, t.Ref)
		}
		if t.Level > level {
			// Bind free variable 'a' as 'forall a.a'
			return t.AsGeneric()
		}
		return t
	case *types.Tuple:
		ts := make([]types.Type, 0, len(t.Elems))
		for _, e := range t.Elems {
			ts = append(ts, generalize(level, e))
		}
		return &types.Tuple{ts}
	case *types.Array:
		return &types.Array{generalize(level, t.Elem)}
	case *types.Option:
		return &types.Option{generalize(level, t.Elem)}
	case *types.Fun:
		ret := generalize(level, t.Ret)
		params := make([]types.Type, 0, len(t.Params))
		for _, p := range t.Params {
			params = append(params, generalize(level, p))
		}
		return &types.Fun{ret, params}
	default:
		return t
	}
}

type instantiator struct {
	vars  map[types.VarID]*types.Var
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
		v, ok := inst.vars[t.ID]
		if !ok {
			v = types.NewVar(nil, inst.level)
			inst.vars[t.ID] = v
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
	i := &instantiator{map[types.VarID]*types.Var{}, level}
	ret := i.apply(t)
	if len(i.vars) == 0 {
		// Should return the original type 't' here?
		// Even if no instantiation occurred, linked type variables may be dereferenced in instantiator.apply().
		return nil
	}
	return &types.Instantiation{
		From:    t,
		To:      ret,
		Mapping: i.vars,
	}
}
