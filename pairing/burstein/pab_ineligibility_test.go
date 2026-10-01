// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"context"
	"errors"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// TestPABIneligibilityForfeitWinner covers FIDE C.04.4.2 C2. The old
// implementation gave p5 a PAB because it only excluded prior PAB recipients.
func TestPABIneligibilityForfeitWinner(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}, {ID: "p3"}, {ID: "p4"}, {ID: "p5"}},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
					{WhiteID: "p5", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
					{WhiteID: "p3", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeZero}},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
					{WhiteID: "p1", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByeZero}},
			},
		},
		CurrentRound: 4,
	}
	_, err := New(Options{}).Pair(context.Background(), state)
	if !errors.Is(err, swisslib.ErrNoPABCandidate) {
		t.Fatalf("Pair() error = %v, want ErrNoPABCandidate", err)
	}
}

func TestByeFullPointContributesToOppositionIndex(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Byes:   []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeFullPoint}},
		}},
	}
	player := &swisslib.PlayerState{ID: "p1", Score: 1}
	index := ComputeOppositionIndex(player, state)
	if index.SonnebornBerger != 1 {
		t.Errorf("SonnebornBerger = %v, want 1", index.SonnebornBerger)
	}
	if got := computePairingScores(state)["p1"]; got != 1 {
		t.Errorf("pairing score = %v, want 1", got)
	}
}
