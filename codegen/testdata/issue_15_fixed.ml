let rec go xs f =
    let x = xs.(0) in
    (f xs.(0)) + (f x)
in
let a = Array.make 1 3 in
let b = go a (fun x -> x) in
print_int b
