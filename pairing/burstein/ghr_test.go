// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestForfeitUnionInBurstein verifies Burstein state respects forfeit union:
// both a forfeit result with IsForfeit: false and the TRF-W shape (plain win
// with IsForfeit: true) leave the encounter unplayed, so the rematch is
// allowed. PairingConfig.System is set to PairingBurstein, as the rematch
// validation (C.04.2 3.5) only forbids repeats for a configured system.
func TestForfeitUnionInBurstein(t *testing.T) {
	tests := []struct {
		name string
		game chesspairing.GameData
	}{
		{
			name: "forfeit result without the flag",
			game: chesspairing.GameData{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: false},
		},
		{
			name: "flag with a plain win (TRF W)",
			game: chesspairing.GameData{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins, IsForfeit: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}},
				Rounds: []chesspairing.RoundData{{
					Number: 1,
					Games:  []chesspairing.GameData{tt.game},
				}},
				CurrentRound:  2,
				PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingBurstein},
			}
			pairer := New(Options{})
			res, err := pairer.Pair(context.Background(), state)
			if err != nil {
				t.Fatalf("Pair failed: %v", err)
			}
			// Rematch allowed between p1 and p2
			if len(res.Pairings) != 1 {
				t.Fatalf("expected 1 pairing (rematch allowed), got %d", len(res.Pairings))
			}
			if (res.Pairings[0].WhiteID != "p1" || res.Pairings[0].BlackID != "p2") &&
				(res.Pairings[0].WhiteID != "p2" || res.Pairings[0].BlackID != "p1") {
				t.Errorf("expected the p1/p2 rematch, got %v", res.Pairings[0])
			}
		})
	}
}

func TestPreAssignedPAB_C2Validation(t *testing.T) {
	pairer := New(Options{})

	stateSecondPAB := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "B", BlackID: "C", Result: chesspairing.ResultDraw},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "A", Type: chesspairing.ByePAB}},
		}},
		CurrentRound: 2,
		PreAssignedByes: []chesspairing.ByeEntry{
			{PlayerID: "A", Type: chesspairing.ByePAB},
		},
	}
	_, err := pairer.Pair(context.Background(), stateSecondPAB)
	if err == nil {
		t.Fatal("expected error for second PAB on player A, got nil")
	}
	if !strings.Contains(err.Error(), "A") || !strings.Contains(err.Error(), "C2") {
		t.Errorf("error should name player A and C2, got %v", err)
	}
}
