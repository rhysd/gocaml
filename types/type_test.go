package types

import (
	"strings"
	"testing"
)

func TestTupleString(t *testing.T) {
	tpl := &Tuple{[]Type{IntType, BoolType, &Tuple{[]Type{FloatType, UnitType}}}}
	s := tpl.String()
	if s != "int * bool * (float * unit)" {
		t.Fatal("Tuple string format is unexpected:", s)
	}
	// Tuple in other type
	v := &Var{&Tuple{[]Type{IntType, BoolType}}, 0}
	s = v.String()
	if s != "int * bool" {
		t.Fatal("Tuple string nested in other type is unexpected:", s)
	}
}

func TestFunString(t *testing.T) {
	fun := &Fun{
		&Fun{IntType, []Type{&Option{StringType}}},
		[]Type{
			IntType,
			&Fun{&Array{BoolType}, []Type{FloatType}},
		},
	}
	s := fun.String()
	if s != "int -> (float -> bool array) -> (string option -> int)" {
		t.Fatal("Function string format is unexpected:", s)
	}
	// Fun in other type
	v := &Var{&Fun{IntType, []Type{BoolType}}, 0}
	s = v.String()
	if s != "bool -> int" {
		t.Fatal("Function string nested in other type is unexpected:", s)
	}
}

func TestVarString(t *testing.T) {
	var_ := func(t Type) *Var {
		return &Var{t, 0}
	}
	v := var_(nil)
	s := v.String()
	if s[0] != '?' {
		t.Fatal("Incorrect empty variable format:", s)
	}
	v = var_(var_(&Option{&Array{StringType}}))
	s = v.String()
	if s != "string array option" {
		t.Fatal("Type variable is not stripped correctly:", s)
	}
}

func TestGenGeneric(t *testing.T) {
	g1 := NewGeneric()
	g2 := NewGeneric()
	if g1.Id == g2.Id {
		t.Fatal("NewGeneric should generate Generic with unique ID")
	}
	v := &Var{}
	g3 := NewGenericFromVar(v)
	if g3.Id != v.ID() {
		t.Fatal("Generic created with NewGenericFromVar() should inherit ID of variable")
	}
}

func TestGenericString(t *testing.T) {
	s := NewGeneric().String()
	if s != "'a" {
		t.Fatal("Generic name must start with 'a", s)
	}

	g := NewGeneric()
	s = (&Tuple{[]Type{g, g}}).String()
	if s != "'a * 'a" {
		t.Fatal("The same name should be given to the same variable:", s)
	}

	g2 := NewGeneric()
	s = (&Fun{g2, []Type{g, g2, g}}).String()
	if s != "'a -> 'b -> 'a -> 'b" {
		t.Fatal("Multiple generic variables must be treated correctly:", s)
	}

	ts := make([]Type, 0, 27)
	for i := 0; i < 27; i++ {
		ts = append(ts, NewGeneric())
	}
	s = (&Tuple{ts}).String()
	if !strings.HasSuffix(s, " * 'a1") {
		t.Fatal("Generic name must be rotated with count:", s)
	}
}
