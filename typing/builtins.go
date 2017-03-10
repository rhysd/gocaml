package typing

func builtinPopulatedTable() map[string]Type {
	table := map[string]Type{
		"print_int":     &Fun{UnitType, []Type{IntType}},
		"print_bool":    &Fun{UnitType, []Type{BoolType}},
		"print_float":   &Fun{UnitType, []Type{FloatType}},
		"print_str":     &Fun{UnitType, []Type{StringType}},
		"println_int":   &Fun{UnitType, []Type{IntType}},
		"println_bool":  &Fun{UnitType, []Type{BoolType}},
		"println_float": &Fun{UnitType, []Type{FloatType}},
		"println_str":   &Fun{UnitType, []Type{StringType}},
		"float_to_int":  &Fun{IntType, []Type{FloatType}},
		"int_to_float":  &Fun{FloatType, []Type{IntType}},
		"str_size":      &Fun{IntType, []Type{StringType}},
		"__str_equal":   &Fun{BoolType, []Type{StringType, StringType}},
		"str_concat":    &Fun{StringType, []Type{StringType, StringType}},
		"substr":        &Fun{StringType, []Type{StringType, IntType, IntType}},
		"int_to_str":    &Fun{StringType, []Type{IntType}},
		"float_to_str":  &Fun{StringType, []Type{FloatType}},
		"str_to_int":    &Fun{IntType, []Type{StringType}},
		"str_to_float":  &Fun{FloatType, []Type{StringType}},
		"get_line":      &Fun{StringType, []Type{UnitType}},
	}
	return table
}
