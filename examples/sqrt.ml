let rec sqrt x =
    let rec abs x = if x > 0.0 then x else -.x in
    let rec go z p =
        if abs (p -. z) <= 0.00001 then z else
        let (p, z) = z, z -. (z *. z -. x) /. (2.0 *. z) in
        go z p
    in
    go x 0.0
in
println_float (sqrt 10.0)
