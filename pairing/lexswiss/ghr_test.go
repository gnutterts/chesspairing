// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestPendingGameScore verifies that an adjourned (pending) game awards 0.5
// pairing score to each player in lexswiss (Double Swiss).
func TestPendingGameScore(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultPending},
			},
		}},
		CurrentRound: 2,
	}
	participants, err := BuildParticipantStatesWithPAB(state, 1.5)
	if err != nil {
		t.Fatalf("BuildParticipantStatesWithPAB failed: %v", err)
	}
	byID := make(map[string]ParticipantState)
	for _, p := range participants {
		byID[p.ID] = p
	}
	if byID["p1"].Score != 0.5 {
		t.Errorf("p1 score: want 0.5, got %v", byID["p1"].Score)
	}
	if byID["p2"].Score != 0.5 {
		t.Errorf("p2 score: want 0.5, got %v", byID["p2"].Score)
	}
}
