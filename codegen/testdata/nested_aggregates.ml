let pairs = Array.make 4 (1, true, 3.14) in
let rec show t =
    let (a, b, c) = t in
    println_int a;
    println_bool b;
    println_float c;
    ()
in
show (pairs.(1));

pairs.(1) <- (42, false, 1.1);
show (pairs.(0));
show (pairs.(1));

let arrays = (Array.make 2 true, Array.make 3 42, Array.make 4 1.1) in
let rec show x =
    let (a, b, c) = x in
    println_bool a.(0);
    println_int b.(1);
    println_float c.(2);
    ()
in
show arrays;
let (a, b, c) = arrays in
a.(0) <- false;
b.(1) <- -3;
c.(2) <- 3.14;
show arrays;

let rec f x =
    let (a, b, c) = pairs.(0) in x + a
in
println_int (f 10);
let rec g x = let (a, b, c) = arrays in c.(3) *. x in
println_float (g 3.3);

let nested = Array.make 3 (true, Array.make 2 (false, Array.make 1 (1, 3.14))) in
let (_, a) = nested.(1) in
let (_, a) = a.(0) in
let (_, f) = a.(0) in
println_float f;

let zero_length = Array.make 0 (true, Array.make 0 (false, true)) in

()
