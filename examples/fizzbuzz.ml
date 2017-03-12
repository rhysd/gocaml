let rec fb n =
    if n % 15 = 0 then println_str "fizzbuzz" else
    if n % 3 = 0  then println_str "fizz" else
    if n % 5 = 0  then println_str "buzz" else
    println_int n
in
let rec fizzbuzz n =
    if n <= 0 then () else
    (fizzbuzz (n - 1); fb n)
in
fizzbuzz 100
