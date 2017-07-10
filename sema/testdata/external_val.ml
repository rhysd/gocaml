external x: int = "c_x";
external y: int = "c_y";
let i = x + y in
let j = x in
let k = y in
println_int (i + j + k)
