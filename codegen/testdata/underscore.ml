let a = 10 in
let b = 12 in
let _ = a + b in
let rec _ _ = a in
let rec f _ = b in
println_int (f ());
let t = 1, "aaa", true in
let (_, s, _) = t in
println_str s;
let rec f _ = 42 in
println_int (f true);
let f = fun _ -> true in
print_bool (f 3.14)

