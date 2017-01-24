package token

import (
	"testing"
)

func TestTokenString(t *testing.T) {
	s := &Source{
		Name: "tmp",
		Code: []byte("abcd"),
	}
	tok := Token{
		Kind:  IDENT,
		Start: Position{1, 1, 2},
		End:   Position{3, 1, 4},
		File:  s,
	}
	actual := tok.String()
	expected := "<IDENT:bc>(1:2:1-1:4:3)"
	if actual != expected {
		t.Fatalf("Expected '%s' but actually '%s'", expected, actual)
	}
}
