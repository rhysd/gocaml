match Some 42 with
  | Some i -> "ok"
  | None   -> "not ok";
let s = match None with None -> "none" | Some i -> "some" in
match Some (Some s) with
    None -> Some 10
  | Some o -> Some 99;
match Some 42 with
  | Some(i) -> "ok"
  | None   -> "not ok";
let s = match None with None -> "none" | Some(i) -> "some" in
()
