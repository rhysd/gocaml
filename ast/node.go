// Package ast provides AST definition for GoCaml.
package ast

import (
	"fmt"
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

// Expr is an interface for node of GoCaml AST.
// All nodes have its position and name.
type Expr interface {
	Pos() token.Position
	End() token.Position
	Name() string
}

// Note:
// This struct cannot be replaced with string because there may be the
// same name symbol.
type Symbol struct {
	DisplayName string
	Name        string
	// Other symbol attributes go here
}

func NewSymbol(name string) *Symbol {
	return &Symbol{name, name}
}

type FuncDef struct {
	Symbol *Symbol
	Params []*Symbol
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
		Value int64
	}

	Float struct {
		Token *token.Token
		Value float64
	}

	String struct {
		Token *token.Token
		Value string
	}

	Not struct {
		OpToken *token.Token
		Child   Expr
	}

	Neg struct {
		MinusToken *token.Token
		Child      Expr
	}

	Add struct {
		Left, Right Expr
	}

	Sub struct {
		Left, Right Expr
	}

	Mul struct {
		Left, Right Expr
	}

	Div struct {
		Left, Right Expr
	}

	Mod struct {
		Left, Right Expr
	}

	FNeg struct {
		MinusToken *token.Token
		Child      Expr
	}

	FAdd struct {
		Left, Right Expr
	}

	FSub struct {
		Left, Right Expr
	}

	FMul struct {
		Left, Right Expr
	}

	FDiv struct {
		Left, Right Expr
	}

	Eq struct {
		Left, Right Expr
	}

	NotEq struct {
		Left, Right Expr
	}

	Less struct {
		Left, Right Expr
	}

	LessEq struct {
		Left, Right Expr
	}

	Greater struct {
		Left, Right Expr
	}

	GreaterEq struct {
		Left, Right Expr
	}

	And struct {
		Left, Right Expr
	}

	Or struct {
		Left, Right Expr
	}

	If struct {
		IfToken          *token.Token
		Cond, Then, Else Expr
	}

	Let struct {
		LetToken    *token.Token
		Symbol      *Symbol
		Bound, Body Expr
	}

	VarRef struct {
		Token  *token.Token
		Symbol *Symbol
	}

	LetRec struct {
		LetToken *token.Token
		Func     *FuncDef
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
		LetToken    *token.Token
		Symbols     []*Symbol
		Bound, Body Expr
	}

	ArrayCreate struct {
		ArrayToken *token.Token
		Size, Elem Expr
	}

	ArraySize struct {
		ArrayToken *token.Token
		Target     Expr
	}

	Get struct {
		Array, Index Expr
	}

	Put struct {
		Array, Index, Assignee Expr
	}

	Match struct {
		StartToken     *token.Token
		Target         Expr
		IfSome, IfNone Expr
		SomeIdent      *Symbol
		EndPos         token.Position
	}

	Some struct {
		StartToken *token.Token
		Child      Expr
	}

	None struct {
		Token *token.Token
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

func (e *String) Pos() token.Position {
	return e.Token.Start
}
func (e *String) End() token.Position {
	return e.Token.End
}

func (e *Not) Pos() token.Position {
	return e.OpToken.Start
}
func (e *Not) End() token.Position {
	return e.Child.End()
}

func (e *Neg) Pos() token.Position {
	return e.MinusToken.Start
}
func (e *Neg) End() token.Position {
	return e.Child.End()
}

func (e *Add) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Add) End() token.Position {
	return e.Right.End()
}

func (e *Sub) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Sub) End() token.Position {
	return e.Right.End()
}

func (e *Mul) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Mul) End() token.Position {
	return e.Right.End()
}

func (e *Div) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Div) End() token.Position {
	return e.Right.End()
}

func (e *Mod) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Mod) End() token.Position {
	return e.Right.End()
}

func (e *FNeg) Pos() token.Position {
	return e.MinusToken.Start
}
func (e *FNeg) End() token.Position {
	return e.Child.End()
}

func (e *FAdd) Pos() token.Position {
	return e.Left.Pos()
}
func (e *FAdd) End() token.Position {
	return e.Right.End()
}

func (e *FSub) Pos() token.Position {
	return e.Left.Pos()
}
func (e *FSub) End() token.Position {
	return e.Right.End()
}

func (e *FMul) Pos() token.Position {
	return e.Left.Pos()
}
func (e *FMul) End() token.Position {
	return e.Right.End()
}

func (e *FDiv) Pos() token.Position {
	return e.Left.Pos()
}
func (e *FDiv) End() token.Position {
	return e.Right.End()
}

func (e *Eq) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Eq) End() token.Position {
	return e.Right.End()
}

func (e *NotEq) Pos() token.Position {
	return e.Left.Pos()
}
func (e *NotEq) End() token.Position {
	return e.Right.End()
}

