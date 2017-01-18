let rec f _ = 123 in
let rec g _ = 456 in
let rec h _ = 789 in

let x = f () in
let y = g () in
print_int ((if h () = 0 then x + 1 else y + 2) + x + y)
(* then節ではxがr0でyがスタックに、else節ではyがr0でxがスタックにある *)
