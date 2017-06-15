let rec foo a b c d e f =
  print_int a;
  print_int b;
  print_int c;
  print_int d;
  print_int e;
  print_int f in
let rec bar a b c d e f =
  foo b a d e f c in
bar 1 2 3 4 5 6
