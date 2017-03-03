let rec compose f g =
    let rec h x = f (g x) in
    h in
let rec twice x = x + x in
let rec plus10 x = x + 10 in
let f = compose twice plus10 in
print_int (f 20)

