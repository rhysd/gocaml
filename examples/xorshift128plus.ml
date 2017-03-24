let rec xorshift128plus seed =
    let state = Array.make 2 0 in
    state.(0) <- bit_xor seed (-6314187572093295703) (* 0xAF4100491F9D38AF *);
    state.(1) <- bit_xor seed (-7552163386978529546) (* 0xD19D592CBD21E214 *);
    let rec gen _ =
        let x = state.(0) in
        let y = state.(1) in
        state.(0) <- y;
        let x = bit_lsft x 23 in
        state.(1) <- bit_xor x (bit_xor y (bit_xor (bit_rsft x 17) (bit_rsft y 26)));
        state.(1) + y
    in
    gen
in
let rand = xorshift128plus (time_now ()) in
println_int (rand ());
println_int (rand ());
println_int (rand ());
println_int (rand ());
println_int (rand ());
println_int (rand ());
println_int (rand ());
println_int (rand ());
()

