package gcil

type Visitor interface {
	Visit(i *Insn) Visitor
}

func Visit(v Visitor, b *Block) bool {
	// Note:
	// Skip first and last instructions because they are NOP.
	// Basic block has the NOPs as cushion to modify instruction
	// sequence easily.
	for i := b.Top.Next; i.Next != nil; i = i.Next {
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
