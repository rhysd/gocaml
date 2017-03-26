let o = Some 42 in
let i = match o with
  | Some i -> i
  | None   -> 0
in print_int i;
let o = Some (Some 42) in
match o with
  | Some i -> i
  | None   -> 0;
let o = None in
match o with
  | Some i -> i
  | None   -> 0;
()
