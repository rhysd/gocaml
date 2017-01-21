package ast

import (
	"github.com/rhysd/gocaml/token"
)

// Type t =
//   | Unit
//   | Bool of bool
//   | Int of int
//   | Float of float
//   | Not of t
//   | Neg of t
//   | Add of t * t
//   | Sub of t * t
//   | FNeg of t
//   | FAdd of t * t
//   | FSub of t * t
//   | FMul of t * t
//   | FDiv of t * t
//   | Eq of t * t
//   | LE of t * t
//   | If of t * t * t
//   | Let of (Id.t * Type.t) * t * t
//   | Var of Id.t
//   | LetRec of fundef * t
//   | App of t * t list
//   | Tuple of t list
//   | LetTuple of (Id.t * Type.t) list * t * t
//   | Array of t * t
//   | Get of t * t
//   | Put of t * t * t
// and fundef = { name : Id.t * Type.t; args : (Id.t * Type.t) list; body : t }

type AST struct {
	Root Expr
	File *token.Source
}

type Expr interface {
	Pos() token.Position
	End() token.Position
	Name() string
	// Type() *typing.Type
}

type Decl struct {
	Name string
	// Type typing.Type
}

type FuncDef struct {
	Decl   Decl
	Params []string
	Body   Expr
}

// AST node which meets Expr interface
type (
	Unit struct {
		LParenToken *token.Token
		RParenToken *token.Token
	}

	Bool struct {
		Token *token.Token
		Value bool
	}

	Int struct {
		Token *token.Token
		Value int
	}

	Float struct {
		Token *token.Token
		Value float
	}

	Not struct {
		Child Expr
	}

	Neg struct {
		Child Expr
	}

	Add struct {
		Left  Expr
		Right Expr
	}

	Sub struct {
		Left  Expr
		Right Expr
	}

	FNeg struct {
		Child Expr
	}

	FAdd struct {
		Left  Expr
		Right Expr
	}

	FSub struct {
		Left  Expr
		Right Expr
	}

	FMul struct {
		Left  Expr
		Right Expr
	}

	FDiv struct {
		Left  Expr
		Right Expr
	}

	Eq struct {
		Left  Expr
		Right Expr
	}

	Less struct {
		Left  Expr
		Right Expr
	}

	If struct {
		IfToken *token.Token
		Cond    Expr
		Then    Expr
		Else    Expr
	}

	Let struct {
		LetToken *token.Token
		end      *token.Token
		Decl     Decl
		Bound    Expr
		Body     Expr
	}

	Var struct {
		Token *token.Token
		Name  string
	}

	LetRec struct {
		LetToken *token.Token
		Func     FuncDef
		Body     Expr
	}

	Apply struct {
		Callee Expr
		Args   []Expr
	}

	Tuple struct {
		Elems []Expr
	}

	LetTuple struct {
		LetToken *token.Token
		Decls    []Decl
		Bound    Expr
		Body     Expr
	}

	Array struct {
		ArrayToken *token.Token
		Size       Expr
		Elem       Expr
	}

	Get struct {
		Array Expr
		Index Expr
	}

	Put struct {
		Array    Expr
		Index    Expr
		Assignee Expr
	}
)

func (e *Unit) Pos() token.Position {
	return e.LParenToken.Start
}
func (e *Unit) End() token.Position {
	return e.RParenToken.End
}

func (e *Bool) Pos() token.Position {
	return e.Token.Start
}
func (e *Bool) End() token.Position {
	return e.Token.End
}

func (e *Int) Pos() token.Position {
	return e.Token.Start
}
func (e *Int) End() token.Position {
	return e.Token.End
}

func (e *Float) Pos() token.Position {
	return e.Token.Start
}
func (e *Float) End() token.Position {
	return e.Token.End
}

func (e *Not) Pos() token.Position {
	return e.Child.Pos()
}
func (e *Not) End() token.Position {
	return e.Child.End()
}

func (e *Neg) Pos() token.Position {
	return e.Child.Pos()
}
func (e *Neg) End() token.Position {
	return e.Child.End()
}

