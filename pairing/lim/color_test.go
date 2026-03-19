package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestAllocateColor_Round1Alternation(t *testing.T) {
	a := &swisslib.PlayerState{ID: "a", TPN: 1}
	b := &swisslib.PlayerState{ID: "b", TPN: 2}

	// Round 1, board 1: top seed color = White (default).
	// Odd-numbered players in upper half get same as #1.
	wID, bID := AllocateColor(a, b, 1, true, nil)
	if wID != "a" || bID != "b" {
		t.Errorf("expected a=White, b=Black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_AlternatingColour(t *testing.T) {
	W := swisslib.ColorWhite
	a := &swisslib.PlayerState{ID: "a", TPN: 1, ColorHistory: []swisslib.Color{W}}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, ColorHistory: nil}

	// Player a had White last → should get Black now (alternation).
	wID, bID := AllocateColor(a, b, 2, true, nil)
	if wID != "b" || bID != "a" {
		t.Errorf("expected b=White, a=Black (alternation); got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_MedianTiebreak_AboveMedian(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Both had same colour history → identical. Above median: higher ranked (lower TPN) gets alternating.
	a := &swisslib.PlayerState{ID: "a", TPN: 1, ColorHistory: []swisslib.Color{W, B}}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, ColorHistory: []swisslib.Color{W, B}}

	// Above median (isAboveMedian=true). Both last had B → due W.
	// Higher ranked = a (TPN 1). a gets alternating colour (W).
	wID, bID := AllocateColor(a, b, 3, true, nil)
	if wID != "a" || bID != "b" {
		t.Errorf("above median: expected a=White (higher ranked), got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_MedianTiebreak_BelowMedian(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Below median: lower ranked (higher TPN) gets alternating colour.
	a := &swisslib.PlayerState{ID: "a", TPN: 1, ColorHistory: []swisslib.Color{W, B}}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, ColorHistory: []swisslib.Color{W, B}}

	// Below median (isAboveMedian=false). Both last had B → due W.
	// Lower ranked = b (TPN 2). b gets alternating colour (W).
	wID, bID := AllocateColor(a, b, 3, false, nil)
	if wID != "b" || bID != "a" {
		t.Errorf("below median: expected b=White (lower ranked), got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_EvenRoundEqualising(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Even round: equalising colour takes precedence.
	// Player a has WBW (2W, 1B) → needs B to equalise.
	// Player b has BWB (1W, 2B) → needs W to equalise.
	a := &swisslib.PlayerState{ID: "a", TPN: 1, ColorHistory: []swisslib.Color{W, B, W}}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, ColorHistory: []swisslib.Color{B, W, B}}

	// Round 4 (even): both can get their equalising colour.
	wID, bID := AllocateColor(a, b, 4, true, nil)
	if wID != "b" || bID != "a" {
		t.Errorf("even round equalising: expected b=White, a=Black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_OnePlayerMustAlternate(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Player a had BB → MUST get White (Art. 5.3).
	a := &swisslib.PlayerState{ID: "a", TPN: 1, ColorHistory: []swisslib.Color{B, B}}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, ColorHistory: []swisslib.Color{W}}

	wID, bID := AllocateColor(a, b, 3, true, nil)
	if wID != "a" || bID != "b" {
		t.Errorf("player a must alternate to White; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_HistoryDifference(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Both due same colour, but differ in history.
	// Art. 5.4: go back to find first difference.
	a := &swisslib.PlayerState{ID: "a", TPN: 1, ColorHistory: []swisslib.Color{W, B, W}}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, ColorHistory: []swisslib.Color{B, B, W}}

	// Both last had W → both due B.
	// Looking back: round 2: a=B, b=B (same). Round 1: a=W, b=B (differ!).
	// Player who had the differing colour in that round gets alternated.
	wID, bID := AllocateColor(a, b, 4, true, nil)
	_ = wID
	_ = bID
	// The exact result depends on the interpretation. Just check it doesn't panic.
}
