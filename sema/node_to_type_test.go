package sema

import (
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
	. "github.com/rhysd/gocaml/types"
	"github.com/rhysd/locerr"
	"strings"
	"testing"
)

func TestSuccess(t *testing.T) {
	any := &Var{}
	pos := locerr.Pos{}
	tok := &token.Token{
		Start: pos,
		End:   pos,
		File:  locerr.NewDummySource(""),
	}
	prim := func(name string) ast.Expr {
		return &ast.CtorType{
			nil,
			tok,
			nil,
			ast.NewSymbol(name),
		}
	}
	ctor := func(name string, child ast.Expr) ast.Expr {
		return &ast.CtorType{
			nil,
			tok,
			[]ast.Expr{child},
			ast.NewSymbol(name),
		}
	}
	decls := []*ast.TypeDecl{
		{tok, ast.NewSymbol("foo"), prim("int")},
		{tok, ast.NewSymbol("bar"), prim("foo")},
		{tok, ast.NewSymbol("piyo"), &ast.FuncType{
			[]ast.Expr{prim("int"), prim("foo")},
			prim("bar"),
		}},
	}

	cases := []struct {
		what string
		node ast.Expr
		want Type
	}{
		{
			what: "primitive",
			node: prim("int"),
			want: IntType,
		},
		{
			what: "_ (any)",
			node: prim("_"),
			want: any,
		},
		{
			what: "tuple",
			node: &ast.TupleType{[]ast.Expr{
				prim("float"),
				prim("string"),
				prim("bool"),
			}},
			want: &Tuple{[]Type{
				FloatType,
				StringType,
				BoolType,
			}},
		},
		{
			what: "array",
			node: ctor("array", prim("float")),
			want: &Array{FloatType},
		},
		{
			what: "option",
			node: ctor("option", prim("unit")),
			want: &Option{UnitType},
		},
		{
			what: "fun",
			node: &ast.FuncType{
				[]ast.Expr{prim("int"), prim("bool")},
				prim("float"),
			},
			want: &Fun{FloatType, []Type{IntType, BoolType}},
		},
		{
			what: "nested any",
			node: ctor("array", prim("_")),
			want: &Array{any},
		},
		{
			what: "nested any in fun",
			node: &ast.FuncType{
				[]ast.Expr{prim("_"), prim("_")},
				prim("_"),
			},
			want: &Fun{any, []Type{any, any}},
		},
		{
			what: "nested any in tuple",
			node: &ast.TupleType{[]ast.Expr{prim("_"), prim("_")}},
			want: &Tuple{[]Type{any, any}},
		},
		{
			what: "compound type",
			node: &ast.FuncType{
				[]ast.Expr{
					&ast.TupleType{[]ast.Expr{
						prim("_"),
						ctor("array", &ast.TupleType{[]ast.Expr{
							prim("int"),
							prim("bool"),
						}}),
					}},
					ctor("option", ctor("option", ctor("array", prim("_")))),
				},
				&ast.FuncType{
					[]ast.Expr{
						prim("_"),
					},
					prim("unit"),
				},
			},
			want: &Fun{
				&Fun{
					UnitType,
					[]Type{any},
				},
				[]Type{
					&Tuple{[]Type{
						any,
						&Array{
							&Tuple{[]Type{
								IntType,
								BoolType,
							}},
						},
					}},
					&Option{
						&Option{
							&Array{
								any,
							},
						},
					},
				},
			},
		},
		{
			what: "simple aliased type",
			node: prim("foo"),
			want: IntType,
		},
		{
			what: "nested aliased type",
			node: prim("bar"),
			want: IntType,
		},
		{
			what: "alias in parameter",
			node: ctor("array", prim("bar")),
			want: &Array{IntType},
		},
		{
			what: "multiple aliases in tuple",
			node: &ast.TupleType{[]ast.Expr{
				prim("foo"),
				prim("bar"),
				prim("piyo"),
			}},
			want: &Tuple{[]Type{
				IntType,
				IntType,
				&Fun{
					IntType,
					[]Type{IntType, IntType},
				},
			}},
		},
	}

	c, err := newNodeTypeConv(decls)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		t.Run(tc.what, func(t *testing.T) {
			have, err := c.nodeToType(tc.node, 0)
			if err != nil {
				t.Fatal(tc.node.Name(), "caused an error:", err)
			}
			if !Equals(have, tc.want) {
				t.Fatal("Converted into unexpected type. want:", tc.want.String(), ", have:", have.String())
			}
		})
	}
}

func TestErrors(t *testing.T) {
	pos := locerr.Pos{}
	tok := &token.Token{
		Start: pos,
		End:   pos,
		File:  locerr.NewDummySource(""),
	}
	prim := func(name string) ast.Expr {
		return &ast.CtorType{
			nil,
			tok,
			nil,
			ast.NewSymbol(name),
		}
	}
	cases := []struct {
		what string
		node ast.Expr
		msg  string
	}{
		{
			what: "unknown type",
			node: prim("foo"),
			msg:  "Unknown type constructor 'foo'",
		},
		{
			what: "invalid array type params",
			node: &ast.CtorType{
				tok,
				tok,
				[]ast.Expr{prim("int"), prim("bool")},
				ast.NewSymbol("array"),
			},
			msg: "'array' only has 1 type parameter",
		},
		{
			what: "invalid option type params",
			node: &ast.CtorType{
				tok,
				tok,
				[]ast.Expr{prim("int"), prim("bool")},
				ast.NewSymbol("option"),
			},
			msg: "'option' only has 1 type parameter",
		},
		{
			what: "unknown type (tuple elem)",
			node: &ast.TupleType{[]ast.Expr{prim("foo")}},
			msg:  "Unknown type constructor 'foo'",
		},
		{
			what: "unknown type (fun param)",
			node: &ast.FuncType{[]ast.Expr{prim("foo")}, prim("unit")},
			msg:  "Unknown type constructor 'foo'",
		},
		{
			what: "unknown type (fun ret)",
			node: &ast.FuncType{[]ast.Expr{prim("_")}, prim("foo")},
			msg:  "Unknown type constructor 'foo'",
		},
	}
	for _, tc := range cases {
		t.Run(tc.what, func(t *testing.T) {
			c, err := newNodeTypeConv([]*ast.TypeDecl{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = c.nodeToType(tc.node, 0)
			if err == nil {
				t.Fatal("Error did not occur")
			}
			if !strings.Contains(err.Error(), tc.msg) {
				t.Fatal("Unexpected error message:", err)
			}
		})
	}
}

func TestInvalidAliases(t *testing.T) {
	pos := locerr.Pos{}
	tok := &token.Token{
		Start: pos,
		End:   pos,
		File:  locerr.NewDummySource(""),
	}
	prim := func(name string) ast.Expr {
		return &ast.CtorType{
			nil,
			tok,
			nil,
			ast.NewSymbol(name),
		}
	}

	cases := []struct {
		what  string
		decls []*ast.TypeDecl
		msg   string
	}{
		{
			what: "invalid aliased type",
			decls: []*ast.TypeDecl{
				{tok, ast.NewSymbol("foo"), prim("piyo")},
			},
			msg: "Type declaration 'foo'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.what, func(t *testing.T) {
			_, err := newNodeTypeConv(tc.decls)
			if err == nil {
				t.Fatal("Error expected but there is no error")
			}
			if !strings.Contains(err.Error(), tc.msg) {
				t.Fatal("Unexpected error", err, "...it should contain:", tc.msg)
			}
		})
	}
}
