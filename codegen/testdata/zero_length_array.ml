let a = Array.make 0 42 in
let rec nest x = Array.make 0 x in
let rec nest_a _ = Array.make 0 a in
let b = nest a in
let c = nest_a () in
let d = Array.make 0 nest in
print_int (Array.length d)


