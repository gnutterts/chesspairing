// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestFootballTieBreakPoints(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:       []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}},
		ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringFootball},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultDraw}}},
		},
	}
	scores := []chesspairing.PlayerScore{{PlayerID: "A", Score: 4}, {PlayerID: "B", Score: 1}}
	if got := buildOpponentRecords(state, scores).records["A"][0].Points; got != 3 {
		t.Errorf("football win record = %v, want 3", got)
	}
	for _, tc := range []struct {
		id   string
		want []float64
	}{
		{"koya", []float64{0, 1}},        // B is below two rounds × 3 / 2.
		{"progressive", []float64{7, 1}}, // A: 3 + (3 + 1).
		{"win", []float64{1, 0}},
		{"sonneborn-berger", []float64{4, 4}}, // A: 3 × B's score 1 + 1 × 1.
		{"standard-points", []float64{1.5, 0.5}},
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
			for i, want := range tc.want {
				if values[i].Value != want {
					t.Errorf("%s = %v, want %v", values[i].PlayerID, values[i].Value, want)
				}
			}
		})
	}
}

func TestFootballTieBreakPointOverrides(t *testing.T) {
	state := &chesspairing.TournamentState{ScoringConfig: chesspairing.ScoringConfig{
		System:  chesspairing.ScoringFootball,
		Options: map[string]any{"pointWin": 4, "pointBye": 1, "pointForfeitWin": 2},
	}}
	opts := tournamentPoints(state)
	if *opts.PointWin != 4 || *opts.PointDraw != 1 || *opts.PointLoss != 0 || *opts.PointBye != 1 || *opts.PointForfeitWin != 2 || *opts.PointForfeitLoss != 0 {
		t.Errorf("football override points = %+v", opts)
	}
	state.ScoringConfig.Options = nil
	if got := *tournamentPoints(state).PointBye; got != 3 {
		t.Errorf("football default bye = %v, want 3", got)
	}
	state.ScoringConfig.System = chesspairing.ScoringSystem("other")
	if got := *tournamentPoints(state).PointWin; got != 1 {
		t.Errorf("unknown system win = %v, want standard default 1", got)
	}
}

func TestFootballWinExcludesDrawValueBye(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:       []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}},
		ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringFootball, Options: map[string]any{"pointBye": 1}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins}}},
			{Number: 2, Byes: []chesspairing.ByeEntry{{PlayerID: "A", Type: chesspairing.ByePAB}}},
		},
	}
	values, err := (&Win{}).Compute(context.Background(), state, []chesspairing.PlayerScore{{PlayerID: "A"}})
	if err != nil {
		t.Fatal(err)
	}
	if values[0].Value != 1 {
		t.Errorf("football WIN with one 3-point win and one 1-point bye = %v, want 1", values[0].Value)
	}
}
