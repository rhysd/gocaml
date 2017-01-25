package ast

import (
	"fmt"
	"strings"
)

// I want to move this file to ../typing but it's not possible
// because this file has a cross reference to ast.go

type Type interface {
	String() string
}

type UnitType struct {
}

func (t *UnitType) String() string {
	return "()"
}

type BoolType struct {
}

func (t *BoolType) String() string {
	return "bool"
}

type IntType struct {
}

func (t *IntType) String() string {
	return "int"
}

type FloatType struct {
}

func (t *FloatType) String() string {
	return "float"
}

type FunType struct {
	Ret    Type
	Params []Type
}

func (t *FunType) String() string {
	params := make([]string, len(t.Params))
	for i, p := range t.Params {
		params[i] = p.String()
	}
	return fmt.Sprintf("fun %s -> %s", strings.Join(params, ", "), t.Ret.String())
}

type TupleType struct {
	Elems []Type
}

func (t *TupleType) String() string {
	elems := make([]string, len(t.Elems))
	for i, e := range t.Elems {
		elems[i] = e.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

type ArrayType struct {
	Elem Type
}

func (t *ArrayType) String() string {
	return fmt.Sprintf("%s array", t.Elem.String())
}

type TypeVar struct {
	Id  int
	Ref Type
}

func (t *TypeVar) String() string {
	if t.Ref == nil {
		return fmt.Sprintf("$%d(unknown)", t.Id)
	}
	return fmt.Sprintf("$%d(%s)", t.Id, t.Ref.String())
}

var (
	// Make singleton type values because it doesn't have any contextual information
	UnitTypeVal  = &UnitType{}
	BoolTypeVal  = &BoolType{}
	IntTypeVal   = &IntType{}
	FloatTypeVal = &FloatType{}

	// ID to identify type variables
	typeVarId = 0
)

func NewTypeVar() *TypeVar {
	typeVarId++
	return &TypeVar{
		Id:  typeVarId,
		Ref: nil,
	}
}
