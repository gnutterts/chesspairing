// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestPairGamePrimaryUsesMatchSecondary(t *testing.T) {
	state := &cp.TournamentState{
		Players:      []cp.PlayerEntry{{ID: "x", PairingNumber: 1}, {ID: "y", PairingNumber: 2}, {ID: "c", PairingNumber: 3}, {ID: "d", PairingNumber: 4}},
		CurrentRound: 3,
	}
	for i, scores := range []struct{ x, y []cp.GameResult }{
		{[]cp.GameResult{cp.ResultWhiteWins, cp.ResultWhiteWins, cp.ResultBlackWins}, []cp.GameResult{cp.ResultWhiteWins, cp.ResultWhiteWins, cp.ResultBlackWins}},
		{[]cp.GameResult{cp.ResultWhiteWins, cp.ResultBlackWins, cp.ResultBlackWins}, []cp.GameResult{cp.ResultDraw, cp.ResultDraw}},
	} {
		round := cp.RoundData{Number: i + 1}
		for _, result := range scores.x {
			round.Games = append(round.Games, cp.GameData{WhiteID: "x", BlackID: "c", Result: result})
		}
		for _, result := range scores.y {
			round.Games = append(round.Games, cp.GameData{WhiteID: "y", BlackID: "d", Result: result})
		}
		state.Rounds = append(state.Rounds, round)
	}
	for _, tc := range []struct {
		name      string
		secondary bool
		white     string
	}{
		{"match secondary", true, "x"}, {"disabled", false, "y"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := New(Options{PrimaryScore: stringPtr("game"), SecondaryScoreForColour: &tc.secondary, ForbiddenPairs: [][]string{{"x", "d"}, {"y", "c"}}}).Pair(context.Background(), state)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, pair := range result.Pairings {
				if pair.WhiteID == "x" && pair.BlackID == "y" || pair.WhiteID == "y" && pair.BlackID == "x" {
					found = true
					if pair.WhiteID != tc.white {
						t.Errorf("white = %s, want %s", pair.WhiteID, tc.white)
					}
				}
			}
			if !found {
				t.Errorf("x and y not paired: %v", result.Pairings)
			}
		})
	}
}
