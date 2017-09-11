(*
 * Rust example written in gocaml
 * http://www.rust-lang.org/
 *)

let prog = "+ + * - /" in
let finish = str_length prog in
let rec char_at s idx = str_sub s idx (idx + 1) in

let rec calc acc pc =
    if pc = finish then acc else
    let ch = char_at prog pc in
    if ch = "+" then calc (acc+1) (pc+1) else
    if ch = "-" then calc (acc-1) (pc+1) else
    if ch = "*" then calc (acc*2) (pc+1) else
    if ch = "/" then calc (acc/2) (pc+1) else
    calc acc (pc+1)
in
print_str "The problem \"";
print_str prog;
print_str "\" calculates the value ";
println_int (calc 0 0)
