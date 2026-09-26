// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// Keizer documentation, "Repeat Avoidance": "If allowRepeatPairings is
// false, any previous encounter is a conflict." With only A and B, there is
// no legal pairing after A-B has been played, so the pairer must fail with an
// error instead of reproducing the repeat.
func TestRepro_KZP02_TwoPlayersAreRepeated(t *testing.T) {
	no := false
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A", Rating: 2}, {ID: "B", Rating: 1}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins}}},
		},
		CurrentRound: 2,
	}
	got, err := New(Options{AllowRepeatPairings: &no}).Pair(context.Background(), state)
	if err == nil {
		t.Fatalf("expected an error for the unavoidable repeat, got pairings=%+v", got.Pairings)
	}
	if err.Error() != "keizer: no pairing satisfies the repeat restrictions" {
		t.Fatalf("error = %q, want %q", err.Error(), "keizer: no pairing satisfies the repeat restrictions")
	}
}

// Keizer documentation, "Repeat Avoidance": "If allowRepeatPairings is
// false, any previous encounter is a conflict." For ranked A,B,C,D with past
// A-B and B-D, the greedy top-down pairing tries A-B and then B-D, but the
// full search must pick the legal A-D and B-C instead.
func TestRepro_KZP02_GreedySwapForcesRepeat(t *testing.T) {
	no := false
	history := buildHistory([]chesspairing.RoundData{
		{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B"}, {WhiteID: "B", BlackID: "D"}}},
	}, false)
	got, err := pairRanked([]string{"A", "B", "C", "D"}, Options{AllowRepeatPairings: &no}, history, nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("input: ranking A,B,C,D; prior pairs A-B and B-D; output: pairings=%+v notes=%v", got.Pairings, got.Notes)
	if len(got.Pairings) != 2 || !hasPair(got, "A", "D") || !hasPair(got, "B", "C") {
		t.Fatalf("expected A-D and B-C, got %+v", got.Pairings)
	}
}
