// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestAllocateColor_Round1_OddBoard(t *testing.T) {
	a := &lexswiss.ParticipantState{ID: "a", TPN: 1}
	b := &lexswiss.ParticipantState{ID: "b", TPN: 2}

	// Round 1, board 1 (odd): higher ranked (a) gets White.
	wID, bID := AllocateColor(a, b, 1, 1, nil)
	if wID != "a" || bID != "b" {
		t.Errorf("odd board: expected a=White, b=Black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_Round1_EvenBoard(t *testing.T) {
	a := &lexswiss.ParticipantState{ID: "a", TPN: 1}
	b := &lexswiss.ParticipantState{ID: "b", TPN: 2}

	// Round 1, board 2 (even): higher ranked (a) gets Black.
	wID, bID := AllocateColor(a, b, 1, 2, nil)
	if wID != "b" || bID != "a" {
		t.Errorf("even board: expected b=White, a=Black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_Round1_TopSeedBlack(t *testing.T) {
	a := &lexswiss.ParticipantState{ID: "a", TPN: 1}
	b := &lexswiss.ParticipantState{ID: "b", TPN: 2}
	black := "black"

	// Round 1, board 1 (odd) with topSeedColor=black: higher ranked gets Black.
	wID, bID := AllocateColor(a, b, 1, 1, &black)
	if wID != "b" || bID != "a" {
		t.Errorf("topSeedColor=black, odd board: expected b=White, a=Black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_Equalise(t *testing.T) {
	W := lexswiss.ColorWhite
	B := lexswiss.ColorBlack
	// Player a had 2 Whites, 1 Black. Player b had 1 White, 2 Blacks.
	// Equalise: a should get Black (fewer Blacks), b should get White.
	a := &lexswiss.ParticipantState{ID: "a", TPN: 1, ColorHistory: []lexswiss.Color{W, W, B}}
	b := &lexswiss.ParticipantState{ID: "b", TPN: 2, ColorHistory: []lexswiss.Color{B, B, W}}

	wID, bID := AllocateColor(a, b, 4, 1, nil)
	if wID != "b" || bID != "a" {
		t.Errorf("equalise: expected b=White, a=Black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_Alternate(t *testing.T) {
	W := lexswiss.ColorWhite
	B := lexswiss.ColorBlack
	// Both have equal white-game counts (1 each).
	// Player a had White last, player b had Black last → alternate.
	a := &lexswiss.ParticipantState{ID: "a", TPN: 1, ColorHistory: []lexswiss.Color{W}}
	b := &lexswiss.ParticipantState{ID: "b", TPN: 2, ColorHistory: []lexswiss.Color{B}}

	wID, bID := AllocateColor(a, b, 2, 1, nil)
	// a had White → gets Black now. b had Black → gets White now.
	if wID != "b" || bID != "a" {
		t.Errorf("alternate: expected b=White, a=Black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_TiebreakByRank(t *testing.T) {
	// Both have identical colour history → higher ranked (lower TPN) gets White.
	a := &lexswiss.ParticipantState{ID: "a", TPN: 1, ColorHistory: nil}
	b := &lexswiss.ParticipantState{ID: "b", TPN: 2, ColorHistory: nil}

	// Round 2 with no history → fall through to rank tiebreak.
	// But round 1 uses board alternation, so test with round 2.
	// Actually with no history at all, round 2 should use rank tiebreak.
	wID, bID := AllocateColor(a, b, 2, 1, nil)
	if wID != "a" || bID != "b" {
		t.Errorf("rank tiebreak: expected a=White (lower TPN), got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_NoThreeConsecutive(t *testing.T) {
	W := lexswiss.ColorWhite
	// Player a had White in Game 1 for 2 consecutive rounds. Must NOT get White again.
	a := &lexswiss.ParticipantState{ID: "a", TPN: 1, ColorHistory: []lexswiss.Color{W, W}}
	b := &lexswiss.ParticipantState{ID: "b", TPN: 2, ColorHistory: nil}

	// Despite being higher ranked, a should get Black (3-consecutive constraint).
	wID, bID := AllocateColor(a, b, 3, 1, nil)
	if wID != "b" || bID != "a" {
		t.Errorf("3-consecutive: expected b=White, a=Black; got w=%s, b=%s", wID, bID)
	}
}
