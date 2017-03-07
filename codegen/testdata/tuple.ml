let t = (1, true) in
let (i, b) = t in
println_int i;
println_bool b;

let (i, b, f) = (12, false, 3.14) in
println_int i;
println_bool b;
println_float f;

let t = (1, 2, (true, false)) in
let (i, j, t) = t in
let (_, b) = t in
println_bool b;

let rec first t = let (x, y) = t in x in
let x = first (3.14, true) in
println_float x;

let rec triple x y z = (x, y, z) in
let (b, i, f) = triple true 42 3.14 in
println_int i;

let t = (30, false) in
let rec add10fst t = let (i, x) = t in (i+10, x) in
let (i, _) = add10fst t in
println_int i;

let t = (10, 20) in
let rec add x = let (i, j) = x in let (p, q) = t in (i+p, j+q) in
let (i, j) = add (11, 22) in
println_int i;
println_int j;

let t = (1, (3.1, true)) in
let rec get _ = let (_, x) = t in x in
let (f, b) = get () in
println_float f;
println_bool b;

println_bool ((3.1, (1, true)) = (3.1, (1, true)));
println_bool (t <> (3, (1.0, false)))
