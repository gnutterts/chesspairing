// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestTeamTieBreakers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Matches: []chesspairing.MatchData{{
					HomeID: "A",
					AwayID: "B",
					Boards: []chesspairing.GameData{
						{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
						{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultWhiteWins},
					},
				}},
				TeamByes: []chesspairing.ByeEntry{{PlayerID: "C", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Matches: []chesspairing.MatchData{{
					HomeID: "B",
					AwayID: "C",
					Result: &chesspairing.TeamMatchResult{HomeGame: 2, AwayGame: 0},
				}},
				TeamByes: []chesspairing.ByeEntry{{PlayerID: "A", Type: chesspairing.ByePAB}},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 2, Team: &chesspairing.TeamPoints{Match: 2, Game: 2}},
		{PlayerID: "B", Score: 3, Team: &chesspairing.TeamPoints{Match: 3, Game: 3}},
		{PlayerID: "C", Score: 1, Team: &chesspairing.TeamPoints{Match: 1, Game: 1}},
	}
	wants := map[string]map[string]float64{
		"mpvgp":                    {"A": 2, "B": 3, "C": 1},
		"emmsb":                    {"A": 5, "B": 4, "C": 1},
		"emgsb":                    {"A": 5, "B": 4, "C": 1},
		"egmsb":                    {"A": 5, "B": 4, "C": 1},
		"eggsb":                    {"A": 5, "B": 4, "C": 1},
		"buchholz-mp":              {"A": 5, "B": 3, "C": 4},
		"buchholz-mp-cut1":         {"A": 3, "B": 2, "C": 3},
		"board-count":              {"A": -1, "B": -2, "C": 0},
		"top-board-results":        {"A": 1, "B": 0, "C": 0},
		"bottom-board-elimination": {"A": 1, "B": 0, "C": 0},
	}
	for id, want := range wants {
		t.Run(id, func(t *testing.T) {
			tb, err := Get(id)
			if err != nil {
				t.Fatal(err)
			}
			values, err := tb.Compute(context.Background(), state, scores)
			if err != nil {
				t.Fatal(err)
			}
			if len(values) != len(scores) {
				t.Fatalf("%s returned %d values, want %d", id, len(values), len(scores))
			}
			for _, value := range values {
				if value.Value != want[value.PlayerID] {
					t.Errorf("%s = %v, want %v", value.PlayerID, value.Value, want[value.PlayerID])
				}
			}
		})
	}
}

func TestTeamBoardTieBreakers_TwoMatches(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Matches: []chesspairing.MatchData{{
					HomeID: "A",
					AwayID: "B",
					Boards: []chesspairing.GameData{
						{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins}, // A gets 1
						{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultBlackWins}, // A gets 0
					},
				}},
			},
			{
				Number: 2,
				Matches: []chesspairing.MatchData{{
					HomeID: "A",
					AwayID: "C",
					Boards: []chesspairing.GameData{
						{WhiteID: "A", BlackID: "C", Result: chesspairing.ResultDraw}, // A gets 0.5
						{WhiteID: "A", BlackID: "C", Result: chesspairing.ResultDraw}, // A gets 0.5
					},
				}},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 2, Team: &chesspairing.TeamPoints{Match: 2, Game: 2}},
	}

	// A points: Match 1: [1, 0]. Match 2: [0.5, 0.5].
	// TopBoardResults: 1 + 0.5 = 1.5
	// BottomBoardElimination: Match 1 drops 0 -> 1. Match 2 drops 0.5 -> 0.5. Total: 1.5
	wants := map[string]float64{
		"top-board-results":        1.5,
		"bottom-board-elimination": 1.5,
	}
	for id, want := range wants {
		t.Run(id, func(t *testing.T) {
			tb, err := Get(id)
			if err != nil {
				t.Fatal(err)
			}
			values, err := tb.Compute(context.Background(), state, scores)
			if err != nil {
				t.Fatal(err)
			}
			if values[0].Value != want {
				t.Errorf("%s = %v, want %v", id, values[0].Value, want)
			}
		})
	}
}
