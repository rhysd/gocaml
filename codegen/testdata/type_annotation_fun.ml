let rec f (x:int): float = int_to_float x in
let g = fun (a:_ array) -> a.(0) in
let x = Array.make 1 42 in
print_int (g x)
