let a = 42 in
let rec f x = if x = 0 then 0 else a + (f (x - 1)) in
print_int (f 3)
