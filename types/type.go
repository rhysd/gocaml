package types

import (
	"fmt"
	"strings"
)

// I want to move this file to ../typing but it's not possible
// because this file has a cross reference to ast.go

type Type interface {
	String() string
}

type Unit struct {
}

func (t *Unit) String() string {
	return "unit"
}

type Bool struct {
}

func (t *Bool) String() string {
	return "bool"
}

type Int struct {
}

func (t *Int) String() string {
	return "int"
}

type Float struct {
}

func (t *Float) String() string {
	return "float"
}

type String struct {
}

func (t *String) String() string {
	return "string"
}

type Fun struct {
	Ret    Type
	Params []Type
}

func funElemTy(t Type) string {
	if f, ok := t.(*Fun); ok {
		return fmt.Sprintf("(%s)", f.String())
	} else {
		return t.String()
	}
}

func (t *Fun) String() string {
	ss := make([]string, 0, len(t.Params)+1)
	for _, p := range t.Params {
		ss = append(ss, funElemTy(p))
	}
	ss = append(ss, funElemTy(t.Ret))
	return strings.Join(ss, " -> ")
}

type Tuple struct {
	Elems []Type
}

func (t *Tuple) String() string {
	elems := make([]string, len(t.Elems))
	for i, e := range t.Elems {
		if tpl, ok := e.(*Tuple); ok {
			elems[i] = fmt.Sprintf("(%s)", tpl.String())
		} else {
			elems[i] = e.String()
		}
	}
	return strings.Join(elems, " * ")
}

type Array struct {
	Elem Type
}

func (t *Array) String() string {
	return fmt.Sprintf("%s array", t.Elem.String())
}

type Option struct {
	Elem Type
}

func (t *Option) String() string {
	return fmt.Sprintf("%s option", t.Elem.String())
}

type Var struct {
	Ref Type
}

func (t *Var) String() string {
	if t.Ref == nil {
		return fmt.Sprintf("?(%p)", t)
	}
	return t.Ref.String()
}

var (
	// Make singleton type values because it doesn't have any contextual information
	UnitType   = &Unit{}
	BoolType   = &Bool{}
	IntType    = &Int{}
	FloatType  = &Float{}
	StringType = &String{}
)
