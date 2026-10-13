// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestC07ConfiguredPointsAndDummy(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		ScoringConfig: chesspairing.ScoringConfig{Options: map[string]any{
			"pointWin": 3, "pointDraw": 1, "pointLoss": 0,
		}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins}}, Byes: []chesspairing.ByeEntry{{PlayerID: "C", Type: chesspairing.ByeHalf}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "B", BlackID: "C", Result: chesspairing.ResultWhiteWins}}, Byes: []chesspairing.ByeEntry{{PlayerID: "A", Type: chesspairing.ByeZero}}},
		},
	}
	scores := []chesspairing.PlayerScore{{PlayerID: "A", Score: 3}, {PlayerID: "B", Score: 3}, {PlayerID: "C", Score: 1}}
	table := buildOpponentRecords(state, scores)
	if table.records["A"][0].Points != 3 || table.records["C"][0].Points != 1 || table.adjustedScores["A"] != 4 || table.adjustedScores["C"] != 1 {
		t.Fatalf("configured records or Article 16.3 adjusted scores: A=%+v, C=%+v, adjusted=%v", table.records["A"], table.records["C"], table.adjustedScores)
	}
	if got := dummyScore("A", table.records["A"][1], table, true); got != 2 {
		t.Errorf("Article 16.4 draw cap = %v, want 2", got)
	}
	std, err := (&StandardPoints{}).Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []float64{1, 1, 0.5} {
		if std[i].Value != want {
			t.Errorf("standard points for %s = %v, want %v", std[i].PlayerID, std[i].Value, want)
		}
	}
	for _, tc := range []struct {
		id   string
		want map[string]float64
	}{
		{"sonneborn-berger", map[string]float64{"A": 9, "B": 3, "C": 1}},
		{"buchholz", map[string]float64{"A": 5, "B": 5, "C": 4}},
	} {
		t.Run(tc.id, func(t *testing.T) {
			tb, err := Get(tc.id)
			if err != nil {
				t.Fatal(err)
			}
			values, err := tb.Compute(context.Background(), state, scores)
			if err != nil {
				t.Fatal(err)
			}
			for _, value := range values {
				if value.Value != tc.want[value.PlayerID] {
					t.Errorf("%s = %v, want %v", value.PlayerID, value.Value, tc.want[value.PlayerID])
				}
			}
		})
	}
}

func TestC07ConfiguredForfeitAndStandardPoints(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}},
		ScoringConfig: chesspairing.ScoringConfig{Options: map[string]any{
			"pointWin": 3, "pointDraw": 1, "pointForfeitWin": 2,
		}},
		Rounds: []chesspairing.RoundData{{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultForfeitWhiteWins}}}},
	}
	scores := []chesspairing.PlayerScore{{PlayerID: "A", Score: 2}, {PlayerID: "B", Score: 0}}
	if got := buildOpponentRecords(state, scores).records["A"][0].Points; got != 2 {
		t.Errorf("forfeit win record = %v, want 2", got)
	}
	values, err := (&StandardPoints{}).Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	if values[0].Value != 1 || values[1].Value != 0 {
		t.Errorf("standard points = %v, want 1 and 0", values)
	}
}

func TestC07BoardCountRequiresEqualGamePoints(t *testing.T) {
	scores := []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 4, Team: &chesspairing.TeamPoints{Match: 4, Game: 7}},
		{PlayerID: "B", Score: 4, Team: &chesspairing.TeamPoints{Match: 4, Game: 6}},
	}
	_, err := (&BoardCount{}).Compute(context.Background(), &chesspairing.TournamentState{}, scores)
	if err == nil || !strings.Contains(err.Error(), "equal game points") {
		t.Fatalf("board count on tied teams with unequal game points: %v", err)
	}
}
