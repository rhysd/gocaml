let rec inprod v1 v2 acc i =
  if i < 0 then acc else
  inprod v1 v2 (acc +. v1.(i) *. v2.(i)) (i - 1) in
let v1 = Array.make 3 1.23 in
let v2 = Array.make 3 4.56 in
print_int (truncate (1000000. *. inprod v1 v2 0. 2))
