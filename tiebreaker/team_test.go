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
		"board-count":              {"A": -4, "B": -2, "C": -3},
		"top-board-results":        {"A": 22, "B": 2, "C": 12},
		"bottom-board-elimination": {"A": 4, "B": 0, "C": 2},
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
	// TopBoardResults compares board totals lexicographically: [1.5, 0.5],
	// in half points [3, 1], encoded with base 4 as 13. BottomBoardElimination
	// compares [1.5], in half points 3.
	wants := map[string]float64{
		"top-board-results":        13,
		"bottom-board-elimination": 3,
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

// Art. 12.2/12.3 compare board by board; a half point on a higher board must
// outrank any total on the boards below it, and the cumulative sums of BBE
// must not overflow into the next position.
func TestTeamBoardTieBreakers_HalfPointsKeepBoardOrder(t *testing.T) {
	match := func(home, away string, results ...chesspairing.GameResult) chesspairing.MatchData {
		games := make([]chesspairing.GameData, len(results))
		for i, result := range results {
			games[i] = chesspairing.GameData{WhiteID: home, BlackID: away, Result: result}
		}
		return chesspairing.MatchData{HomeID: home, AwayID: away, Boards: games}
	}
	draw, win, loss := chesspairing.ResultDraw, chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "X"}, {ID: "Y"}, {ID: "Z"}},
		Rounds: []chesspairing.RoundData{
			// A: board totals [0.5, 0, 0]; B: [0, 1, 1] over two matches.
			{Number: 1, Matches: []chesspairing.MatchData{match("A", "X", draw, loss, loss), match("B", "Y", loss, win, win)}},
			{Number: 2, Matches: []chesspairing.MatchData{match("B", "Z", loss, loss, loss)}},
		},
	}
	scores := []chesspairing.PlayerScore{{PlayerID: "A"}, {PlayerID: "B"}}
	tb, err := Get("top-board-results")
	if err != nil {
		t.Fatal(err)
	}
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	if values[0].Value <= values[1].Value {
		t.Errorf("top-board-results A = %v, B = %v; A must rank higher (board 1: 0.5 > 0)", values[0].Value, values[1].Value)
	}
}
