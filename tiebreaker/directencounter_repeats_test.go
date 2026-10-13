// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker_test

import (
	"context"
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/tiebreaker"
)

func TestSwissDirectEncounterMissingGameMaximumAward(t *testing.T) {
	state := &cp.TournamentState{
		Players: []cp.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		Rounds: []cp.RoundData{
			{Games: []cp.GameData{{WhiteID: "A", BlackID: "B", Result: cp.ResultWhiteWins}}},
			{Games: []cp.GameData{{WhiteID: "A", BlackID: "C", Result: cp.ResultWhiteWins}}},
		},
	}
	group := []cp.PlayerScore{{PlayerID: "A", Score: 2}, {PlayerID: "B", Score: 2}, {PlayerID: "C", Score: 2}}
	values := map[string]float64{"A": 2, "B": 0, "C": 0}
	for _, tc := range []struct {
		name    string
		options map[string]any
		want    string
	}{
		{"ordinary-win", nil, "A"},
		{"forfeit-win", map[string]any{"pointForfeitWin": 3}, ""},
		{"bye", map[string]any{"pointBye": 3}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state.ScoringConfig.Options = tc.options
			leader, complete := tiebreaker.SwissDirectEncounterLeader(state, group, values)
			if complete || leader != tc.want {
				t.Errorf("leader = %q, complete = %v; want %q, false", leader, complete, tc.want)
			}
		})
	}
}

func TestDirectEncounterQuarterAverage(t *testing.T) {
	state := &cp.TournamentState{
		Players: []cp.PlayerEntry{{ID: "B"}, {ID: "C"}},
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{{WhiteID: "B", BlackID: "C", Result: cp.ResultWhiteWins}}},
			{Number: 2, Games: []cp.GameData{{WhiteID: "B", BlackID: "C", Result: cp.ResultBlackWins}}},
			{Number: 3, Games: []cp.GameData{{WhiteID: "B", BlackID: "C", Result: cp.ResultBlackWins}}},
			{Number: 4, Games: []cp.GameData{{WhiteID: "B", BlackID: "C", Result: cp.ResultBlackWins}}},
		},
	}
	scores := []cp.PlayerScore{{PlayerID: "B", Score: 2}, {PlayerID: "C", Score: 2}}
	tb, _ := tiebreaker.Get("direct-encounter")
	got, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Value != 0.25 || got[1].Value != 0.75 {
		t.Errorf("quarter averages: %+v", got)
	}
}

func TestDirectEncounterAveragesRepeatedOpponents(t *testing.T) {
	// Repeated B-C and D-F encounters contribute one average each.
	state := &cp.TournamentState{
		Players: []cp.PlayerEntry{{ID: "B"}, {ID: "C"}, {ID: "D"}, {ID: "F"}},
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{{WhiteID: "B", BlackID: "C", Result: cp.ResultWhiteWins}, {WhiteID: "D", BlackID: "F", Result: cp.ResultDraw}}},
			{Number: 2, Games: []cp.GameData{{WhiteID: "C", BlackID: "B", Result: cp.ResultWhiteWins}, {WhiteID: "F", BlackID: "D", Result: cp.ResultDraw}}},
			{Number: 3, Games: []cp.GameData{{WhiteID: "C", BlackID: "D", Result: cp.ResultWhiteWins}, {WhiteID: "F", BlackID: "B", Result: cp.ResultWhiteWins}}},
		},
	}
	scores := []cp.PlayerScore{{PlayerID: "B", Score: 2}, {PlayerID: "C", Score: 2}, {PlayerID: "D", Score: 2}, {PlayerID: "F", Score: 2}}
	tb, _ := tiebreaker.Get("direct-encounter")
	got, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]float64{"B": 0.5, "C": 1.5, "D": 0.5, "F": 1.5}
	for _, v := range got {
		if v.Value != want[v.PlayerID] {
			t.Errorf("%s = %v, want %v", v.PlayerID, v.Value, want[v.PlayerID])
		}
	}
}
