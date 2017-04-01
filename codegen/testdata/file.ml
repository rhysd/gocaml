let f = read_file "unknown_file" in
println_str (match f with Some c -> "found" | None -> "not found");

let f = read_file "testdata/test.txt" in
print_str (match f with Some c -> c | None -> "not found");

let b = write_file "testdata/piyo.txt" "this is test for write_file()" in
if not b then println_str "failed to write!" else
let f = read_file "testdata/piyo.txt" in
println_str (match f with Some c -> c | None -> "failed to open file")
