(* Shallow multi-dimension array *)
let rec shallow_dim2 m n x = Array.create m (Array.create n x) in
let a = shallow_dim2 3 2 42 in
println_int a.(0).(1);

a.(0).(1) <- 21;
println_int a.(0).(1);
println_int a.(1).(1);

a.(1).(1) <- a.(0).(1) + a.(1).(1);
println_int a.(1).(1);
println_int a.(2).(1);
println_int a.(0).(1);

(* Deep-copied multi-dimension array *)
let rec deep_dim2 m n x =
    let dummy = Array.create 0 x in
    let buf = Array.create m dummy in
    let rec set_elems m =
        if m < 0 then () else (
            buf.(m) <- Array.create n x;
            set_elems (m - 1)
        )
    in
    set_elems (m - 1);
    buf
in

let b = deep_dim2 3 2 42 in
println_int b.(0).(1);

b.(0).(1) <- 21;
println_int b.(0).(1);

b.(1).(1) <- b.(0).(1) + b.(1).(1);
println_int b.(1).(1);
println_int b.(2).(0);
println_int b.(0).(1);

(* Assign dynbamic length array to 2-dim array *)
b.(2) <- Array.create 11 99;
println_int b.(2).(0);
println_int b.(2).(8);

println_int (Array.size b);
println_int (Array.size (b.(0)));
print_int (Array.size (b.(2)))

