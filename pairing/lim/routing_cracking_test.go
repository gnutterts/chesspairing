// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"context"
	"errors"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPair_BlockedMedianCrackingReturnsImpossible(t *testing.T) {
	// The 0.5 scoregroup is the median after round 1. Its only two new
	// pairings are forbidden, and there are no adjacent pairs to crack under
	// Art. 2.6, so the required result is PairingImpossible.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", Rating: 2400},
			{ID: "p2", Rating: 2300},
			{ID: "p3", Rating: 2200},
			{ID: "p4", Rating: 2100},
		},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},
				{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultDraw},
			},
		}},
		CurrentRound: 2,
	}

	forbidden := [][]string{
		{"p1", "p2"},
		{"p1", "p4"},
		{"p2", "p3"},
		{"p3", "p4"},
	}
	result, err := New(Options{ForbiddenPairs: forbidden}).Pair(context.Background(), state)
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	var pairingErr *chesspairing.PairingError
	if !errors.As(err, &pairingErr) || pairingErr.Kind != chesspairing.PairingImpossible {
		t.Fatalf("error = %#v, want PairingImpossible", err)
	}
}
