package types

import (
	"fmt"
	"strings"
)

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

// INT32_MAX. When this value is specified to variable's level, it means that the variable is
// 'forall a.a' (generic bound type variable). It's because any other level is smaller than
// the GenericLevel. Type inference algorithm treats type variables whose level is larger than
// current level as generic type.
const GenericLevel = 2147483647

type VarID uint64
type Var struct {
	Ref   Type
	Level int
	ID    VarID
}

func (t *Var) String() string {
	return newToString().ofVar(t)
}

var currentVarID VarID = 0

func NewVar(t Type, l int) *Var {
	currentVarID++
	return &Var{t, l, currentVarID}
}

func (t *Var) SetGeneric() {
	if t.Ref != nil {
		panic("FATAL: Cannot promote linked type variable to generic variable")
	}
	t.Level = GenericLevel
}

func (t *Var) IsGeneric() bool {
	return t.Level == GenericLevel
}

func NewGeneric() *Var {
	currentVarID++
	return &Var{nil, GenericLevel, currentVarID}
}

// Make singleton type values because it doesn't have any contextual information
var (
	UnitType   = &Unit{}
	BoolType   = &Bool{}
	IntType    = &Int{}
	FloatType  = &Float{}
	StringType = &String{}
)

type toString struct {
	generics map[VarID]string
	count    int
	char     rune
	debug    bool
}

func newToString() *toString {
	return &toString{map[VarID]string{}, 0, 'a', false}
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
	if v.Ref != nil {
		if toStr.debug {
			return fmt.Sprintf("?(%s, %d, %d)", toStr.ofType(v.Ref), v.ID, v.Level)
		}
		return toStr.ofType(v.Ref)
	}
	if v.Level != GenericLevel {
		if toStr.debug {
			return fmt.Sprintf("?(%d, %d)", v.ID, v.Level)
		}
		return fmt.Sprintf("?(%d)", v.ID)
	}
	if s, ok := toStr.generics[v.ID]; ok {
		return s
	}
	s := toStr.newGenName()
	if toStr.debug {
		s = fmt.Sprintf("%s(%d)", s, v.ID)
	}
	toStr.generics[v.ID] = s
	return s
}

// Debug represents the given type as string with detailed type variable information.
func Debug(t Type) string {
	tos := &toString{map[VarID]string{}, 0, 'a', true}
	return tos.ofType(t)
}
