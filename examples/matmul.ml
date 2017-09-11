let rec each arr pred =
    let n = Array.length arr in
    let rec f i =
        if i = n then () else
        (pred arr.(i) i; f (i + 1))
    in
    f 0
in

let rec make_mat n m v =
    let ret = Array.make n (Array.make 0 v) in
    each ret (fun _ i -> ret.(i) <- Array.make m v);
    ret
in

let rec each2 arr pred =
    let m = Array.length arr in
    let n = Array.length arr.(0) in
    let rec f i =
        if i = m then () else
        let rec g j =
            if j = n then () else
            (pred arr.(i).(j) i j; g (j + 1))
        in
        (g 0;
        f (i + 1))
    in
    f 0
in

let rec transpose arr =
    let row = Array.length arr in
    let col = Array.length arr.(0) in
    let ret = make_mat col row 0.0 in
    each2 arr (fun e i j -> ret.(j).(i) <- arr.(i).(j));
    ret
in

let rec multiplication a b =
    let n = Array.length a.(0) in
    let ret = make_mat (Array.length a) (Array.length b) 0.0 in
    each a (fun ai i ->
        each b (fun bj j ->
            let rec h k =
                if k < 0 then 0.0 else
                (h (k - 1)) +. ai.(k) *. bj.(k)
            in
            ret.(i).(j) <- h n
        )
    );
    ret
in

let rec matmul a b =
    let b = transpose b in
    multiplication a b
in

let rec matgen n =
    let f = int_to_float n in
    let tmp = 1.0 /. f /. f in
    let a = make_mat n n 0.0 in
    each2 a (fun _ i j ->
        let x = int_to_float (i - j) in
        let y = int_to_float (i + j) in
        a.(i).(j) <- tmp *. x *. y
    );
    a
in

let n = 500 in
let a = matgen n in
let b = matgen n in
let c = matmul a b in
println_float c.(n / 2).(n / 2)
