let rec f _ =
    let a = Array.make 100 "aaa" in
    a.(10)
in
let s = f () in
do_garbage_collection ();
println_str s;
disable_garbage_collection ();
let s = f () in
do_garbage_collection ();
println_str s;
enable_garbage_collection ()
