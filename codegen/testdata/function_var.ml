let a = 42 in
let rec f x = a + x in
let rec getf _ = f in
print_int ((getf ()) 13)
