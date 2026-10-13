// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestKoyaDoubleSwissMaximumUsesTwoWins(t *testing.T) {
	state := &chesspairing.TournamentState{
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDoubleSwiss},
		ScoringConfig: chesspairing.ScoringConfig{Options: map[string]any{"pointWin": 3.0, "pointDraw": 1.0}},
		Players:       []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}},
		Rounds: []chesspairing.RoundData{{Number: 1, Games: []chesspairing.GameData{
			{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "b", BlackID: "a", Result: chesspairing.ResultWhiteWins},
		}}},
	}
	// 2.5 points fall below half of two wins (6), but not half of a win plus a draw (4).
	scores := []chesspairing.PlayerScore{{PlayerID: "a", Score: 2.5}, {PlayerID: "b", Score: 2.5}}
	values, err := (&Koya{}).Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range values {
		if value.Value != 0 {
			t.Errorf("%s Koya = %g, want 0", value.PlayerID, value.Value)
		}
	}
}

func TestKoyaLimit60(t *testing.T) {
	state := &chesspairing.TournamentState{
		ScoringConfig: chesspairing.ScoringConfig{Options: map[string]any{"koyaLimit": 60.0}},
		Players:       []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}},
		Rounds:        []chesspairing.RoundData{{Number: 1, Games: []chesspairing.GameData{{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultDraw}}}},
	}
	scores := []chesspairing.PlayerScore{{PlayerID: "a", Score: 0.5}, {PlayerID: "b", Score: 0.5}}
	values, err := (&Koya{}).Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range values {
		if value.Value != 0 {
			t.Errorf("%s Koya at 60%% = %g, want 0", value.PlayerID, value.Value)
		}
	}
	state.ScoringConfig.Options = nil
	values, err = (&Koya{}).Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	if values[0].Value != 0.5 || values[1].Value != 0.5 {
		t.Errorf("default Koya at 50%% = %v, want two draws", values)
	}
}
