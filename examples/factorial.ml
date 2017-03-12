let rec fact n =
    if n <= 0 then 1 else
    n * (fact (n - 1))
in
println_int (fact 10)
