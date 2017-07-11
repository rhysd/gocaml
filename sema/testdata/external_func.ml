external print: int -> string -> unit = "print";
external print2: int -> unit = "print2";
let a = print 42 "hello" in
let f = print in f 42 "foo";
let g = print2 in g 10;
()
