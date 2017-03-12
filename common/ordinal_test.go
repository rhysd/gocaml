package common

import (
	"fmt"
	"testing"
)

func TestOrdinal(t *testing.T) {
	for _, tc := range []struct {
		input  int
		suffix string
	}{
		{1, "st"},
		{2, "nd"},
		{3, "rd"},
		{4, "th"},
		{11, "th"},
		{13, "th"},
		{20, "th"},
		{21, "st"},
		{33, "rd"},
		{100, "th"},
		{101, "st"},
		{102, "nd"},
		{111, "th"},
		{112, "th"},
		{141, "st"},
	} {
		want := fmt.Sprintf("%d%s", tc.input, tc.suffix)
		had := Ordinal(tc.input)
		if want != had {
			t.Errorf("Ordinal(%d) == %s (want %s)", tc.input, had, want)
		}
	}
}
