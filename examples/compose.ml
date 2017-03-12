let rec twice x = x *. 2.0 in
let rec squre x = x *. x in
let rec compose f g =
    let rec h x = (f (g x)) in
    h
in
let f = compose (compose squre twice) twice in
println_float (f 10.0)
(* Output: 1600 *)
