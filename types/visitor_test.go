package types

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type testVisitPath struct {
	w io.Writer
}

func (v *testVisitPath) VisitTopdown(t Type) Visitor {
	fmt.Fprintf(v.w, " -> %s (top)", t.String())
	return v
}

func (v *testVisitPath) VisitBottomup(t Type) {
	fmt.Fprintf(v.w, " -> %s (bottom)", t.String())
}

func TestVisitTypes(t *testing.T) {
	cases := []struct {
		input  Type
		output string
	}{
		{
			IntType,
			"int (top) -> int (bottom)",
		},
		{
			&Array{IntType},
			"int array (top) -> int (top) -> int (bottom) -> int array (bottom)",
		},
		{
			&Tuple{[]Type{UnitType, BoolType}},
			"unit * bool (top) -> unit (top) -> unit (bottom) -> bool (top) -> bool (bottom) -> unit * bool (bottom)",
		},
		{
			&Option{StringType},
			"string option (top) -> string (top) -> string (bottom) -> string option (bottom)",
		},
		{
			NewVar(IntType, 0),
			"int (top) -> int (top) -> int (bottom) -> int (bottom)",
		},
	}

	for _, tc := range cases {
		var buf bytes.Buffer
		v := &testVisitPath{&buf}
		Visit(v, tc.input)
		have := buf.String()
		want := " -> " + tc.output
		if have != want {
			t.Errorf("Unexpected visiting path for type '%s': <root> %s", tc.input.String(), have)
		}
	}
}

type testVisitShallow struct {
	last Type
}

func (v *testVisitShallow) VisitTopdown(t Type) Visitor {
	v.last = t
	return nil
}

func (v *testVisitShallow) VisitBottomup(t Type) {
	panic(t.String())
}

func TestVisitStop(t *testing.T) {
	v := &testVisitShallow{}
	ty := &Tuple{[]Type{UnitType, &Array{NewVar(IntType, 0)}}}
	Visit(v, ty)
	if v.last == nil {
		t.Fatal("No child was visited")
	}
	if v.last != ty {
		t.Fatal("Only root should be visited:", v.last.String())
	}
}
