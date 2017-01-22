(* http://smpl.seesaa.net/article/9342186.html#comment *)
let rec f _ = 12345 in
let rec g y = y + 1 in
let z = Array.make 10 1 in
let x = f () in
let y = 67890 in
let z0 = z.(0) in
let z1 = z0 + z0 in
let z2 = z1 + z1 in
let z3 = z2 + z2 in
let z4 = z3 + z3 in
let z5 = z4 + z4 in
let z6 = z5 + z5 in
let z7 = z6 + z6 in
let z8 = z7 + z7 in
let z9 = z8 + z8 in
let z10 = z9 + z9 in
let z11 = z10 + z10 in
let z12 = z11 + z11 in
let z13 = z12 + z12 in
let z14 = z13 + z13 in
let z15 = z14 + z14 in
print_int
  (if z.(1) = 0 then g y else
  z0 + z1 + z2 + z3 + z4 + z5 + z6 + z7 +
    z8 + z9 + z10 + z11 + z12 + z13 + z14 + z15 + x)
