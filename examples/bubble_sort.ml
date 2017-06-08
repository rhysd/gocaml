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
let a = [| 4; 8; 1; 8; 3; 0; 5; 6; 3; 0 |] in
show_array (bubble_sort a)
