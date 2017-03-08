println_bool (true && true);
println_bool (true && false);
println_bool (false && true);
println_bool (false && false);
println_bool (true || true);
println_bool (true || false);
println_bool (false || true);
println_bool (false || false);

let a = true || true && false in
let b = true || false && false in
let c = false && true || true in
let d = false && false || true in
println_bool a;
println_bool b;
println_bool c;
println_bool d;

let e = false || true || false in
let f = true && true && false in
println_bool e;
println_bool f;

let rec t _ = true in
let rec f _ = false in
println_bool ((t ()) && (f ()));
print_bool ((t ()) || (f ()))
