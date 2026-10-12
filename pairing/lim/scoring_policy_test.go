package lim

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestFootballFloatHistoryAndMedian(t *testing.T) {
	state := &cp.TournamentState{ScoringConfig: cp.ScoringConfig{System: cp.ScoringFootball}, CurrentRound: 3, Rounds: []cp.RoundData{
		{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultWhiteWins}, {WhiteID: "c", BlackID: "d", Result: cp.ResultDraw}}},
		{Games: []cp.GameData{{WhiteID: "a", BlackID: "c", Result: cp.ResultDraw}}},
	}}
	scores := map[string]float64{}
	addRoundScores(scores, state.Rounds[0], state, 1)
	if scores["a"] != 3 || scores["b"] != 0 || scores["c"] != 1 || scores["d"] != 1 {
		t.Fatalf("football pairing scores: %v", scores)
	}
	players, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range players {
		if p.ID == "a" && p.Score != 4 {
			t.Fatalf("Lim tournament score for a = %v, want 4", p.Score)
		}
	}
	if got := previousRoundFloaters(state); !got["a"] || !got["c"] {
		t.Fatalf("median 1.5 after one round: floaters %v, want a and c", got)
	}
	groups := []swisslib.ScoreGroup{{Score: 3}, {Score: 1.5}, {Score: 1}}
	above, below, median := splitByMedian(groups, swisslib.PairingWinPoints(state.ScoringConfig)*float64(1)/2)
	if len(above) != 1 || len(below) != 1 || median == nil || median.Score != 1.5 {
		t.Fatalf("football median groups: %v %v %v", above, below, median)
	}
}

func TestUnrecordedAbsenceAndDecimalFloatHistory(t *testing.T) {
	state := &cp.TournamentState{
		CurrentRound:  3,
		Players:       []cp.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		ScoringConfig: cp.ScoringConfig{Options: map[string]any{"pointAbsent": 0.5}},
		Rounds: []cp.RoundData{
			{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultDraw}}},
			{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultDraw}}},
		},
	}
	scores := map[string]float64{}
	addRoundScores(scores, state.Rounds[0], state, 1)
	players, err := swisslib.BuildPlayerStates(&cp.TournamentState{
		CurrentRound: 2, Players: state.Players, ScoringConfig: state.ScoringConfig, Rounds: state.Rounds[:1],
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, player := range players {
		if scores[player.ID] != player.Score {
			t.Fatalf("Lim history score for %s = %v, player score = %v", player.ID, scores[player.ID], player.Score)
		}
	}
	if got := previousRoundFloaters(state); len(got) != 0 {
		t.Fatalf("equal-score game produced floaters: %v", got)
	}

	state.ScoringConfig.Options = map[string]any{"pointDraw": 0.1, "pointWin": 0.3}
	state.CurrentRound = 5
	// A score accumulated from three draws must compare equal to one win.
	state.Rounds = []cp.RoundData{
		{Games: []cp.GameData{{WhiteID: "a", BlackID: "c", Result: cp.ResultDraw}}},
		{Games: []cp.GameData{{WhiteID: "a", BlackID: "c", Result: cp.ResultDraw}}},
		{Games: []cp.GameData{{WhiteID: "a", BlackID: "c", Result: cp.ResultDraw}, {WhiteID: "b", BlackID: "c", Result: cp.ResultWhiteWins}}},
		{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultDraw}}},
	}
	if got := previousRoundFloaters(state); len(got) != 0 {
		t.Fatalf("decimal equal-score game produced floaters: %v", got)
	}
}
