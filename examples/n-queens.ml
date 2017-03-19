(*
 * Solve N-Queens puzzle with back-tracking
 * https://en.wikipedia.org/wiki/Eight_queens_puzzle
 *)
let rec make_board size =
    let ret = Array.make size (Array.make 0 0) in
    let rec set idx =
        if idx < 0 then () else
        (ret.(idx) <- Array.make size 0;
         set (idx - 1))
    in
    set (size - 1);
    ret
in
let rec n_queens n =
    let SOLVED = true in
    let FAILED = false in
    let QUEEN = -1 in
    let board = make_board n in
    let rec in_board x y = x >= 0 && y >= 0 && n > x && n > y in
    let rec update x y delta =
        let rec f x y dx dy =
            let x = x + dx in
            let y = y + dy in
            if not (in_board x y) then () else
            (board.(x).(y) <- board.(x).(y) + delta; f x y dx dy)
        in
        f x y 1 0;
        f x y 1 1;
        f x y 0 1;
        f x y (-1) 1;
        f x y (-1) 0;
        f x y (-1) (-1);
        f x y 0 (-1);
        f x y 1 (-1);
        board.(x).(y) <- board.(x).(y) + delta
    in
    let rec put_queen x y = update x y 1 in
    let rec remove_queen x y = update x y (-1) in
    let rec solve nth x y =
        if not in_board x y then FAILED else
        let rec go_next _ =
            if x < n then
                solve nth (x+1) y
            else
                solve nth 0 (y+1)
        in
        if board.(x).(y) > 0 then
            go_next ()
        else (
            put_queen x y;
            let nth = nth + 1 in
            if nth >= n || solve nth 0 (y+1) then
                board.(x).(y) <- QUEEN;
                SOLVED
            else
                (remove_queen x y; go_next ())
        )
    in
    if solve 0 0 0 then
        (* When answer was found, show fancy output. *)
        let rec show _ =
            let rec show_cell v = print_str (if v = QUEEN then "x" else "."); print_str " " in
            let rec show_y y =
                if y >= n then () else
                let rec show_x x =
                    if x >= n then () else (show_cell board.(x).(y); show_x (x+1))
                in
                (show_x 0; print_str "\n"; show_y (y+1))
            in
            show_y 0
        in
        show ()
    else
        println_str "No answer"
in
let rec usage _ = print_str "Usage: "; print_str argv.(0); println_str " NUMBER" in
if Array.length argv = 1 then usage () else
let n = str_to_int (argv.(1)) in
if n = 0 then usage () else
n_queens n

(* Output of `./n-queens 8`:

  x . . . . . . .
  . . . . x . . .
  . . . . . . . x
  . . . . . x . .
  . . x . . . . .
  . . . . . . x .
  . x . . . . . .
  . . . x . . . .

*)
