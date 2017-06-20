package types

import (
	"fmt"
	"strings"
	"unsafe"
)

type Type interface {
	String() string
}

type toString struct {
	generics map[GenericId]string
	count    int
	char     rune
}

func newToString() *toString {
	return &toString{map[GenericId]string{}, 0, 'a'}
}

func (toStr *toString) newGenName() string {
	var n string
	if toStr.count == 0 {
		n = "'" + string(toStr.char)
	} else {
		n = fmt.Sprintf("'%c%d", toStr.char, toStr.count)
	}
	if toStr.char == 'z' {
		toStr.char = 'a'
		toStr.count++
	} else {
		toStr.char++
	}
	return n
}

func (toStr *toString) ofType(t Type) string {
	switch t := t.(type) {
	case *Unit, *Bool, *Int, *Float, *String:
		// Monomorphic types
		return t.String()
	case *Fun:
		return toStr.ofFun(t)
	case *Tuple:
		return toStr.ofTuple(t)
	case *Array:
		return toStr.ofArray(t)
	case *Option:
		return toStr.ofOption(t)
	case *Var:
		return toStr.ofVar(t)
	case *Generic:
		return toStr.ofGeneric(t)
	default:
		panic("FATAL: Unreachable: Cannot stringify unknown type")
	}
}

func (toStr *toString) ofNestedType(t Type) string {
	switch t := t.(type) {
	case *Fun:
		return fmt.Sprintf("(%s)", toStr.ofFun(t))
	case *Tuple:
		return fmt.Sprintf("(%s)", toStr.ofTuple(t))
	default:
		return toStr.ofType(t)
	}
}

func (toStr *toString) ofFun(f *Fun) string {
	ss := make([]string, 0, len(f.Params)+1)
	for _, p := range f.Params {
		ss = append(ss, toStr.ofNestedType(p))
	}
	ss = append(ss, toStr.ofNestedType(f.Ret))
	return strings.Join(ss, " -> ")
}

func (toStr *toString) ofTuple(t *Tuple) string {
	elems := make([]string, len(t.Elems))
	for i, e := range t.Elems {
		elems[i] = toStr.ofNestedType(e)
	}
	return strings.Join(elems, " * ")
}

func (toStr *toString) ofArray(a *Array) string {
	return toStr.ofNestedType(a.Elem) + " array"
}

func (toStr *toString) ofOption(o *Option) string {
	return toStr.ofNestedType(o.Elem) + " option"
}

func (toStr *toString) ofVar(v *Var) string {
	if v.Ref == nil {
		return fmt.Sprintf("?(%p)", v)
	}
	return toStr.ofType(v.Ref)
}

func (toStr *toString) ofGeneric(g *Generic) string {
	if s, ok := toStr.generics[g.Id]; ok {
		return s
	}
	s := toStr.newGenName()
	toStr.generics[g.Id] = s
	return s
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

func (t *Fun) String() string {
	return newToString().ofFun(t)
}

type Tuple struct {
	Elems []Type
}

func (t *Tuple) String() string {
	return newToString().ofTuple(t)
}

type Array struct {
	Elem Type
}

func (t *Array) String() string {
	return newToString().ofArray(t)
}

type Option struct {
	Elem Type
}

func (t *Option) String() string {
	return newToString().ofOption(t)
}

type Var struct {
	Ref   Type
	Level int
}

func (t *Var) String() string {
	return newToString().ofVar(t)
}

func (t *Var) ID() GenericId {
	return GenericId(unsafe.Pointer(t))
}

type GenericId uintptr
type Generic struct {
	Id GenericId
}

func (t *Generic) String() string {
	return newToString().ofGeneric(t)
}

func NewGeneric() *Generic {
	g := &Generic{}
	g.Id = GenericId(unsafe.Pointer(g))
	return g
}

func NewGenericFromVar(v *Var) *Generic {
	return &Generic{v.ID()}
}

// Make singleton type values because it doesn't have any contextual information
var (
	UnitType   = &Unit{}
	BoolType   = &Bool{}
	IntType    = &Int{}
	FloatType  = &Float{}
	StringType = &String{}
)
