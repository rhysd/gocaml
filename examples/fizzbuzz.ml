let rec fizzbuzz max =
    let rec fb n =
        if n % 15 = 0 then println_str "fizzbuzz" else
        if n % 3 = 0  then println_str "fizz" else
        if n % 5 = 0  then println_str "buzz" else
        println_int n
    in
    let rec go n =
        if n = max then () else
        (fb n; go (n+1))
    in
    go 1
in
let max = 100 in
fizzbuzz max
