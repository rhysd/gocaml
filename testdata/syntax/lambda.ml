let foo = fun x y -> x + y in
let rec print f a b = println_int(f a b) in
print(foo, 10, 42);
print(fun x y -> x - y, 10, 42)
