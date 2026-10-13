// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestConfiguredProgressiveScore(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}},
		ScoringConfig: chesspairing.ScoringConfig{Options: map[string]any{
			"pointWin": 3, "pointDraw": 1, "pointLoss": 0, "pointForfeitWin": 2,
			"pointBye": 3,
		}},
		Rounds: []chesspairing.RoundData{
			{Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins}}},
			{Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultDraw}}},
			{Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultForfeitWhiteWins}}},
			{Byes: []chesspairing.ByeEntry{{PlayerID: "A", Type: chesspairing.ByePAB}}},
		},
	}
	scores := []chesspairing.PlayerScore{{PlayerID: "A"}}
	for _, tc := range []struct {
		cut  bool
		want float64
	}{{false, 22}, {true, 19}} {
		values, err := (&Progressive{cut1: tc.cut}).Compute(context.Background(), state, scores)
		if err != nil {
			t.Fatal(err)
		}
		if values[0].Value != tc.want {
			t.Errorf("cut1=%v: got %v, want %v", tc.cut, values[0].Value, tc.want)
		}
	}
}

func TestConfiguredKoyaThreshold(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:       []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"}, {ID: "E"}},
		ScoringConfig: chesspairing.ScoringConfig{Options: map[string]any{"pointWin": 3, "pointDraw": 1}},
		Rounds: []chesspairing.RoundData{
			{Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "C", Result: chesspairing.ResultDraw}}},
			{Games: []chesspairing.GameData{{WhiteID: "B", BlackID: "D", Result: chesspairing.ResultDraw}}},
			{Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "E", Result: chesspairing.ResultDraw}}},
		},
	}
	// E's score qualifies under rounds/2, but not under rounds*pointWin/2.
	scores := []chesspairing.PlayerScore{{PlayerID: "A", Score: 5}, {PlayerID: "B", Score: 1}, {PlayerID: "C", Score: 5}, {PlayerID: "D", Score: 1}, {PlayerID: "E", Score: 3}}
	values, err := (&Koya{}).Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range values {
		want := 0.0
		if i == 0 || i == 2 || i == 4 {
			want = 1
		}
		if value.Value != want {
			t.Errorf("%s Koya = %v, want %v", scores[i].PlayerID, value.Value, want)
		}
	}
}

func TestConfiguredWinAward(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		ScoringConfig: chesspairing.ScoringConfig{Options: map[string]any{
			"pointWin": 3, "pointForfeitWin": 2, "pointBye": 3,
		}},
		Rounds: []chesspairing.RoundData{
			{Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultForfeitWhiteWins}}, Byes: []chesspairing.ByeEntry{{PlayerID: "C", Type: chesspairing.ByePAB}}},
			{Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins}}},
		},
	}
	values, err := (&Win{}).Compute(context.Background(), state, []chesspairing.PlayerScore{{PlayerID: "A"}, {PlayerID: "B"}, {PlayerID: "C"}})
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []float64{1, 0, 1} {
		if values[i].Value != want {
			t.Errorf("%s WIN = %v, want %v", values[i].PlayerID, values[i].Value, want)
		}
	}
}

func TestPerfectPerformanceLowestInteger(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ratings []int
		score   float64
		want    float64
	}{
		{"mixed", []int{2000, 1900, 1800, 1700}, 2, 1847},
		{"equal", []int{1500, 1500, 1500, 1500, 1500, 1500}, 2, 1379},
		// Table 8.1b: at 1879 the -121 difference yields 0.34, not 0.33.
		// At 1878 it yields 0.33; 1880 is sufficient but not the lowest.
		{"negative-121-boundary", []int{2000, 2000, 2000}, 1, 1879},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := &chesspairing.TournamentState{Players: []chesspairing.PlayerEntry{{ID: "A"}}}
			for i, rating := range tc.ratings {
				id := fmt.Sprintf("opp%d", i)
				state.Players = append(state.Players, chesspairing.PlayerEntry{ID: id, Rating: rating})
				state.Rounds = append(state.Rounds, chesspairing.RoundData{Number: i + 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: id, Result: chesspairing.ResultDraw}}})
			}
			values, err := (&PerformancePoints{}).Compute(context.Background(), state, []chesspairing.PlayerScore{{PlayerID: "A", Score: tc.score}})
			if err != nil {
				t.Fatal(err)
			}
			if values[0].Value != tc.want {
				t.Errorf("PTP = %v, want %v", values[0].Value, tc.want)
			}
		})
	}
}

func TestExpectedScoreNegativeTieSymmetry(t *testing.T) {
	if got := expectedScore(-121); got < 0.339999999 || got > 0.340000001 {
		t.Errorf("expectedScore(-121) = %v, want 0.34", got)
	}
	if got := expectedScore(121) + expectedScore(-121); got != 1 {
		t.Errorf("opposite differences sum to %v, want 1", got)
	}
}

func TestPlayerRatingUnratedRule(t *testing.T) {
	state := &chesspairing.TournamentState{Players: []chesspairing.PlayerEntry{{ID: "A", Rating: 0}, {ID: "B", Rating: 1800}}}
	scores := []chesspairing.PlayerScore{{PlayerID: "A"}, {PlayerID: "B"}}
	_, err := (&PlayerRating{}).Compute(context.Background(), state, scores)
	if err == nil || !strings.Contains(err.Error(), "unratedRating") {
		t.Fatalf("unrated RTNG error = %v", err)
	}
	state.ScoringConfig.Options = map[string]any{"unratedRating": 1400}
	values, err := (&PlayerRating{}).Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	if values[0].Value != 1400 || values[1].Value != 1800 {
		t.Errorf("RTNG with override = %v", values)
	}
}
