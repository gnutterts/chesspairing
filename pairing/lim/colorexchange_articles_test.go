package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestColourExchange_Art5BCSwapWithTwoPairs(t *testing.T) {
	// Art. 5.2 applies even to two pairs.  The first permitted exchange,
	// b↔c, makes the legal S1-S1 and S2-S2 pairs of Art. 4.2.
	pairs := [][2]*swisslib.PlayerState{
		{
			makePlayerWithColor("a", 1, 2000, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("b", 2, 1900, []swisslib.Color{swisslib.ColorBlack}),
		},
		{
			makePlayerWithColor("c", 3, 1800, []swisslib.Color{swisslib.ColorWhite}),
			makePlayerWithColor("d", 4, 1700, []swisslib.Color{swisslib.ColorWhite}),
		},
	}

	got := ColourExchange(pairs, nil, false)
	assertPairs(t, got, [][2]string{{"a", "c"}, {"b", "d"}})
	if conflicts := conflictCount(got); conflicts != 0 {
		t.Errorf("got %d colour conflicts, want none", conflicts)
	}
}
