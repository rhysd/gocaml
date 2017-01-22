(* http://blog.livedoor.jp/azounoman/archives/50392600.html *)
let rec f x0 =
  let x1 = x0 + 1 in
  let x2 = x1 + 1 in
  let x3 = x2 + 1 in
  let x4 = x3 + 1 in
  let x5 = x4 + 1 in
  let x6 = x5 + 1 in
  let x7 = x6 + 1 in
  let x8 = x7 + 1 in
  let x9 = x8 + 1 in
  let x10 = x9 + 1 in
  let x11 = x10 + 1 in
  let x12 = x11 + 1 in
  let x13 = x12 + 1 in
  let x14 = x13 + 1 in
  let x15 = x14 + 1 in
  let x16 = x15 + 1 in
  let x17 = x16 + 1 in
  let x18 = x17 + 1 in
  let x19 = x18 + x1 in
  let x20 = x19 + x2 in
  let x21 = x20 + x3 in
  let x22 = x21 + x4 in
  let x23 = x22 + x5 in
  let x24 = x23 + x6 in
  let x25 = x24 + x7 in
  let x26 = x25 + x8 in
  let x27 = x26 + x9 in
  let x28 = x27 + x10 in
  let x29 = x28 + x11 in
  let x30 = x29 + x12 in
  let x31 = x30 + x13 in
  let x32 = x31 + x14 in
  let x33 = x32 + x15 in
  let x34 = x33 + x16 in
  let x35 = x34 + x17 in
  let x36 = x35 + x0 in
  x1 + x2 + x3 + x4 + x5 + x6 + x7 + x8 + x9 + 
    x10 + x11 + x12 + x13 + x14 + x15 + x16 + x17 + x18 + x19 +
    x20 + x21 + x22 + x23 + x24 + x25 + x26 + x27 + x28 + x29 +
    x30 + x31 + x32 + x33 + x34 + x35 + x36 + x0 in
print_int (f 0)
