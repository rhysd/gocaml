let rec f _ = 123 in
let rec g _ = 456 in
let rec h _ = 789 in

let x = f () in
let y = g () in
print_int ((if h () = 0 then x - y else y - x) + x + y)
(* then節ではxがr0でyがr1に、else節ではyがr0でxがr1にある *)
