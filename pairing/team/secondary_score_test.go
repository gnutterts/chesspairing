// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestPairSecondaryScoreForColour(t *testing.T) {
	state := &cp.TournamentState{
		Players: []cp.PlayerEntry{{ID: "y", PairingNumber: 1}, {ID: "x", PairingNumber: 2}, {ID: "c", PairingNumber: 3}, {ID: "d", PairingNumber: 4}},
		Rounds: []cp.RoundData{{
			Number: 1,
			Games: []cp.GameData{
				{WhiteID: "x", BlackID: "c", Result: cp.ResultWhiteWins},
				{WhiteID: "x", BlackID: "c", Result: cp.ResultWhiteWins},
				{WhiteID: "x", BlackID: "c", Result: cp.ResultWhiteWins},
				{WhiteID: "x", BlackID: "c", Result: cp.ResultBlackWins},
				{WhiteID: "y", BlackID: "d", Result: cp.ResultWhiteWins},
				{WhiteID: "y", BlackID: "d", Result: cp.ResultDraw},
				{WhiteID: "y", BlackID: "d", Result: cp.ResultWhiteWins},
				{WhiteID: "y", BlackID: "d", Result: cp.ResultBlackWins},
			},
		}},
		CurrentRound: 2,
	}
	for _, tc := range []struct {
		name      string
		secondary bool
		white     string
	}{
		{"default game points", true, "y"}, {"disabled", false, "x"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := New(Options{SecondaryScoreForColour: &tc.secondary, ForbiddenPairs: [][]string{{"x", "d"}, {"y", "c"}}}).Pair(context.Background(), state)
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
