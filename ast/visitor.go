package ast

// Visitor is an interface for the structs which is used for traversing AST.
type Visitor interface {
	// Visit defines the process when a node is visit.
	// Visitor is a next visitor to use for visit.
	// When wanting to stop visiting, return nil.
	Visit(e Expr) Visitor
}

// Visit visits the tree with the visitor.
func Visit(v Visitor, e Expr) {
	if v = v.Visit(e); v == nil {
		return
	}

	switch n := e.(type) {
	case *Not:
		Visit(v, n.Child)
	case *Neg:
		Visit(v, n.Child)
	case *Add:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *Sub:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *Mul:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *Div:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *FNeg:
		Visit(v, n.Child)
	case *FAdd:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *FSub:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *FMul:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *FDiv:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *Eq:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *NotEq:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *Less:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *LessEq:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *Greater:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *GreaterEq:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *And:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *Or:
		Visit(v, n.Left)
		Visit(v, n.Right)
	case *If:
		Visit(v, n.Cond)
		Visit(v, n.Then)
		Visit(v, n.Else)
	case *Let:
		Visit(v, n.Bound)
		Visit(v, n.Body)
	case *LetRec:
		Visit(v, n.Func.Body)
		Visit(v, n.Body)
	case *Apply:
		Visit(v, n.Callee)
		for _, e := range n.Args {
			Visit(v, e)
		}
	case *Tuple:
		for _, e := range n.Elems {
			Visit(v, e)
		}
	case *LetTuple:
		Visit(v, n.Bound)
		Visit(v, n.Body)
	case *ArrayCreate:
		Visit(v, n.Size)
		Visit(v, n.Elem)
	case *ArraySize:
		Visit(v, n.Target)
	case *Get:
		Visit(v, n.Array)
		Visit(v, n.Index)
	case *Put:
		Visit(v, n.Array)
		Visit(v, n.Index)
		Visit(v, n.Assignee)
	}
}

type finder struct {
	found     bool
	predicate func(Expr) bool
}

func (f *finder) Visit(e Expr) Visitor {
	if f.found {
		return nil
	}
	if f.predicate(e) {
		f.found = true
		return nil
	}
	return f
}

func Find(e Expr, p func(Expr) bool) bool {
	f := &finder{
		found:     false,
		predicate: p,
	}
	Visit(f, e)
	return f.found
}

type childrenVisitor struct {
	isChild   bool
	predicate func(Expr)
}

func (v *childrenVisitor) Visit(e Expr) Visitor {
	if v.isChild {
		v.predicate(e)
		return nil
	}
	v.isChild = true
	return v // Visit children
}

func VisitChildren(e Expr, pred func(e Expr)) {
	v := &childrenVisitor{
		isChild:   false,
		predicate: pred,
	}
	Visit(v, e)
}
