package types

// Visitor is an interface for the structs which is used for traversing Type.
type Visitor interface {
	// VisitTopdown defines the process when a type is visited. This method is called before
	// children are visited.
	// Returned value is a next visitor to use for succeeding visit. When wanting to stop
	// visiting, please return nil.
	// A visitor visits in depth-first order.
	VisitTopdown(t Type) Visitor
	// VisitBottomup defines the process when a type is visited. This method is called after
	// children were visited. When VisitTopdown returned nil, this method won't be caled for the type.
	VisitBottomup(t Type)
}

// Visit visits the given type with the visitor.
func Visit(vis Visitor, t Type) {
	v := vis.VisitTopdown(t)
	if v == nil {
		return
	}

	switch t := t.(type) {
	case *Fun:
		Visit(v, t.Ret)
		for _, p := range t.Params {
			Visit(v, p)
		}
	case *Tuple:
		for _, e := range t.Elems {
			Visit(v, e)
		}
	case *Array:
		Visit(v, t.Elem)
	case *Option:
		Visit(v, t.Elem)
	case *Var:
		if t.Ref != nil {
			Visit(v, t.Ref)
		}
	}

	vis.VisitBottomup(t)
}
