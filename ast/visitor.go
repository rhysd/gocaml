package ast

type Visitor interface {
	Visit(e Expr) Visitor
}

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
	case *Less:
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
	case *Array:
		Visit(v, n.Size)
		Visit(v, n.Elem)
	case *Get:
		Visit(v, n.Array)
		Visit(v, n.Index)
	case *Put:
		Visit(v, n.Array)
		Visit(v, n.Index)
		Visit(v, n.Assignee)
	}
}

type predicate func(Expr) bool

func (p predicate) Visit(e Expr) Visitor {
	if !p(e) {
		return nil
	}
	return p
}

func Find(e Expr, f func(Expr) bool) {
	Visit(predicate(f), e)
}
