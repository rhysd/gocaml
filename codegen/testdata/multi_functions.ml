let rec f x = x + x in
let rec g x = (f x) + (f x) in
let a = 42 in
let rec h x = a + x in
let rec i x = a + (h x) in
let rec j x =
    let rec p x = f (g x) in
    let rec q x = h (i x) in
    (p x) + (q x) in
println_int (f 10);
println_int (g 10);
println_int (h 10);
println_int (i 10);
print_int (j 10)
