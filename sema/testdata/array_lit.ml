let a = [| 1; 2; 3 |] in
let b = [| [|1|]; [| |]; |] in
let c = [| |] in
c.(0) <- a.(0);
let a = [| |] in
println_int a.(0)
