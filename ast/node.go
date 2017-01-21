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
		start *token.Token
		end   *token.Token
	}

	Bool struct {
		start *token.Token
		end   *token.Token
		Value bool
	}

	Int struct {
		start *token.Token
		end   *token.Token
		Value int
	}

	Float struct {
		start *token.Token
		end   *token.Token
		Value float
	}

	Not struct {
		start *token.Token
		end   *token.Token
		Child Expr
	}

	Neg struct {
		start *token.Token
		end   *token.Token
		Child Expr
	}

	Add struct {
		start *token.Token
		end   *token.Token
		Left  Expr
		Right Expr
	}

	Sub struct {
		start *token.Token
		end   *token.Token
		Left  Expr
		Right Expr
	}

	FNeg struct {
		start *token.Token
		end   *token.Token
		Child Expr
	}

	FAdd struct {
		start *token.Token
		end   *token.Token
		Left  Expr
		Right Expr
	}

	FSub struct {
		start *token.Token
		end   *token.Token
		Left  Expr
		Right Expr
	}

	FMul struct {
		start *token.Token
		end   *token.Token
		Left  Expr
		Right Expr
	}

	FDiv struct {
		start *token.Token
		end   *token.Token
		Left  Expr
		Right Expr
	}

	Eq struct {
		start *token.Token
		end   *token.Token
		Left  Expr
		Right Expr
	}

	Less struct {
		start *token.Token
		end   *token.Token
		Left  Expr
		Right Expr
	}

	If struct {
		start *token.Token
		end   *token.Token
		Cond  Expr
		Then  Expr
		Else  Expr
	}

	Let struct {
		start *token.Token
		end   *token.Token
		Decl  Decl
		Bound Expr
		Body  Expr
	}

	Var struct {
		start *token.Token
		end   *token.Token
		Name  string
	}

	LetRec struct {
		start *token.Token
		end   *token.Token
		Func  FuncDef
		Body  Expr
	}

	Apply struct {
		start  *token.Token
		end    *token.Token
		Callee Expr
		Args   []Expr
	}

	Tuple struct {
		start *token.Token
		end   *token.Token
		Elems []Expr
	}

	LetTuple struct {
		start *token.Token
		end   *token.Token
		Decls []Decl
		Bound Expr
		Body  Expr
	}

	Array struct {
		start *token.Token
		end   *token.Token
		Size  Expr
		Elem  Expr
	}

	Get struct {
		start *token.Token
		end   *token.Token
		Array Expr
		Index Expr
	}

	Put struct {
		start    *token.Token
		end      *token.Token
		Array    Expr
		Index    Expr
		Assignee Expr
	}
)

func (e *Unit) Pos()     { return e.start.Start }
func (e *Unit) End()     { return e.end.End }
func (e *Bool) Pos()     { return e.start.Start }
func (e *Bool) End()     { return e.end.End }
func (e *Int) Pos()      { return e.start.Start }
func (e *Int) End()      { return e.end.End }
func (e *Float) Pos()    { return e.start.Start }
func (e *Float) End()    { return e.end.End }
func (e *Not) Pos()      { return e.start.Start }
func (e *Not) End()      { return e.end.End }
func (e *Neg) Pos()      { return e.start.Start }
func (e *Neg) End()      { return e.end.End }
func (e *Add) Pos()      { return e.start.Start }
func (e *Add) End()      { return e.end.End }
func (e *Sub) Pos()      { return e.start.Start }
func (e *Sub) End()      { return e.end.End }
func (e *FNeg) Pos()     { return e.start.Start }
func (e *FNeg) End()     { return e.end.End }
func (e *FAdd) Pos()     { return e.start.Start }
func (e *FAdd) End()     { return e.end.End }
func (e *FSub) Pos()     { return e.start.Start }
func (e *FSub) End()     { return e.end.End }
func (e *FMul) Pos()     { return e.start.Start }
func (e *FMul) End()     { return e.end.End }
func (e *FDiv) Pos()     { return e.start.Start }
func (e *FDiv) End()     { return e.end.End }
func (e *Eq) Pos()       { return e.start.Start }
func (e *Eq) End()       { return e.end.End }
func (e *Less) Pos()     { return e.start.Start }
func (e *Less) End()     { return e.end.End }
func (e *If) Pos()       { return e.start.Start }
func (e *If) End()       { return e.end.End }
func (e *Let) Pos()      { return e.start.Start }
func (e *Let) End()      { return e.end.End }
func (e *Var) Pos()      { return e.start.Start }
func (e *Var) End()      { return e.end.End }
func (e *LetRec) Pos()   { return e.start.Start }
func (e *LetRec) End()   { return e.end.End }
func (e *Apply) Pos()    { return e.start.Start }
func (e *Apply) End()    { return e.end.End }
func (e *Tuple) Pos()    { return e.start.Start }
func (e *Tuple) End()    { return e.end.End }
func (e *LetTuple) Pos() { return e.start.Start }
func (e *LetTuple) End() { return e.end.End }
func (e *Array) Pos()    { return e.start.Start }
func (e *Array) End()    { return e.end.End }
func (e *Get) Pos()      { return e.start.Start }
func (e *Get) End()      { return e.end.End }
func (e *Put) Pos()      { return e.start.Start }
func (e *Put) End()      { return e.end.End }

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
