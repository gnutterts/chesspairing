// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "testing"

// C.04.3 article 1.2 orders the players by score first and by TPN second, and
// article 5.2.5 gives the initial-colour to the higher ranked player when their
// TPN is odd. Two players without a colour preference who meet in a
// heterogeneous bracket must therefore be ranked by score, not by TPN alone.
func TestAllocateColor_NoPreference_HigherScoreRanksFirst(t *testing.T) {
	// a has the lower TPN (8, even) but the lower score; b has TPN 24 (even).
	// Ranked by TPN alone a would be the higher ranked player; ranked by score
	// b is, and b has an even TPN, so b receives the colour opposite to the
	// initial-colour (Black) and a receives White.
	lowScoreLowTPN := &PlayerState{ID: "a", PairingNumber: 8, PairingScore: 0}
	highScoreHighTPN := &PlayerState{ID: "b", PairingNumber: 24, PairingScore: 0.5}
	white, black := AllocateColor(lowScoreLowTPN, highScoreHighTPN, false, 1, nil, FixedNumberParity)
	if white != "a" || black != "b" {
		t.Errorf("even TPN, higher score: got %s-%s, want a-b", white, black)
	}

	// The same with odd TPNs: the higher ranked player (higher score, TPN 23)
	// receives the initial-colour (White).
	lowScoreLowTPN = &PlayerState{ID: "a", PairingNumber: 15, PairingScore: 0}
	highScoreHighTPN = &PlayerState{ID: "b", PairingNumber: 23, PairingScore: 0.5}
	white, black = AllocateColor(lowScoreLowTPN, highScoreHighTPN, false, 1, nil, FixedNumberParity)
	if white != "b" || black != "a" {
		t.Errorf("odd TPN, higher score: got %s-%s, want b-a", white, black)
	}

	// With equal scores the lower TPN ranks higher, as before.
	x := &PlayerState{ID: "x", PairingNumber: 7, PairingScore: 0.5}
	y := &PlayerState{ID: "y", PairingNumber: 14, PairingScore: 0.5}
	white, black = AllocateColor(y, x, false, 1, nil, FixedNumberParity)
	if white != "x" || black != "y" {
		t.Errorf("equal scores: got %s-%s, want x-y", white, black)
	}
}
