let rec a (x:int) = x in
let rec b x: int = x in
let c = fun (x:int) -> x in
let d = fun x: int -> x in
let e = fun (x:_ option) -> match x with Some(i) -> -i | None -> 0 in
()
