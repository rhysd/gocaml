type foo = int;
type bar = foo array;
type piyo = foo option;
type foo = float;
let b:bar = [| 1; 2 |] in
let p:piyo = Some 3 in
let rec f (p:piyo) = p in
f (Some 3);
let rec f x: foo = x in
f 3.1;
let foo: foo = (1.41: foo) in
let rec foo x: foo = x in
()
