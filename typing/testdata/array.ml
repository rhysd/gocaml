let a: bool array = Array.make 3 true in
let b: (int * int) array = Array.make 3 (1, 1) in
let c: int array array = Array.make 3 (Array.make 3 1) in
let d: bool = a.(0) in
let e: unit = a.(0) <- false in
()
