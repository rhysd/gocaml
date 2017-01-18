(* thanks to https://twitter.com/gan13027830/status/791239623959687168 *)
(* exactly one more arguments than registers; does not copmile *)
let x = 42 in
let rec f y1 y2 y3 y4 y5 y6 = print_int x in
f 1 2 3 4 5 6
