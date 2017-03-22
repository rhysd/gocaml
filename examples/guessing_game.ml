let rec rand x =
    let high = x / 127773 in
    let low = x % 127773 in
    let t = 16807 * low - 2836 * high in
    if t <= 0 then t + 9223372036854775807 else t
in
let n = rand (time_now ()) % 100 + 1 in
let rec play count =
    print_str "Guess a number (1~100): ";
    let input = str_to_int (get_line ()) in
    if input > 100 || 0 >= input then
        println_str "Please enter 1~100";
        play count
    else
        let msg =
            if input > n then "Too large!" else
            if input < n then "Too small!" else
            "gotcha!"
        in
        println_str msg;
        if input = n then count else play (count+1)
in
let tried = play 1 in
print_str "You hit a correct answer in "; print_int tried; println_str " times"
