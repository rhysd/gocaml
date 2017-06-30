package types

// Equals returns given two types are equivalent or not. Note that type variable's ID and level are
// not seen, but free or bound (.IsGeneric() or not) is seen.
func Equals(l, r Type) bool {
	switch l := l.(type) {
	case *Unit, *Int, *Float, *Bool, *String:
		return l == r
	case *Tuple:
		r, ok := r.(*Tuple)
		if !ok || len(l.Elems) != len(r.Elems) {
			return false
		}
		for i, e := range l.Elems {
			if !Equals(e, r.Elems[i]) {
				return false
			}
		}
		return true
	case *Array:
		r, ok := r.(*Array)
		if !ok {
			return false
		}
		return Equals(l.Elem, r.Elem)
	case *Fun:
		r, ok := r.(*Fun)
		if !ok || !Equals(l.Ret, r.Ret) || len(l.Params) != len(r.Params) {
			return false
		}
		for i, p := range l.Params {
			if !Equals(p, r.Params[i]) {
				return false
			}
		}
		return true
	case *Var:
		r, ok := r.(*Var)
		if !ok {
			return false
		}
		if l.Ref == nil && r.Ref == nil {
			lgen, rgen := l.IsGeneric(), r.IsGeneric()
			if lgen && rgen {
				return l.ID == r.ID
			}
			return !lgen && !rgen
		}
		if l.Ref == nil || r.Ref == nil {
			return false
		}
		return Equals(l.Ref, r.Ref)
	case *Option:
		r, ok := r.(*Option)
		if !ok {
			return false
		}
		return Equals(l.Elem, r.Elem)
	default:
		panic("Unreachable")
	}
}
