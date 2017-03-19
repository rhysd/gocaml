let arr = Array.make 1 0 in
let rec f x =
    arr.(0) <- 42
in
f 0;
print_int arr.(0)
