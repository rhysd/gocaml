(* Make a function for the case where both side is None *)
let rec f a b =
    a = b
in
println_bool (f (Some 42) (Some 42));
println_bool (f (Some 21) (Some 42));
println_bool (f None (Some 42));
println_bool (f (Some 42) None);
println_bool (f None None);
println_str "";
let rec f a b =
    a <> b
in
println_bool (f (Some 42) (Some 42));
println_bool (f (Some 21) (Some 42));
println_bool (f None (Some 42));
println_bool (f (Some 42) None);
println_bool (f None None);
println_str "";

println_bool ((Some (Some 42)) = (Some (Some 42)));
println_bool ((Some (Some 42)) = (Some (Some 21)));
println_bool ((Some None) = (Some (Some 21)));
println_bool (None = (Some (Some 21)));
println_str "";

println_bool ((Some 3.14) = (Some 3.14));
println_bool ((Some f) = (Some f));
println_bool ((Some true) = (Some true));
println_bool ((Some "foo") = (Some "foo"));
println_bool ((Some ()) = (Some ()));
()
