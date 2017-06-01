let f = fun x y -> (x: int), (y: bool) in
let g = (f: int -> bool -> int * bool) in
let (a, b) = g 1 true in
println_int (a: int);
println_bool (b: bool)
