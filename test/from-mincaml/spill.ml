let rec f a b c d =
  let e = a + b in
  let f = a + c in
  let g = a + d in
  let h = b + c in
  let i = b + d in
  let j = c + d in

  let k = e + f in
  let l = e + g in
  let m = e + h in
  let n = e + i in
  let o = e + j in
  let p = f + g in
  let q = f + h in
  let r = f + i in
  let s = f + j in
  let t = g + h in
  let u = g + i in
  let v = g + j in
  let w = h + i in
  let x = h + j in
  let y = i + j in
    
  let aa = k + l in
  let ab = k + m in
  let ac = k + n in 
  let ad = k + o in
  let ae = k + p in 
  let af = k + q in
  let ag = k + r in 
  let ah = k + s in
  let ai = k + t in
  let aj = k + u in
  let ak = k + v in
  let al = k + w in
  let am = k + x in
  let an = k + y in

  let z = a + b + c + d +
      e + f + g + h + i + j +
      k + l + m + n + o + p + q + r + s + t + u + v + w + x + y +
      aa + ab + ac + ad + ae + af + ag + ah + ai + aj + ak + al + am + an in
  -z in
print_int (f 1 2 3 4)
