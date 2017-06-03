package token

import (
	"github.com/rhysd/loc"
	"testing"
)

func TestTokenString(t *testing.T) {
	s := loc.NewDummySource("abcd")
	tok := Token{
		Kind:  IDENT,
		Start: loc.Pos{1, 1, 2, s},
		End:   loc.Pos{3, 1, 4, s},
		File:  s,
	}
	actual := tok.String()
	expected := "<IDENT:bc>(1:2:1-1:4:3)"
	if actual != expected {
		t.Fatalf("Expected '%s' but actually '%s'", expected, actual)
	}
}
