(* basic *)
let o = Some 42 in
let rec f x = match x with
    | Some i -> println_int i
    | None -> println_str "none"
in
f None; (* can be inferred from signature of println_int *)
f o;

(* nested *)
let rec f x = match x with
    | Some x -> match x with
        | Some i -> println_int i
        | None -> println_str "none2"
    | None -> println_str "none1"
in
let o = Some (Some 42) in
f o;
let o = Some None in
f o;
f None;

(* return option *)
let rec f x = Some x in
match f 10 with Some(i) -> println_int i | None -> println_str "oops";

(* capture option value *)
let o = Some 3.14 in
let rec f x = match o with
  | Some f -> f +. x
  | None   -> -.x
in
println_float (f 1.1);

(* capture dereferenced variable *)
let rec f o = match o with
    | Some i -> let rec f x = x + i in f
    | None   -> let rec f x = -x in f
in
println_int ((f (Some 42)) 11);
println_int ((f None) 11);

(* check contents with 'match' expression *)
let rec is_some o = match o with Some _ -> true | None -> false in
println_bool (is_some (Some f));
println_bool (is_some None);

(* check contents with operator <> (not equal) *)
let rec is_some o = o <> None in
println_bool (is_some (Some true));
println_bool (is_some None);

(* tuple *)
let t = (Some 4, None, Some (1, "one")) in
let (a, b, c) = t in
println_int (match a with Some i -> i | None -> -99);
println_int (match b with Some i -> i | None -> -99);
match c with
  | Some pair ->
    let (i, s) = pair in
    println_int i;
    println_str s
  | None ->
    println_str "ooooops!";
let o = None in
match o with Some p -> let (_, _): int * int = p in () | None -> println_str "none of tuple!";

(* array *)
let arr = Array.make 7 None in
arr.(3) <- Some (Array.make 7 3.14);
println_float (match arr.(1) with Some a -> a.(0) | None -> 0.1);
println_float (match arr.(3) with Some a -> a.(0) | None -> 0.1);
()

