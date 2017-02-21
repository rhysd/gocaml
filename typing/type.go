package typing

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
	return "()"
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

type Fun struct {
	Ret    Type
	Params []Type
}

func (t *Fun) String() string {
	params := make([]string, len(t.Params))
	for i, p := range t.Params {
		params[i] = p.String()
	}
	return fmt.Sprintf("%s -> %s", strings.Join(params, " -> "), t.Ret.String())
}

type Tuple struct {
	Elems []Type
}

func (t *Tuple) String() string {
	elems := make([]string, len(t.Elems))
	for i, e := range t.Elems {
		elems[i] = e.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

type Array struct {
	Elem Type
}

func (t *Array) String() string {
	return fmt.Sprintf("%s array", t.Elem.String())
}

type Var struct {
	Ref Type
}

func (t *Var) String() string {
	if t.Ref == nil {
		return "(unknown)"
	}
	return t.Ref.String()
}

var (
	// Make singleton type values because it doesn't have any contextual information
	UnitType  = &Unit{}
	BoolType  = &Bool{}
	IntType   = &Int{}
	FloatType = &Float{}
)