func (e *Less) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Less) End() token.Position {
	return e.Right.End()
}

func (e *LessEq) Pos() token.Position {
	return e.Left.Pos()
}
func (e *LessEq) End() token.Position {
	return e.Right.End()
}

func (e *Greater) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Greater) End() token.Position {
	return e.Right.End()
}

func (e *GreaterEq) Pos() token.Position {
	return e.Left.Pos()
}
func (e *GreaterEq) End() token.Position {
	return e.Right.End()
}

func (e *And) Pos() token.Position {
	return e.Left.Pos()
}
func (e *And) End() token.Position {
	return e.Right.End()
}

func (e *Or) Pos() token.Position {
	return e.Left.Pos()
}
func (e *Or) End() token.Position {
	return e.Right.End()
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

func (e *VarRef) Pos() token.Position {
	return e.Token.Start
}
func (e *VarRef) End() token.Position {
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
	if len(e.Args) == 0 {
		return e.Callee.End()
	}
	return e.Args[len(e.Args)-1].End()
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

func (e *ArrayCreate) Pos() token.Position {
	return e.ArrayToken.Start
}
func (e *ArrayCreate) End() token.Position {
	return e.Elem.End()
}

func (e *ArraySize) Pos() token.Position {
	return e.ArrayToken.Start
}
func (e *ArraySize) End() token.Position {
	return e.Target.End()
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

func (e *Match) Pos() token.Position {
	return e.StartToken.Start
}
func (e *Match) End() token.Position {
	return e.EndPos
}

func (e *Some) Pos() token.Position {
	return e.StartToken.Start
}
func (e *Some) End() token.Position {
	return e.Child.End()
}

func (e *None) Pos() token.Position {
	return e.Token.Start
}
func (e *None) End() token.Position {
	return e.Token.End
}

func (e *Unit) Name() string      { return "Unit" }
func (e *Bool) Name() string      { return "Bool" }
func (e *Int) Name() string       { return "Int" }
func (e *Float) Name() string     { return "Float" }
func (e *String) Name() string    { return fmt.Sprintf("String (%s)", e.Token.Value()) }
func (e *Not) Name() string       { return "Not" }
func (e *Neg) Name() string       { return "Neg" }
func (e *Add) Name() string       { return "Add" }
func (e *Sub) Name() string       { return "Sub" }
func (e *Mul) Name() string       { return "Mul" }
func (e *Div) Name() string       { return "Div" }
func (e *Mod) Name() string       { return "Mod" }
func (e *FNeg) Name() string      { return "FNeg" }
func (e *FAdd) Name() string      { return "FAdd" }
func (e *FSub) Name() string      { return "FSub" }
func (e *FMul) Name() string      { return "FMul" }
func (e *FDiv) Name() string      { return "FDiv" }
func (e *Eq) Name() string        { return "Eq" }
func (e *NotEq) Name() string     { return "NotEq" }
func (e *Less) Name() string      { return "Less" }
func (e *LessEq) Name() string    { return "LessEq" }
func (e *Greater) Name() string   { return "Greater" }
func (e *GreaterEq) Name() string { return "GreaterEq" }
func (e *And) Name() string       { return "And" }
func (e *Or) Name() string        { return "Or" }
func (e *If) Name() string        { return "If" }
func (e *Let) Name() string       { return fmt.Sprintf("Let (%s)", e.Symbol.DisplayName) }
func (e *VarRef) Name() string    { return fmt.Sprintf("VarRef (%s)", e.Symbol.DisplayName) }
func (e *LetRec) Name() string {
	if len(e.Func.Params) == 0 {
		panic("LetTuple's symbols field must not be empty")
	}
	params := e.Func.Params[0].DisplayName
	for _, s := range e.Func.Params[1:] {
		params = fmt.Sprintf("%s, %s", params, s.DisplayName)
	}
	return fmt.Sprintf("LetRec (fun %s %s)", e.Func.Symbol.DisplayName, params)
}
func (e *Apply) Name() string { return "Apply" }
func (e *Tuple) Name() string { return "Tuple" }
func (e *LetTuple) Name() string {
	if len(e.Symbols) == 0 {
		panic("LetTuple's symbols field must not be empty")
	}
	vars := e.Symbols[0].DisplayName
	for _, s := range e.Symbols[1:] {
		vars = fmt.Sprintf("%s, %s", vars, s.DisplayName)
	}
	return fmt.Sprintf("LetTuple (%s)", vars)
}
func (e *ArrayCreate) Name() string { return "ArrayCreate" }
func (e *ArraySize) Name() string   { return "ArraySize" }
func (e *Get) Name() string         { return "Get" }
func (e *Put) Name() string         { return "Put" }
func (e *Match) Name() string       { return fmt.Sprintf("Match (%s)", e.SomeIdent.DisplayName) }
func (e *Some) Name() string        { return "Some" }
func (e *None) Name() string        { return "None" }
