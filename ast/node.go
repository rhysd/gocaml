// Package ast provides AST definition for GoCaml.
package ast

import (
	"fmt"
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/loc"
	"strings"
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
	Root      Expr
	File      *loc.Source
	TypeDecls []*TypeDecl
}

// Expr is an interface for node of GoCaml AST.
// All nodes have its position and name.
type Expr interface {
	Pos() loc.Pos
	End() loc.Pos
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

var unusedSymCount = 0

func IgnoredSymbol() *Symbol {
	unusedSymCount++
	s := fmt.Sprintf("$unused%d", unusedSymCount)
	return &Symbol{"_", s}
}

func (s *Symbol) IsIgnored() bool {
	return strings.HasPrefix(s.Name, "$unused")
}

type Param struct {
	Ident *Symbol
	Type  Expr
}

type FuncDef struct {
	Symbol  *Symbol
	Params  []Param
	Body    Expr
	RetType Expr
}

func (d *FuncDef) ParamSymbols() []*Symbol {
	syms := make([]*Symbol, 0, len(d.Params))
	for _, p := range d.Params {
		syms = append(syms, p.Ident)
	}
	return syms
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
		Type        Expr // Maybe nil
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
		Type        Expr // Maybe nil
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
		EndPos         loc.Pos
	}

	Some struct {
		StartToken *token.Token
		Child      Expr
	}

	None struct {
		Token *token.Token
	}

	FuncType struct {
		ParamTypes []Expr
		RetType    Expr
	}

	TupleType struct {
		ElemTypes []Expr
	}

	// Note: `int` has no param
	CtorType struct {
		StartToken *token.Token // Maybe nil
		EndToken   *token.Token
		ParamTypes []Expr
		Ctor       string // TODO: It should be identifier after types namespace added
	}

	Typed struct {
		Child Expr
		Type  Expr
	}

	TypeDecl struct {
		Token *token.Token
		Ident string
		Type  Expr
	}
)

func (e *Unit) Pos() loc.Pos {
	return e.LParenToken.Start
}
func (e *Unit) End() loc.Pos {
	return e.RParenToken.End
}

func (e *Bool) Pos() loc.Pos {
	return e.Token.Start
}
func (e *Bool) End() loc.Pos {
	return e.Token.End
}

func (e *Int) Pos() loc.Pos {
	return e.Token.Start
}
func (e *Int) End() loc.Pos {
	return e.Token.End
}

func (e *Float) Pos() loc.Pos {
	return e.Token.Start
}
func (e *Float) End() loc.Pos {
	return e.Token.End
}

func (e *String) Pos() loc.Pos {
	return e.Token.Start
}
func (e *String) End() loc.Pos {
	return e.Token.End
}

func (e *Not) Pos() loc.Pos {
	return e.OpToken.Start
}
func (e *Not) End() loc.Pos {
	return e.Child.End()
}

func (e *Neg) Pos() loc.Pos {
	return e.MinusToken.Start
}
func (e *Neg) End() loc.Pos {
	return e.Child.End()
}

func (e *Add) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *Add) End() loc.Pos {
	return e.Right.End()
}

func (e *Sub) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *Sub) End() loc.Pos {
	return e.Right.End()
}

func (e *Mul) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *Mul) End() loc.Pos {
	return e.Right.End()
}

func (e *Div) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *Div) End() loc.Pos {
	return e.Right.End()
}

func (e *Mod) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *Mod) End() loc.Pos {
	return e.Right.End()
}

func (e *FNeg) Pos() loc.Pos {
	return e.MinusToken.Start
}
func (e *FNeg) End() loc.Pos {
	return e.Child.End()
}

func (e *FAdd) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *FAdd) End() loc.Pos {
	return e.Right.End()
}

func (e *FSub) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *FSub) End() loc.Pos {
	return e.Right.End()
}

func (e *FMul) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *FMul) End() loc.Pos {
	return e.Right.End()
}

func (e *FDiv) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *FDiv) End() loc.Pos {
	return e.Right.End()
}

func (e *Eq) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *Eq) End() loc.Pos {
	return e.Right.End()
}

func (e *NotEq) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *NotEq) End() loc.Pos {
	return e.Right.End()
}

func (e *Less) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *Less) End() loc.Pos {
	return e.Right.End()
}

func (e *LessEq) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *LessEq) End() loc.Pos {
	return e.Right.End()
}

func (e *Greater) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *Greater) End() loc.Pos {
	return e.Right.End()
}

func (e *GreaterEq) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *GreaterEq) End() loc.Pos {
	return e.Right.End()
}

func (e *And) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *And) End() loc.Pos {
	return e.Right.End()
}

