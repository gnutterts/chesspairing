// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestKeizer1956TwoRoundExample reproduces Keizer's 1956 two-round example
// (KBSB-vademecum (2018), pp. 129-130). The method recalculates all scores
// once after each round using the ranking that stood before that round, uses
// the opponent's value number from that same ranking for all earlier games,
// and adds each player's own value number.
//
// Round 1 values: A=99, B=49, C=71.5, D=71, E=46, F=91, G=65.5, H=65, I=42,
// J=83.
// Round 2 values: A=93, B=87, C=70, D=92, E=62.5, F=141, G=90, H=66.5, I=62,
// J=136.
func TestKeizer1956TwoRoundExample(t *testing.T) {
	method := methodKeizer1956
	base, step := 50, 1
	opts := Options{Method: &method, ValueNumberBase: &base, ValueNumberStep: &step}

	players := []chesspairing.PlayerEntry{
		{ID: "A", DisplayName: "A", Rating: 10},
		{ID: "B", DisplayName: "B", Rating: 9},
		{ID: "C", DisplayName: "C", Rating: 8},
		{ID: "D", DisplayName: "D", Rating: 7},
		{ID: "E", DisplayName: "E", Rating: 6},
		{ID: "F", DisplayName: "F", Rating: 5},
		{ID: "G", DisplayName: "G", Rating: 4},
		{ID: "H", DisplayName: "H", Rating: 3},
		{ID: "I", DisplayName: "I", Rating: 2},
		{ID: "J", DisplayName: "J", Rating: 1},
	}
	rounds := []chesspairing.RoundData{
		{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "C", BlackID: "D", Result: chesspairing.ResultDraw},
				{WhiteID: "E", BlackID: "F", Result: chesspairing.ResultBlackWins},
				{WhiteID: "G", BlackID: "H", Result: chesspairing.ResultDraw},
				{WhiteID: "I", BlackID: "J", Result: chesspairing.ResultBlackWins},
			},
		},
		{
			Number: 2,
			Games: []chesspairing.GameData{
				{WhiteID: "A", BlackID: "F", Result: chesspairing.ResultBlackWins},
				{WhiteID: "J", BlackID: "C", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "D", BlackID: "G", Result: chesspairing.ResultDraw},
				{WhiteID: "H", BlackID: "B", Result: chesspairing.ResultBlackWins},
				{WhiteID: "E", BlackID: "I", Result: chesspairing.ResultDraw},
			},
		},
	}

	roundWants := []map[string]float64{
		{
			"A": 99, "B": 49, "C": 71.5, "D": 71, "E": 46,
			"F": 91, "G": 65.5, "H": 65, "I": 42, "J": 83,
		},
		{
			"A": 93, "B": 87, "C": 70, "D": 92, "E": 62.5,
			"F": 141, "G": 90, "H": 66.5, "I": 62, "J": 136,
		},
	}

	scorer := New(opts)
	state := &chesspairing.TournamentState{Players: players, Rounds: rounds, CurrentRound: len(rounds)}
	for round, want := range roundWants {
		current := *state
		current.Rounds = rounds[:round+1]
		current.CurrentRound = round + 1
		scores, err := scorer.Score(context.Background(), &current)
		if err != nil {
			t.Fatalf("round %d Score: %v", round+1, err)
		}
		got := make(map[string]float64, len(scores))
		for _, s := range scores {
			got[s.PlayerID] = s.Score
		}
		for id, wantScore := range want {
			if got[id] != wantScore {
				t.Errorf("round %d player %s = %v, want %v", round+1, id, got[id], wantScore)
			}
		}
	}
}
