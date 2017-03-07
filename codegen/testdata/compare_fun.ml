let rec f x = x + x in
let rec g x = x + x in
let a = 42 in
let rec h x = a + x in
let rec i x = a + x in
let rec getf _ = f in
let rec geth _ = h in
let rec dummy b = () in
let rec dummycl b = a; () in
println_bool (f = f);
println_bool (f <> f);
println_bool (f = g);
println_bool (f <> g);
println_bool (h = h);
println_bool (h <> h);
println_bool (h = i);
println_bool (h <> i);
println_bool (g = h);
println_bool (g <> h);
println_bool ((getf ()) = f);
println_bool ((getf ()) <> f);
println_bool ((getf ()) = g);
println_bool ((getf ()) <> g);
println_bool ((geth ()) = h);
println_bool ((geth ()) = g);
println_bool (dummy = println_bool);
print_bool (dummycl = println_bool)

