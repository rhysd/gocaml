package gcil

type Visitor interface {
	Visit(i *Insn) Visitor
}

func Visit(v Visitor, b *Block) bool {
	for i := b.Top; i != nil; i = i.Next {
		if v = v.Visit(i); v == nil {
			return true
		}
		switch val := i.Val.(type) {
		case *If:
			if Visit(v, val.Then) || Visit(v, val.Else) {
				return true
			}
		case *Fun:
			if Visit(v, val.Body) {
				return true
			}
		}
	}
	return false
}
