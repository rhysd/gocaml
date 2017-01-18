let rec mul l m n a b c =
  let rec loop1 i =
    if i < 0 then () else
    let rec loop2 j =
      if j < 0 then () else
      let rec loop3 k =
	if k < 0 then () else
	(c.(i).(j) <- c.(i).(j) +. a.(i).(k) *. b.(k).(j);
	 loop3 (k - 1)) in
      loop3 (m - 1);
      loop2 (j - 1) in
    loop2 (n - 1);
    loop1 (i - 1) in
  loop1 (l - 1) in
let dummy = Array.make 0 0. in
let rec make m n =
  let mat = Array.make m dummy in
  let rec init i =
    if i < 0 then () else
    (mat.(i) <- Array.make n 0.;
     init (i - 1)) in
  init (m - 1);
  mat in
let a = make 2 3 in
let b = make 3 2 in
let c = make 2 2 in
a.(0).(0) <- 1.; a.(0).(1) <- 2.; a.(0).(2) <- 3.;
a.(1).(0) <- 4.; a.(1).(1) <- 5.; a.(1).(2) <- 6.;
b.(0).(0) <- 7.; b.(0).(1) <- 8.;
b.(1).(0) <- 9.; b.(1).(1) <- 10.;
b.(2).(0) <- 11.; b.(2).(1) <- 12.;
mul 2 3 2 a b c;
print_int (truncate (c.(0).(0)));
print_newline ();
print_int (truncate (c.(0).(1)));
print_newline ();
print_int (truncate (c.(1).(0)));
print_newline ();
print_int (truncate (c.(1).(1)));
print_newline ()
