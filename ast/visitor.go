package ast

// Visitor is an interface for the structs which is used for traversing AST.
type Visitor interface {
	// VisitTopdown defines the process when a node is visited. This method is called before
	// children are visited.
	// Returned value is a next visitor to use for succeeding visit. When wanting to stop
	// visiting, please return nil.
	// A visitor visits in depth-first order.
	VisitTopdown(e Expr) Visitor
	// VisitBottomup defines the process when a node is visited. This method is called after
	// children were visited. When VisitTopdown returned nil, this method won't be caled for the node.
	VisitBottomup(e Expr)
}

// Visit visits the tree with the visitor.
func Visit(vis Visitor, e Expr) {
	v := vis.VisitTopdown(e)
	if v == nil {
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
	case *Mod:
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
		if n.Type != nil {
			Visit(v, n.Type)
		}
		Visit(v, n.Bound)
		Visit(v, n.Body)
	case *LetRec:
		for _, p := range n.Func.Params {
			if p.Type != nil {
				Visit(v, p.Type)
			}
		}
		if n.Func.RetType != nil {
			Visit(v, n.Func.RetType)
		}
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
		if n.Type != nil {
			Visit(v, n.Type)
		}
		Visit(v, n.Bound)
		Visit(v, n.Body)
	case *ArrayMake:
		Visit(v, n.Size)
		Visit(v, n.Elem)
	case *ArraySize:
		Visit(v, n.Target)
	case *ArrayGet:
		Visit(v, n.Array)
		Visit(v, n.Index)
	case *ArrayPut:
		Visit(v, n.Array)
		Visit(v, n.Index)
		Visit(v, n.Assignee)
	case *Match:
		Visit(v, n.Target)
		Visit(v, n.IfSome)
		Visit(v, n.IfNone)
	case *Some:
		Visit(v, n.Child)
	case *ArrayLit:
		for _, e := range n.Elems {
			Visit(v, e)
		}
	case *FuncType:
		for _, e := range n.ParamTypes {
			Visit(v, e)
		}
		Visit(v, n.RetType)
	case *TupleType:
		for _, e := range n.ElemTypes {
			Visit(v, e)
		}
	case *CtorType:
		for _, e := range n.ParamTypes {
			Visit(v, e)
		}
	case *Typed:
		Visit(v, n.Child)
		Visit(v, n.Type)
	case *TypeDecl:
		Visit(v, n.Type)
	case *External:
		Visit(v, n.Type)
	}

	vis.VisitBottomup(e)
}
