let a = Array.create 0 42 in
let rec nest x = Array.create 0 x in
let rec nest_a _ = Array.create 0 a in
let b = nest a in
let c = nest_a () in
let d = Array.create 0 nest in
print_int (Array.size d)


