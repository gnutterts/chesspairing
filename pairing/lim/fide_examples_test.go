// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// TestFIDEExample_lim_1 verifies the worked example in FIDE C.04.4.3
// (Lim System, effective from 1 February 2026), Article 4.2.
//
// Source text: "Six players in a scoregroup with proposed pairings as
// follows: 1 v 4, 2 v 5, 3 v 6. If the pairing 1 v 4 is not compatible, ...
// The original proposed pairing and possible exchanges made to find a
// compatible opponent for #1 are as follows:
//
//	1 v 4  1 v 5  1 v 6  1 v 3  1 v 2"
//
// The exchange sequence is rebuilt with the package-internal helper
// generateExchangeOrder (same package), pairing downward.
func TestFIDEExample_lim_1(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "1", TPN: 1},
		{ID: "2", TPN: 2},
		{ID: "3", TPN: 3},
		{ID: "4", TPN: 4},
		{ID: "5", TPN: 5},
		{ID: "6", TPN: 6},
	}

	// Transcription guard: six players numbered 1..6 in ascending TPN order.
	if len(players) != 6 {
		t.Fatalf("input must contain exactly 6 players, got %d", len(players))
	}
	for i, p := range players {
		if p.TPN != i+1 {
			t.Fatalf("input players must be in ascending TPN order, got TPN %d at position %d", p.TPN, i)
		}
	}

	const half = 3
	order := generateExchangeOrder(0, half, true)

	// Map the unified indices ([S1 | S2]) back to the FIDE player numbers.
	onsParts := make([]string, len(order))
	for i, idx := range order {
		onsParts[i] = players[idx].ID
	}
	got := strings.Join(onsParts, " ")
	fide := "4 5 6 3 2"
	if got != fide {
		t.Errorf("lim-4.2: exchange order = %s, want FIDE %s", got, fide)
	}
}

// TestFIDEExample_lim_2 verifies the worked example in FIDE C.04.4.3
// (Lim System, effective from 1 February 2026), Article 7.2.
//
// Source text: "Depending on the draw, the pairings for the first round in a
// tournament of forty players would be either 1 v 21, 22 v 2, 3 v 23, 24 v 4,
// ... 40 v 20; or 21 v 1, 2 v 22, 23 v 3, 4 v 24 ... 20 v 40, where the
// player having White is mentioned first."
//
// The example is rebuilt with the public Pair API for round 1. The code's
// default round-1 colour for the top seed is White, which corresponds to the
// first branch ("1 v 21, 22 v 2, ...") in the source.
func TestFIDEExample_lim_2(t *testing.T) {
	players := make([]chesspairing.PlayerEntry, 40)
	for i := range players {
		n := i + 1
		players[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("p%02d", n),
			DisplayName: fmt.Sprintf("P%02d", n),
			Rating:      4000 - i, // descending ratings -> initial ranks 1..40
		}
	}

	// Transcription guard: forty players, ratings strictly descending.
	if len(players) != 40 {
		t.Fatalf("input must contain exactly 40 players, got %d", len(players))
	}
	for i := 1; i < len(players); i++ {
		if players[i].Rating >= players[i-1].Rating {
			t.Fatalf("ratings must be strictly descending, got %d then %d",
				players[i-1].Rating, players[i].Rating)
		}
	}

	pairer := New(Options{})
	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() round 1 error: %v", err)
	}

	var onsParts []string
	for _, gp := range result.Pairings {
		onsParts = append(onsParts, gp.WhiteID+"-"+gp.BlackID)
	}
	got := strings.Join(onsParts, " ")

	// FIDE expected, White-first branch (top seed White), expanded from
	// "1 v 21, 22 v 2, 3 v 23, 24 v 4, ... 40 v 20".
	var fideParts []string
	for i := 1; i <= 20; i++ {
		if i%2 == 1 {
			fideParts = append(fideParts, fmt.Sprintf("p%02d-p%02d", i, i+20))
		} else {
			fideParts = append(fideParts, fmt.Sprintf("p%02d-p%02d", i+20, i))
		}
	}
	fide := strings.Join(fideParts, " ")

	if got != fide {
		t.Errorf("lim-7.2: pairings = %s, want FIDE %s", got, fide)
	}
}
