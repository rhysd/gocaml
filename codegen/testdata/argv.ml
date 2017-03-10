print_str "argc: "; println_int (Array.size argv);
let prog = argv.(0) in
let size = str_size prog in
(* prog is a full path to executable. Check only file extension. *)
print_str "prog: "; print_str (substr prog (size - 4) size)
