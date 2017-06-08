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
let rec quick_sort xs less =
    let rec swap i j =
        let tmp = xs.(i) in
        xs.(i) <- xs.(j);
        xs.(j) <- tmp
    in
    let rec go left right =
        if left >= right then () else
        let pivot = xs.((left + right) / 2) in
        let rec partition l r =
            let rec next_left i =
                (* pivot <= xs.(i) *)
                if not (less xs.(i) pivot) then i else
                next_left (i+1)
            in
            let rec next_right i =
                (* pivot >= xs.(i) *)
                if not (less pivot xs.(i)) then i else
                next_right (i-1)
            in
            let l = next_left l in
            let r = next_right r in
            if l >= r then (l, r) else
            (swap l r; partition (l+1) (r-1))
        in
        let (l, r) = partition left right in
        go left (l-1);
        go (r+1) right
    in
    go 0 (Array.length xs - 1);
    xs
in
let a = [| 4; 8; 1; 8; 3; 0; 5; 6; 3; 0 |] in
let sorted = quick_sort a (fun x y -> x < y) in
show_array sorted
