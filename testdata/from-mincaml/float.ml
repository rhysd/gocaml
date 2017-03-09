(* このテストを実行する場合は、Main.file等を呼び出す前に
   Typing.extenvを:=等で書き換えて、あらかじめsinやcosなど
   外部関数の型を陽に指定する必要があります（そうしないと
   MinCamlでは勝手にint -> intと推論されるため）。 *)
print_int
  ((float_to_int
     (((sin ((cos ((sqrt ((abs_float (-.12.3)) +. 0.0)) +. 0.0)) +. 0.0)
	 +. 4.5 -. 6.7 *. 8.9 /. 1.23456789) +. 0.0)
	*. int_to_float 1000000)) + 0)
