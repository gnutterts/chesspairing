// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPairTopSeedColorSynonyms(t *testing.T) {
	for _, topSeedColor := range []string{"black", "B", "black1"} {
		t.Run(topSeedColor, func(t *testing.T) {
			state := &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "1", Rating: 2000},
					{ID: "2", Rating: 1500},
				},
				CurrentRound: 1,
			}
			result, err := New(Options{TopSeedColor: &topSeedColor}).Pair(context.Background(), state)
			if err != nil {
				t.Fatalf("Pair: %v", err)
			}
			if len(result.Pairings) != 1 {
				t.Fatalf("pairings = %d, want 1", len(result.Pairings))
			}
			pairing := result.Pairings[0]
			if pairing.WhiteID != "2" || pairing.BlackID != "1" {
				t.Errorf("board 1 = %s-%s, want 2-1", pairing.WhiteID, pairing.BlackID)
			}
		})
	}
}

func TestPairInvalidTopSeedColor(t *testing.T) {
	invalid := "green"
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "1", Rating: 2000},
			{ID: "2", Rating: 1500},
		},
		CurrentRound: 1,
	}

	_, err := New(Options{TopSeedColor: &invalid}).Pair(context.Background(), state)
	if err == nil {
		t.Fatal("Pair returned nil error")
	}
	if !strings.Contains(err.Error(), invalid) {
		t.Errorf("error = %q, want invalid value %q", err, invalid)
	}
}
