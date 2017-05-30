let x: int = 42 in
let u: unit = () in
let y: int array = Array.make 1 1 in
let t: int * bool = 42, true in
let o: int option = Some(42) in
let f: int -> bool = fun x -> x <> 1 in
let (a, b): int * bool = 10, false in
let g: int -> bool -> int * bool = fun i b -> i, b in
()
