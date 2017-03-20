print_str "argc: "; println_int (Array.length argv);
let prog = argv.(0) in
let size = str_length prog in
(* prog is a full path to executable. Check only file extension. *)
print_str "prog: "; print_str (str_sub prog (size - 4) size)
