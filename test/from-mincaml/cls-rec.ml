(* 自由変数のある再帰関数 *)
let x = 10 in
let rec f y =
  if y = 0 then 0 else
  x + f (y - 1) in
print_int (f 123)
