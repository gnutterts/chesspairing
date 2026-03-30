package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// conflictCount returns the total number of colour conflicts across all pairs.
// Test-only helper — the production code uses conflictCount2 for a pair of pairs.
func conflictCount(pairs [][2]*swisslib.PlayerState) int {
	n := 0
	for _, p := range pairs {
		if hasColourConflict(p[0], p[1]) {
			n++
		}
	}
	return n
}

func makePlayerWithColor(id string, tpn int, rating int, colors []swisslib.Color) *swisslib.PlayerState {
	return &swisslib.PlayerState{
		ID:           id,
		TPN:          tpn,
		Rating:       rating,
		Active:       true,
		ColorHistory: colors,
	}
}

func TestColourExchange_NoConflicts(t *testing.T) {
	// Three pairs where colour preferences are compatible — no swaps needed.
	pairs := [][2]*swisslib.PlayerState{
		{
			makePlayerWithColor("p1", 1, 2000, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p2", 2, 1900, []swisslib.Color{swisslib.ColorWhite}),
		},
		{
			makePlayerWithColor("p3", 3, 1800, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p4", 4, 1700, []swisslib.Color{swisslib.ColorWhite}),
		},
		{
			makePlayerWithColor("p5", 5, 1600, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p6", 6, 1500, []swisslib.Color{swisslib.ColorWhite}),
		},
	}

	result := ColourExchange(pairs, nil, false)

	if len(result) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(result))
	}
	// Pairs should be unchanged.
	if result[0][0].ID != "p1" || result[0][1].ID != "p2" {
		t.Errorf("pair 0 changed: got %s vs %s", result[0][0].ID, result[0][1].ID)
	}
	if result[1][0].ID != "p3" || result[1][1].ID != "p4" {
		t.Errorf("pair 1 changed: got %s vs %s", result[1][0].ID, result[1][1].ID)
	}
	if result[2][0].ID != "p5" || result[2][1].ID != "p6" {
		t.Errorf("pair 2 changed: got %s vs %s", result[2][0].ID, result[2][1].ID)
	}
}

func TestColourExchange_SwapResolvesConflict(t *testing.T) {
	// Pair 0: p1 (due White) vs p2 (due White) → conflict.
	// Pair 1: p3 (due Black) vs p4 (due Black) → conflict.
	// Pair 2: bystander, no conflict.
	// Swapping p2↔p3: {p1, p3} (W vs B) and {p2, p4} (W vs B) → resolves both.
	pairs := [][2]*swisslib.PlayerState{
		{
			makePlayerWithColor("p1", 1, 2000, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p2", 2, 1900, []swisslib.Color{swisslib.ColorBlack}),
		},
		{
			makePlayerWithColor("p3", 3, 1800, []swisslib.Color{swisslib.ColorWhite}),
			makePlayerWithColor("p4", 4, 1700, []swisslib.Color{swisslib.ColorWhite}),
		},
		{
			makePlayerWithColor("p5", 5, 1600, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p6", 6, 1500, []swisslib.Color{swisslib.ColorWhite}),
		},
	}

	result := ColourExchange(pairs, nil, false)

	if len(result) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(result))
	}
	if c := conflictCount(result); c != 0 {
		t.Errorf("expected 0 conflicts after swap, got %d", c)
	}
	// Bystander pair should be unchanged.
	if result[2][0].ID != "p5" || result[2][1].ID != "p6" {
		t.Errorf("bystander pair changed: got %s vs %s", result[2][0].ID, result[2][1].ID)
	}
}

func TestColourExchange_MaxiBlocksSwap(t *testing.T) {
	// Three pairs. Pairs 0 and 1 have colour conflicts.
	// In maxi mode, all swappable player pairs across the three pairs have
	// ratings differing by >100 → all swaps blocked.
	pairs := [][2]*swisslib.PlayerState{
		{
			makePlayerWithColor("p1", 1, 2400, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p2", 2, 2300, []swisslib.Color{swisslib.ColorBlack}),
		},
		{
			makePlayerWithColor("p3", 3, 2000, []swisslib.Color{swisslib.ColorWhite}),
			makePlayerWithColor("p4", 4, 1900, []swisslib.Color{swisslib.ColorWhite}),
		},
		{
			makePlayerWithColor("p5", 5, 1500, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p6", 6, 1400, []swisslib.Color{swisslib.ColorWhite}),
		},
	}

	result := ColourExchange(pairs, nil, true)

	// All swaps blocked by maxi 100-point constraint.
	if c := conflictCount(result); c != 2 {
		t.Errorf("expected 2 conflicts (swap blocked by maxi), got %d", c)
	}
}

func TestColourExchange_NonMaxiAllowsSwap(t *testing.T) {
	// Same scenario as MaxiBlocksSwap but with isMaxi=false.
	// Ratings differ by >100 but that doesn't matter in non-maxi mode.
	pairs := [][2]*swisslib.PlayerState{
		{
			makePlayerWithColor("p1", 1, 2400, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p2", 2, 2300, []swisslib.Color{swisslib.ColorBlack}),
		},
		{
			makePlayerWithColor("p3", 3, 2000, []swisslib.Color{swisslib.ColorWhite}),
			makePlayerWithColor("p4", 4, 1900, []swisslib.Color{swisslib.ColorWhite}),
		},
		{
			makePlayerWithColor("p5", 5, 1500, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p6", 6, 1400, []swisslib.Color{swisslib.ColorWhite}),
		},
	}

	result := ColourExchange(pairs, nil, false)

	if c := conflictCount(result); c != 0 {
		t.Errorf("expected 0 conflicts (non-maxi allows swap), got %d", c)
	}
}

func TestColourExchange_IncompatibleSwapSkipped(t *testing.T) {
	// Three pairs, all with colour conflicts.
	// Every possible swap between any two pairs creates a rematch → all skipped.
	//
	// Setup: each player has already played every other player except their
	// current partner. This means ANY swap would create a rematch.
	pairs := [][2]*swisslib.PlayerState{
		{
			makePlayerWithColor("p1", 1, 2000, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p2", 2, 1900, []swisslib.Color{swisslib.ColorBlack}),
		},
		{
			makePlayerWithColor("p3", 3, 1800, []swisslib.Color{swisslib.ColorWhite}),
			makePlayerWithColor("p4", 4, 1700, []swisslib.Color{swisslib.ColorWhite}),
		},
		{
			makePlayerWithColor("p5", 5, 1600, []swisslib.Color{swisslib.ColorBlack}),
			makePlayerWithColor("p6", 6, 1500, []swisslib.Color{swisslib.ColorBlack}),
		},
	}
	// p1 played everyone except p2.
	pairs[0][0].Opponents = []string{"p3", "p4", "p5", "p6"}
	// p2 played everyone except p1.
	pairs[0][1].Opponents = []string{"p3", "p4", "p5", "p6"}
	// p3 played everyone except p4.
	pairs[1][0].Opponents = []string{"p1", "p2", "p5", "p6"}
	// p4 played everyone except p3.
	pairs[1][1].Opponents = []string{"p1", "p2", "p5", "p6"}
	// p5 played everyone except p6.
	pairs[2][0].Opponents = []string{"p1", "p2", "p3", "p4"}
	// p6 played everyone except p5.
	pairs[2][1].Opponents = []string{"p1", "p2", "p3", "p4"}

	result := ColourExchange(pairs, nil, false)

	// All swaps create rematches → conflicts remain.
	if c := conflictCount(result); c != 3 {
		t.Errorf("expected 3 conflicts (all swaps blocked), got %d", c)
	}
	// Pairs should be unchanged.
	if result[0][0].ID != "p1" || result[0][1].ID != "p2" {
		t.Errorf("pair 0 changed: got %s vs %s", result[0][0].ID, result[0][1].ID)
	}
	if result[1][0].ID != "p3" || result[1][1].ID != "p4" {
		t.Errorf("pair 1 changed: got %s vs %s", result[1][0].ID, result[1][1].ID)
	}
	if result[2][0].ID != "p5" || result[2][1].ID != "p6" {
		t.Errorf("pair 2 changed: got %s vs %s", result[2][0].ID, result[2][1].ID)
	}
}
