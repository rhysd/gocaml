package types

import (
	"testing"
)

func TestTupleString(t *testing.T) {
	tpl := &Tuple{[]Type{IntType, BoolType, &Tuple{[]Type{FloatType, UnitType}}}}
	s := tpl.String()
	if s != "int * bool * (float * unit)" {
		t.Fatal("Tuple string format is unexpected:", s)
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
}

func TestVarString(t *testing.T) {
	v := &Var{}
	s := v.String()
	if s[0] != '?' {
		t.Fatal("Incorrect empty variable format:", s)
	}
	v = &Var{&Var{&Option{&Array{StringType}}}}
	s = v.String()
	if s != "string array option" {
		t.Fatal("Type variable is not stripped correctly:", s)
	}
}
