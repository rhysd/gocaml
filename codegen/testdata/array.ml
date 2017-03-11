let a = Array.make 3 3.14 in
let e0 = a.(0) in
let e1 = a.(1) in
let e2 = a.(2) in
println_float e0;
println_float e1;
println_float e2;
println_float (a.(0) +. a.(1) +. a.(2));

let b = Array.make 3 true in
let rec first x = x.(0) in
println_bool ((first b) = true);

let c = Array.make 3 1.14 in
let rec g x = c.(0) -. x.(1) in
println_float (g a);

let d = Array.make 3 first in
println_bool ((d.(0)) b);

let rec getarr _ = Array.make 7 (-1) in
println_int (getarr ()).(1);

a.(1) <- 1.1;
println_float (a.(1));

b.(1) <- false;
println_bool (b.(1));

a.(0) <- (a.(1)) +. a.(0) +. c.(0);
println_float (a.(0));

println_int (Array.length b);
print_int (Array.length (getarr()))
