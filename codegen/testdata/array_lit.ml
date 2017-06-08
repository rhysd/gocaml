let a = [| 1; 2; 3; |] in
println_int (Array.length a);
println_int a.(0);
println_int a.(1);
println_int a.(2);

let rec f _: int array = [| |] in
println_int (Array.length (f ()));

let a = [| [| 1 |]; [| 10; 11; |]; [| 20; 21; 22; |] |] in
println_int a.(0).(0);
println_int a.(1).(1);
println_int a.(2).(2);

let rec f b =
    let x = [| a.(1).(0) |] in
    [| x.(0) + a.(2).(1) + b.(1) |]
in
println_int (f [| 100; 200 |]).(0);

let a = [| (1, 3.14); (10, 1.0) |] in
let (i, f) = a.(1) in
a.(0) <- (i + 33, f);
let (i, _) = a.(0) in
println_int i;

let a = [| [| true |]; [| |]; [| false |] |] in
println_bool a.(0).(0);
println_bool a.(2).(0);
a.(1) <- [| false |];
println_int (Array.length a.(1));
println_bool a.(1).(0)
