let rec sum acc x =
  if x <= 0 then acc else
  sum (acc + x) (x - 1) in
print_int (sum 0 10000)
