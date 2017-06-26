let x = float_to_int 1.23 in
let y = float_to_int 4.56 in
let z = float_to_int (-.7.89) in
print_int
  ((if z < 0 then y else x) +
   (if x > 0 then z else y) +
   (if y < 0 then x else z))
