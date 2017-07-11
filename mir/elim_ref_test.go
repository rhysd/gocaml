package mir

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rhysd/gocaml/types"
)

func TestEliminatingRefNew(t *testing.T) {
	i := func(i string, v Val) *Insn { return &Insn{Ident: i, Val: v} }
	cases := []struct {
		what  string
		table map[string]types.Type
		ext   map[string]types.Type
		block []*Insn
		want  []string
	}{
		{
			what: "binary operator",
			table: map[string]types.Type{
				"foo":  types.IntType,
				"t1":   types.IntType,
				"bar":  types.IntType,
				"t2":   types.IntType,
				"piyo": types.IntType,
			},
			block: []*Insn{
				i("foo", &Int{1}),
				i("bar", &Int{2}),
				i("t1", &Ref{"foo"}),
				i("t2", &Ref{"bar"}),
				i("piyo", &Binary{ADD, "t1", "t2"}),
			},
			want: []string{
				"piyo = binary + foo bar",
			},
		},
		{
			what: "unary operator",
			table: map[string]types.Type{
				"a":  types.FloatType,
				"t1": types.FloatType,
				"b":  types.FloatType,
			},
			block: []*Insn{
				i("a", &Float{3.1}),
				i("t1", &Ref{"a"}),
				i("b", &Unary{FNEG, "t1"}),
			},
			want: []string{
				"a = float 3.1",
				"b = unary -. a",
			},
		},
		{
			what: "if expression",
			table: map[string]types.Type{
				"a":  types.StringType,
				"t1": types.StringType,
				"t2": types.StringType,
				"t3": types.BoolType,
				"t4": types.StringType,
				"t5": types.StringType,
				"t6": types.StringType,
				"t7": types.StringType,
				"t8": types.StringType,
			},
			block: []*Insn{
				i("a", &String{"hello"}),
				i("t1", &Ref{"a"}),
				i("t2", &String{"bye"}),
				i("t3", &Binary{EQ, "t1", "t2"}),
				i("t4", &If{
					"t3",
					NewBlockFromArray("then", []*Insn{
						i("t5", &Ref{"a"}),
					}),
					NewBlockFromArray("else", []*Insn{
						i("t6", &Ref{"a"}),
						i("t7", &String{"foo"}),
						i("t8", &Binary{NEQ, "t6", "t7"}),
					}),
				}),
			},
			want: []string{
				"t3 = binary = a t2",
				"t5 = ref a",
				"t8 = binary <> a t7",
			},
		},
		{
			what: "function",
			table: map[string]types.Type{
				"x":  types.UnitType,
				"f":  &types.Fun{types.UnitType, []types.Type{types.UnitType}},
				"t2": types.BoolType,
				"t3": types.BoolType,
				"t4": types.BoolType,
				"t5": types.UnitType,
				"t6": types.UnitType,
				"t7": types.UnitType,
			},
			block: []*Insn{
				i("x", UnitVal),
				i("f", &Fun{
					[]string{"a"},
					NewBlockFromArray("f", []*Insn{
						i("t2", &Ref{"x"}),
						i("t3", &App{"f", []string{"t2"}, CLOSURE_CALL}),
					}),
					true,
				}),
				i("t4", &Bool{false}),
				i("t5", &App{"f", []string{"t4"}, DIRECT_CALL}),
				i("t6", UnitVal),
				i("t7", &Binary{ADD, "t5", "t6"}),
			},
			want: []string{
				"t3 = appcls f x",
				"t5 = app f t4",
				"t7 = binary + t5 t6",
			},
		},
		{
			what: "tuple",
			table: map[string]types.Type{
				"x":  types.IntType,
				"t1": types.IntType,
				"t2": types.IntType,
				"y":  &types.Tuple{[]types.Type{types.IntType, types.IntType}},
				"a":  types.IntType,
				"b":  types.IntType,
				"t3": types.IntType,
			},
			block: []*Insn{
				i("x", &Int{42}),
				i("t1", &Ref{"x"}),
				i("t2", &Int{1}),
				i("y", &Tuple{[]string{"t1", "t2"}}),
				i("a", &TplLoad{"y", 0}),
				i("b", &TplLoad{"y", 1}),
				i("t3", &Ref{"a"}),
			},
			want: []string{
				"t2 = int 1",
				"y = tuple x,t2",
				"a = tplload 0 y",
				"b = tplload 1 y",
				"t3 = ref a",
			},
		},
		{
			what: "array",
			table: map[string]types.Type{
				"i":   types.IntType,
				"t1":  types.IntType,
				"t2":  types.IntType,
				"a1":  &types.Array{types.IntType},
				"t3":  types.IntType,
				"t4":  types.IntType,
				"t5":  types.IntType,
				"t6":  types.IntType,
				"t7":  types.IntType,
				"a2":  &types.Array{types.IntType},
				"t8":  types.IntType,
				"t9":  types.IntType,
				"t10": types.IntType,
				"t11": types.IntType,
			},
			block: []*Insn{
				i("i", &Int{42}),
				i("t1", &Ref{"i"}),
				i("t2", &Ref{"i"}),
				i("a1", &Array{"t1", "t2"}),
				i("t3", &Ref{"a1"}),
				i("t4", &Ref{"i"}),
				i("t5", &ArrLoad{"t3", "t4"}),
				i("t6", &Ref{"i"}),
				i("t7", &Int{4}),
				i("a2", &ArrLit{[]string{"t6", "t7"}}),
				i("t8", &Ref{"a2"}),
				i("t9", &Ref{"i"}),
				i("t10", &Ref{"i"}),
				i("t11", &ArrStore{"t8", "t9", "t10"}),
			},
			want: []string{
				"i = int 42",
				"a1 = array i i",
				"t5 = arrload i a1",
				"a2 = arrlit i,t7",
				"t11 = arrstore i a2 i",
			},
		},
		{
			what: "external variable",
			table: map[string]types.Type{
				"t1": types.IntType,
				"t2": types.IntType,
				"t3": types.IntType,
			},
			ext: map[string]types.Type{
				"x": types.IntType,
			},
			block: []*Insn{
				i("t1", &XRef{"x"}),
				i("t2", &Int{0}),
				i("t3", &Binary{ADD, "t1", "t2"}),
			},
			want: []string{
				"t1 = xref x",
			},
		},
		{
			what: "external variable",
			table: map[string]types.Type{
				"i":  types.IntType,
				"t1": types.IntType,
				"t2": types.IntType,
			},
			ext: map[string]types.Type{
				"xf": &types.Fun{types.IntType, []types.Type{types.UnitType}},
			},
			block: []*Insn{
				i("i", &Int{0}),
				i("t1", &Ref{"i"}),
				i("t2", &App{"xf", []string{"t1"}, EXTERNAL_CALL}),
			},
			want: []string{
				"t2 = appx xf i",
			},
		},
		{
			what: "option value",
			table: map[string]types.Type{
				"i":  types.IntType,
				"t1": types.IntType,
				"o":  &types.Option{types.IntType},
				"t2": &types.Option{types.IntType},
				"t3": types.BoolType,
				"t4": types.IntType,
				"j":  types.IntType,
				"t5": types.IntType,
				"t6": types.IntType,
				"t7": types.IntType,
			},
			block: []*Insn{
				i("i", &Int{0}),
				i("t1", &Ref{"i"}),
				i("o", &Some{"t1"}),
				i("t2", &Ref{"o"}),
				i("t3", &IsSome{"t2"}),
				i("t4", &If{
					"t3",
					NewBlockFromArray("then", []*Insn{
						i("j", &DerefSome{"t2"}),
						i("t5", &Ref{"j"}),
						i("t6", &Unary{NEG, "t5"}),
					}),
					NewBlockFromArray("else", []*Insn{
						i("t7", &Ref{"i"}),
					}),
				}),
			},
			want: []string{
				"o = some i",
				"t3 = issome o",
				"t6 = unary - j",
				"t7 = ref i",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.what, func(t *testing.T) {
			if len(tc.want) == 0 {
				t.Fatal("no expectation")
			}
			env := types.NewEnv()
			for i, t := range tc.table {
				env.DeclTable[i] = t
			}
			for i, t := range tc.ext {
				env.Externals[i] = &types.External{t, i}
			}
			block := NewBlockFromArray("PROGRAM", tc.block)

			ElimRefs(block, env)

			var buf bytes.Buffer
			block.Println(&buf, env)
			have := buf.String()
			for _, want := range tc.want {
				if !strings.Contains(have, want) {
					t.Fatalf("'%s' does not contain '%s'", have, want)
				}
			}
		})
	}
}
