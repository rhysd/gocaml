let o: int option = Some 42 in
let o2: (int * unit) array option = None in
let rec f x = () in f (Some 42); f None; let a = None in f a;
()
