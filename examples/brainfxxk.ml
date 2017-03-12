let rec char_at s idx = substr s idx (idx+1) in
let rec program tape =
    let mem = Array.make 30000 0 in
    let tape_size = str_length tape in
    let rec jump_fwd pc stack =
        let op = char_at tape pc in
        if op = "[" then jump_fwd (pc + 1) (stack + 1) else
        if op = "]" then (
            if stack = 0 then pc + 1 else jump_fwd (pc + 1) (stack - 1)
        ) else
        jump_fwd (pc + 1) stack
    in
    let rec jump_bkwd pc stack =
        let op = char_at tape pc in
        if op = "[" then (
            if stack = 0 then pc + 1 else jump_bkwd (pc - 1) (stack - 1)
        ) else
        if op = "]" then jump_bkwd (pc - 1) (stack + 1) else
        jump_bkwd (pc - 1) stack
    in
    let rec step pc ptr =
        if pc >= tape_size then () else
        let op = char_at tape pc in
        if op = ">" then step (pc + 1) (ptr + 1) else
        if op = "<" then step (pc + 1) (ptr - 1) else
        if op = "+" then (
            mem.(ptr) <- (mem.(ptr) + 1);
            step (pc + 1) ptr
        ) else
        if op = "-" then (
            mem.(ptr) <- (mem.(ptr) - 1);
            step (pc + 1) ptr
        ) else
        if op = "." then (
            print_str (from_char_code mem.(ptr));
            step (pc + 1) ptr
        ) else
        if op = "," then (
            mem.(ptr) <- (to_char_code (get_char ()));
            step (pc + 1) ptr
        ) else
        if op = "[" then (
            if mem.(ptr) = 0 then
                step (jump_fwd (pc + 1) 0) ptr
            else
                step (pc + 1) ptr
        ) else
        if op = "]" then (
            if mem.(ptr) <> 0 then
                step (jump_bkwd (pc - 1) 0) ptr
            else
                step (pc + 1) ptr
        ) else
        step (pc + 1) ptr
    in
    step 0 0
in
program "+++++++++[>++++++++>+++++++++++>+++++<<<-]>.>++.+++++++..+++.>-.------------.<++++++++.--------.+++.------.--------.>+."
