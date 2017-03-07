(* Return type of print_int cannot be inferred because print_int
 * is an external symbol. In this case, I decided to assign unit
 * to the return type.
 * In this case, ; will create $tmp = print_int 42 and type of
 * $tmp won't be inferred. So falling back into unit type. *)
print_int 42; ()
