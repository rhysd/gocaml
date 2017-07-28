let rec make_mat n m v =
    let ret = Array.make n (Array.make 0 v) in
    let rec set idx =
        if idx < 0 then () else (
            ret.(idx) <- Array.make m v;
            set (idx - 1)
        )
    in
    set (n - 1);
    ret
in

let rec transpose arr =
    let row = Array.length arr in
    let col = Array.length arr.(0) in
    let ret = make_mat col row 0.0 in
    let rec set_row r =
        if r < 0 then () else (
            let rec set_col c =
                if c < 0 then () else (
                    ret.(c).(r) <- arr.(r).(c);
                    set_col (c - 1)
                )
            in
            set_col (col - 1);
            set_row (r - 1)
        )
    in
    set_row (row - 1);
    ret
in

let rec multiplication a b =
    let m = Array.length a in
    let n = Array.length a.(0) in
    let p = Array.length b in
    let ret = make_mat m p 0.0 in
    let rec f i =
        if i < 0 then () else (
            let rec g j =
                if j < 0 then () else (
                    let ai = a.(i) in
                    let bj = b.(j) in
                    let rec h k =
                        if k < 0 then 0.0 else
                        (h (k - 1)) +. ai.(k) *. bj.(k)
                    in
                    ret.(i).(j) <- h n;
                    g (j - 1)
                )
            in
            g (p - 1);
            f (i - 1)
        )
    in
    f (m - 1);
    ret
in

let rec matmul a b =
    let b = transpose b in
    multiplication a b
in

let rec matgen n =
    let f = int_to_float n in
    let tmp = 1.0 /. f /. f in
    println_float tmp;
    let a = make_mat n n 0.0 in
    let rec set_i i =
        if i < 0 then () else (
            let rec set_j j =
                if j < 0 then () else (
                    let x = int_to_float (i - j) in
                    let y = int_to_float (i + j) in
                    a.(i).(j) <- tmp *. x *. y
                )
            in
            set_j (n - 1);
            set_i (i - 1)
        )
    in
    set_i (n - 1);
    a
in

let n = 500 in
let a = matgen n in
let b = matgen n in
let c = matmul a b in
println_float c.(n / 2).(n / 2)
