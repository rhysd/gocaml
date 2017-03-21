print_int 42;
print_bool true;
print_float 3.14;
print_str "PIYOOO";
println_bool false;
println_int 10;
println_float 3.14;
println_str "POYOOO";

println_int (float_to_int 3.14);
println_float ((int_to_float 3) +. 1.1);
println_int (str_length "aiueo");

let a = int_to_str 42 in
println_str a;
println_int (str_length a);
let a = int_to_str (-10) in
println_str a;
println_int (str_length a);

let a = float_to_str 3.14 in
println_str a;
println_int (str_length a);
let a = float_to_str (-.1.123) in
println_str a;
println_int (str_length a);

println_int (str_to_int "1234");
println_int (str_to_int "-1234");

println_float (str_to_float "3.14");
println_float (str_to_float "-1.4142");

println_int (bit_and 2 1);
println_int (bit_or 3 1);
println_int (bit_xor 3 1);
println_int (bit_lsft 1 10);
println_int (bit_rsft 256 4);
println_int (bit_inv 0);

()