func (e *Or) Pos() loc.Pos {
	return e.Left.Pos()
}
func (e *Or) End() loc.Pos {
	return e.Right.End()
}

func (e *If) Pos() loc.Pos {
	return e.IfToken.Start
}
func (e *If) End() loc.Pos {
	return e.Else.End()
}

func (e *Let) Pos() loc.Pos {
	return e.LetToken.Start
}
func (e *Let) End() loc.Pos {
	return e.Body.End()
}

func (e *VarRef) Pos() loc.Pos {
	return e.Token.Start
}
func (e *VarRef) End() loc.Pos {
	return e.Token.End
}

func (e *LetRec) Pos() loc.Pos {
	return e.LetToken.Start
}
func (e *LetRec) End() loc.Pos {
	return e.Body.End()
}

func (e *Apply) Pos() loc.Pos {
	return e.Callee.Pos()
}
func (e *Apply) End() loc.Pos {
	if len(e.Args) == 0 {
		return e.Callee.End()
	}
	return e.Args[len(e.Args)-1].End()
}

func (e *Tuple) Pos() loc.Pos {
	return e.Elems[0].Pos()
}
func (e *Tuple) End() loc.Pos {
	return e.Elems[len(e.Elems)-1].End()
}

func (e *LetTuple) Pos() loc.Pos {
	return e.LetToken.Start
}
func (e *LetTuple) End() loc.Pos {
	return e.Body.End()
}

func (e *ArrayCreate) Pos() loc.Pos {
	return e.ArrayToken.Start
}
func (e *ArrayCreate) End() loc.Pos {
	return e.Elem.End()
}

func (e *ArraySize) Pos() loc.Pos {
	return e.ArrayToken.Start
}
func (e *ArraySize) End() loc.Pos {
	return e.Target.End()
}

func (e *Get) Pos() loc.Pos {
	return e.Array.Pos()
}
func (e *Get) End() loc.Pos {
	return e.Index.End()
}

func (e *Put) Pos() loc.Pos {
	return e.Array.Pos()
}
func (e *Put) End() loc.Pos {
	return e.Assignee.End()
}

func (e *Match) Pos() loc.Pos {
	return e.StartToken.Start
}
func (e *Match) End() loc.Pos {
	return e.EndPos
}

func (e *Some) Pos() loc.Pos {
	return e.StartToken.Start
}
func (e *Some) End() loc.Pos {
	return e.Child.End()
}

func (e *None) Pos() loc.Pos {
	return e.Token.Start
}
func (e *None) End() loc.Pos {
	return e.Token.End
}

func (e *FuncType) Pos() loc.Pos {
	return e.ParamTypes[0].Pos()
}
func (e *FuncType) End() loc.Pos {
	return e.RetType.End()
}

func (e *TupleType) Pos() loc.Pos {
	return e.ElemTypes[0].Pos()
}
func (e *TupleType) End() loc.Pos {
	return e.ElemTypes[len(e.ElemTypes)-1].End()
}

func (e *CtorType) Pos() loc.Pos {
	switch len(e.ParamTypes) {
	case 0:
		// foo
		return e.EndToken.Start
	case 1:
		// a foo
		return e.ParamTypes[0].Pos()
	default:
		// (a, b) foo
		return e.StartToken.Start
	}
}
func (e *CtorType) End() loc.Pos {
	return e.EndToken.End
}

func (e *Typed) Pos() loc.Pos {
	return e.Child.Pos()
}
func (e *Typed) End() loc.Pos {
	return e.Type.End()
}

func (e *TypeDecl) Pos() loc.Pos {
	return e.Token.Start
}
func (e *TypeDecl) End() loc.Pos {
	return e.Type.End()
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
	params := e.Func.Params[0].Ident.DisplayName
	for _, p := range e.Func.Params[1:] {
		params = fmt.Sprintf("%s, %s", params, p.Ident.DisplayName)
	}
	return fmt.Sprintf("LetRec (fun %s %s)", e.Func.Symbol.DisplayName, params)
}
func (e *Apply) Name() string { return "Apply" }
func (e *Tuple) Name() string { return "Tuple" }
func (e *LetTuple) Name() string {
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
func (e *FuncType) Name() string    { return "FuncType" }
func (e *TupleType) Name() string   { return fmt.Sprintf("TupleType (%d)", len(e.ElemTypes)) }
func (e *CtorType) Name() string {
	len := len(e.ParamTypes)
	if len == 0 {
		return fmt.Sprintf("CtorType (%s)", e.Ctor)
	} else {
		return fmt.Sprintf("CtorType (%s (%d))", e.Ctor, len)
	}
}
func (e *Typed) Name() string    { return "Typed" }
func (e *TypeDecl) Name() string { return fmt.Sprintf("TypeDecl (%s)", e.Ident) }
