(* 「素朴」なknown function optimizationでは駄目な場合 *)
(* Cf. http://www.yl.is.s.u-tokyo.ac.jp/~sumii/pub/compiler-enshu-2002/Mail/8 *)
let rec f x = x + 123 in
let rec g y = f in
print_int ((g 456) 789)