func (e *Add) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Add) End() token.Position {
	return e.Right.Pos()
}

func (e *Sub) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Sub) End() token.Position {
	return e.Right.Pos()
}

func (e *FNeg) Pos() token.Position {
	return e.Left.Pos()
}
func (e *FNeg) End() token.Position {
	return e.Right.Pos()
}

func (e *FAdd) Pos() token.Position {
	return e.Left.Pos()
}
func (e *FAdd) End() token.Position {
	return e.Right.Pos()
}

func (e *FSub) Pos() token.Position {
	return e.Left.Pos()
}
func (e *FSub) End() token.Position {
	return e.Right.Pos()
}

func (e *FMul) Pos() token.Position {
	return e.Left.Pos()
}
func (e *FMul) End() token.Position {
	return e.Right.Pos()
}

func (e *FDiv) Pos() token.Position {
	return e.Left.Pos()
}
func (e *FDiv) End() token.Position {
	return e.Right.Pos()
}

func (e *Eq) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Eq) End() token.Position {
	return e.Right.Pos()
}

func (e *Less) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Less) End() token.Position {
	return e.Right.Pos()
}

func (e *If) Pos() token.Position {
	return e.IfToken.Start
}
func (e *If) End() token.Position {
	return e.Else.End()
}

func (e *Let) Pos() token.Position {
	return e.LetToken.Start
}
func (e *Let) End() token.Position {
	return e.Body.End()
}

func (e *Var) Pos() token.Position {
	return e.Token.Start
}
func (e *Var) End() token.Position {
	return e.Token.End
}

func (e *LetRec) Pos() token.Position {
	return e.LetToken.Start
}
func (e *LetRec) End() token.Position {
	return e.Body.End()
}

func (e *Apply) Pos() token.Position {
	return e.Callee.Pos()
}
func (e *Apply) End() token.Position {
	if len(e.Elems) == 0 {
		return e.Callee.End()
	}
	return e.Elems[len(e.Elems)-1].End()
}

func (e *Tuple) Pos() token.Position {
	return e.Elems[0].Pos()
}
func (e *Tuple) End() token.Position {
	return e.Elems[len(e.Elems)-1].End()
}

func (e *LetTuple) Pos() token.Position {
	return e.LetToken.Start
}
func (e *LetTuple) End() token.Position {
	return e.Body.End()
}

func (e *Array) Pos() token.Position {
	return e.ArrayToken.Start
}
func (e *Array) End() token.Position {
	return e.Elem.End()
}

func (e *Get) Pos() token.Position {
	return e.Array.Pos()
}
func (e *Get) End() token.Position {
	return e.Index.End()
}

func (e *Put) Pos() token.Position {
	return e.Array.Pos()
}
func (e *Put) End() token.Position {
	return e.Assignee.End()
}

func (e *Unit) Name()     { return "Unit" }
func (e *Bool) Name()     { return "Bool" }
func (e *Int) Name()      { return "Int" }
func (e *Float) Name()    { return "Float" }
func (e *Not) Name()      { return "Not" }
func (e *Neg) Name()      { return "Neg" }
func (e *Add) Name()      { return "Add" }
func (e *Sub) Name()      { return "Sub" }
func (e *FNeg) Name()     { return "FNeg" }
func (e *FAdd) Name()     { return "FAdd" }
func (e *FSub) Name()     { return "FSub" }
func (e *FMul) Name()     { return "FMul" }
func (e *FDiv) Name()     { return "FDiv" }
func (e *Eq) Name()       { return "Eq" }
func (e *Less) Name()     { return "Less" }
func (e *If) Name()       { return "If" }
func (e *Let) Name()      { return "Let" }
func (e *Var) Name()      { return "Var" }
func (e *LetRec) Name()   { return "LetRec" }
func (e *Apply) Name()    { return "Apply" }
func (e *Tuple) Name()    { return "Tuple" }
func (e *LetTuple) Name() { return "LetTuple" }
func (e *Array) Name()    { return "Array" }
func (e *Get) Name()      { return "Get" }
func (e *Put) Name()      { return "Put" }
