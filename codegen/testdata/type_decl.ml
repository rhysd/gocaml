type point = int * int;
let rec x (p:point): int = let (x, _) = p in x in
let rec y (p:point): int = let (_, y) = p in y in
let p: point = (10, 20) in
println_int (x p);
print_int (y p)
