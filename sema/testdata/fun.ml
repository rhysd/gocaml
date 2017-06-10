let rec foo x y = x + y in
let bar: int -> int -> int = foo in
let piyo: bool -> int -> int * bool = fun x y -> (y, x) in
()
