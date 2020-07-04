package types

import (
	"testing"
)

func TestEquals(t *testing.T) {
	gen := NewGeneric()
	free := NewVar(nil, 0)
	cases := []Type{
		IntType,
		FloatType,
		free,
		gen,
		&Array{IntType},
		&Option{free},
		NewVar(&Tuple{[]Type{UnitType, NewVar(free, 0), NewVar(gen, 0)}}, 0),
		&Fun{free, []Type{&Array{gen}, StringType, BoolType}},
	}

	for i, l := range cases {
		if !Equals(l, l) {
			s := Debug(l)
			t.Error("`%s` == `%s` is false", s, s)
		}
		j := i + 1
		if j == len(cases) {
			j = 0
		}
		r := cases[j]
		if Equals(l, r) {
			t.Error("`%s` != `%s` is false", Debug(l), Debug(r))
		}
	}
}
