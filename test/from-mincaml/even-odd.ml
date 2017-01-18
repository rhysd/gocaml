let t = 123 in
let f = 456 in
let rec even x =
  let rec odd x =
    if x > 0 then even (x - 1) else
    if x < 0 then even (x + 1) else
    f in
  if x > 0 then odd (x - 1) else
  if x < 0 then odd (x + 1) else
  t in
print_int (even 789)
