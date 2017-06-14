let rec a (x:int) y = x + y in
let rec b (x:int) (y:int): int = x + y in
let rec c x y: int = x + y in
let d = fun (x:int) y -> x + y in
let e = fun (x:int) (y:int): int -> x + y in
let f = fun x y: int -> x + y in
let g = fun _: (int -> int -> int) -> f in
()
