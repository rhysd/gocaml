let a = "aaa" in
let b = "bbb" in
println_str a;
println_str b;

let c = str_concat a b in
println_str c;

let d = substr c 2 4 in
println_str d;

(* do not break previous strings *)
println_str a;
println_str b;
println_str c;

let e = "abcdef" in
println_str (substr e 0 2);
println_str (substr e 0 6);
println_str (substr e 0 8);
println_str (substr e 0 (-1));
println_str (substr e 0 0);
println_str (substr e 9 3);
println_str (substr e (-1) 4);
println_str (substr e 9 99);
println_str (substr e 5 6);
println_str (substr e 5 5);
println_str (substr e 0 1);

let rec addfoo s = str_concat s "foo" in
println_str (addfoo "piyo");

let rec add_a s = str_concat s a in
println_str (add_a "poyo");

let rec substr2 s a b = substr s a b in
println_str (substr2 e 2 4);

println_bool (a = b);
println_bool (a <> b);
println_bool (a = a);
println_bool (a <> a);
(* different length string *)
println_bool (a = e);
println_bool (a <> e);

println_str "";
println_str (str_concat "" "");
println_str "foo\tbar\tbaz";
println_str "これは日本語です";
println_str "つらい\t😭";

()
