(*
 * 64bit Mersenne Twister random number generator.
 * http://www.math.sci.hiroshima-u.ac.jp/~m-mat/MT/VERSIONS/C-LANG/mt19937-64.c
 *)
let rec make_rng seeds =
    let NN = 312 in
    let MM = 156 in
    let MATRIX_A = -5403634167711393303 (* 0xB5026F5AA96619E9 *) in
    let UM = -2147483648 (* 0xFFFFFFFF80000000 *) in
    let LM = 2147483647 (* 0x7FFFFFFF *) in
    let mt = Array.make NN 0 in
    let rec init_genrand64 seed =
        mt.(0) <- seed;
        let rec f n =
            if n = NN then () else
            (mt.(n) <- 6364136223846793005 * (bit_xor mt.(n-1) (bit_rsft mt.(n-1) 62)) + n; f (n+1))
        in
        f 1
    in
    let rec init_by_array64 init_key =
        let key_length = Array.length init_key in
        init_genrand64 19650218;
        let rec f i j k =
            if k = 0 then i else (
                mt.(i) <- (bit_xor mt.(i) ((bit_xor mt.(i-1) (bit_rsft mt.(i-1) 62)) * 3935559000370003845)) + init_key.(j) + j;
                let i = i + 1 in
                let j = j + 1 in
                if i >= NN then
                    mt.(0) <- mt.(NN-1);
                    f 1 j (k-1)
                else if j >= key_length then
                    f i 0 (k-1)
                else
                    f i j (k-1)
            )
        in
        let i = f 1 0 (if NN > key_length then NN else key_length) in
        let rec f i k =
            if k = 0 then () else (
                mt.(i) <- (bit_xor mt.(i) ((bit_xor mt.(i-1) (bit_rsft mt.(i-1) 62)) * 2862933555777941757)) - i;
                if i+1 >= NN then
                    mt.(0) <- mt.(NN-1);
                    f 1 (k-1)
                else
                    f (i+1) (k-1)
            )
        in
        f i (NN-1);
        mt.(0) <- bit_lsft 1 63 (*MSB is 1; assuring non-zero initial array*)
    in
    let mag01 = [| 0; MATRIX_A |] in
    let mti = [| NN+1 |] in
    let rec genrand64 _ =
        if mti.(0) >= NN then
            let rec f i =
                if i = (NN - MM) then () else
                let x = bit_or (bit_and mt.(i) UM) (bit_and mt.(i+1) LM) in
                mt.(i) <- bit_xor (bit_xor mt.(i + MM) (bit_rsft x 1)) mag01.(bit_and x 1);
                f (i+1)
            in
            f 0;
            let rec f i =
                if i = (NN-1) then () else
                let x = bit_or (bit_and mt.(i) UM) (bit_and mt.(i+1) LM) in
                mt.(i) <- bit_xor (bit_xor mt.(i + (MM - NN)) (bit_rsft x 1)) mag01.(bit_and x 1);
                f (i + 1)
            in
            f (NN-MM);
            let x = bit_or (bit_and mt.(NN-1) UM) (bit_and mt.(0) LM) in
            mt.(NN-1) <- bit_xor (bit_xor mt.(MM-1) (bit_rsft x 1)) mag01.(bit_and x 1);
            mti.(0) <- 0
        else ();
        let x = mt.(mti.(0)) in
        let x = bit_xor x (bit_and (bit_rsft x 29) 6148914691236517205 (* 0x5555555555555555 *)) in
        let x = bit_xor x (bit_and (bit_lsft x 17) 8202884508482404352 (* 0x71D67FFFEDA60000 *)) in
        let x = bit_xor x (bit_and (bit_lsft x 37) (-2270628950310912) (* 0xFFF7EEE000000000 *)) in
        let x = bit_xor x (bit_rsft x 43) in
        mti.(0) <- mti.(0) + 1;
        x
    in
    init_by_array64 seeds;
    genrand64
in

let seeds = [| 123; 234; 345; 456 |] in
let gen = make_rng seeds in

println_int (gen ());
println_int (gen ());
println_int (gen ());
println_int (gen ());
println_int (gen ());
println_int (gen ())
