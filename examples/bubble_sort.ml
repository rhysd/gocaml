let rec show_array a =
    let size = Array.length a in
    let rec show_elems idx =
        if idx >= size then () else
        (print_int a.(idx); print_str " "; show_elems (idx + 1))
    in
    print_str "[ ";
    show_elems 0;
    println_str "]"
in
let rec bubble_sort xs =
    let len = Array.length xs in
    let rec swap idx max =
        if max - 1 <= idx then () else
        let l = xs.(idx) in
        let r = xs.(idx+1) in
        if l <= r then swap (idx+1) max else
        (xs.(idx) <- r;
         xs.(idx+1) <- l;
         swap (idx+1) max)
    in
    let rec go idx =
        if idx >= (len - 1) then () else
        (swap 0 (len - idx); go (idx + 1))
    in
    go 0;
    xs
in
let a = Array.make 10 0 in
a.(0) <- 4;
a.(1) <- 8;
a.(2) <- 1;
a.(3) <- 8;
a.(4) <- 3;
a.(5) <- 0;
a.(6) <- 5;
a.(7) <- 6;
a.(8) <- 3;
a.(9) <- 0;
show_array (bubble_sort a)
