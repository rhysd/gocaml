package gcil

import (
	"fmt"
	"io"
	"strings"
)

type Val interface {
	Print(io.Writer)
}

type OperatorKind int

// Operators
const (
	NOT OperatorKind = iota
	NEG
	FNEG
	ADD
	SUB
	FADD
	FSUB
	FMUL
	FDIV
	LESS
	EQ
)

var opTable = [...]string{
	NOT:  "not",
	NEG:  "-",
	FNEG: "-.",
	ADD:  "+",
	SUB:  "-",
	FADD: "+.",
	FSUB: "-.",
	FMUL: "*.",
	FDIV: "/.",
	LESS: "<",
	EQ:   "=",
}

type (
	Unit struct{}
	Bool struct {
		Const bool
	}
	Int struct {
		Const int
	}
	Float struct {
		Const float64
	}
	Unary struct {
		Op    OperatorKind
		Child string
	}
	Binary struct {
		Op  OperatorKind
		Lhs string
		Rhs string
	}
	Ref struct {
		Ident string
	}
	If struct {
		Cond string
		Then *Block
		Else *Block
	}
	Fun struct {
		Params []string
		Body   *Block
	}
	App struct {
		Callee string
		Args   []string
	}
	Tuple struct {
		Elems []string
	}
	Array struct {
		Size string
		Elem string
	}
	TplLoad struct { // Used for each element of LetTuple
		From  string
		Index int
	}
	ArrLoad struct {
		From  string
		Index string
	}
	ArrStore struct {
		To    string
		Index string
		Rhs   string
	}
	XRef struct {
		Ident string
	}
)

var (
	UnitVal = &Unit{}
)

func (v *Unit) Print(out io.Writer) {
	fmt.Fprintf(out, "unit")
}
func (v *Bool) Print(out io.Writer) {
	fmt.Fprintf(out, "bool %v", v.Const)
}
func (v *Int) Print(out io.Writer) {
	fmt.Fprintf(out, "int %d", v.Const)
}
func (v *Float) Print(out io.Writer) {
	fmt.Fprintf(out, "float %f", v.Const)
}
func (v *Unary) Print(out io.Writer) {
	fmt.Fprintf(out, "unary %s %s", opTable[v.Op], v.Child)
}
func (v *Binary) Print(out io.Writer) {
	fmt.Fprintf(out, "binary %s %s %s", opTable[v.Op], v.Lhs, v.Rhs)
}
func (v *Ref) Print(out io.Writer) {
	fmt.Fprintf(out, "ref %s", v.Ident)
}
func (v *If) Print(out io.Writer) {
	fmt.Fprintf(out, "if %s\n", v.Cond)
	v.Then.Println(out)
	v.Else.Println(out)
}
func (v *Fun) Print(out io.Writer) {
	fmt.Fprintf(out, "fun %s\n", strings.Join(v.Params, ","))
	v.Body.Println(out)
}
func (v *App) Print(out io.Writer) {
	fmt.Fprintf(out, "app %s %s", v.Callee, strings.Join(v.Args, ","))
}
func (v *Tuple) Print(out io.Writer) {
	fmt.Fprintf(out, "tuple %s", strings.Join(v.Elems, ","))
}
func (v *Array) Print(out io.Writer) {
	fmt.Fprintf(out, "array %s %s", v.Size, v.Elem)
}
func (v *TplLoad) Print(out io.Writer) {
	fmt.Fprintf(out, "tplload %d %s", v.Index, v.From)
}
func (v *ArrLoad) Print(out io.Writer) {
	fmt.Fprintf(out, "arrload %s %s", v.Index, v.From)
}
func (v *ArrStore) Print(out io.Writer) {
	fmt.Fprintf(out, "arrstore %s %s %s", v.Index, v.To, v.Rhs)
}
func (v *XRef) Print(out io.Writer) {
	fmt.Fprintf(out, "xref %s", v.Ident)
}
