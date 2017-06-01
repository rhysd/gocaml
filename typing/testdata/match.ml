let s: string = match Some 42 with
  | Some i -> let j: int = i in  "ok"
  | None   -> "not ok"
in ()
