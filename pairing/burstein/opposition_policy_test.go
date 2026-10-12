package burstein

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestFootballOppositionUsesTournamentPoints(t *testing.T) {
	state := &cp.TournamentState{ScoringConfig: cp.ScoringConfig{System: cp.ScoringFootball}, CurrentRound: 3, Players: []cp.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}}, Rounds: []cp.RoundData{
		{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultWhiteWins}}, Byes: []cp.ByeEntry{{PlayerID: "c", Type: cp.ByeZero}}},
		{Games: []cp.GameData{{WhiteID: "a", BlackID: "c", Result: cp.ResultDraw}}, Byes: []cp.ByeEntry{{PlayerID: "b", Type: cp.ByeHalf}}},
	}}
	scores := computePairingScores(state)
	if scores["a"] != 4 || scores["b"] != 1 || scores["c"] != 1 {
		t.Fatalf("football scores = %v", scores)
	}
	index := ComputeOppositionIndex(&swisslib.PlayerState{ID: "a"}, state)
	// b's preceding half-point bye breaks its zero-bye run; c's draw is worth one.
	if index.Buchholz != 2 || index.SonnebornBerger != 4 {
		t.Fatalf("football opposition = %+v, want BH=2 SB=4", index)
	}
}

func TestUnrecordedAbsenceMatchesPlayerScore(t *testing.T) {
	state := &cp.TournamentState{
		CurrentRound:  2,
		Players:       []cp.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		ScoringConfig: cp.ScoringConfig{Options: map[string]any{"pointAbsent": 0.5}},
		Rounds:        []cp.RoundData{{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultDraw}}}},
	}
	players, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		t.Fatal(err)
	}
	scores := computePairingScores(state)
	for _, player := range players {
		if scores[player.ID] != player.Score {
			t.Fatalf("opposition score for %s = %v, player score = %v", player.ID, scores[player.ID], player.Score)
		}
		if player.ID == "c" {
			index := ComputeOppositionIndex(&player, state)
			if index.Buchholz != player.Score {
				t.Fatalf("absent player's self-play Buchholz = %v, score = %v", index.Buchholz, player.Score)
			}
		}
	}
}

func TestDecimalOppositionIndexNormalizes(t *testing.T) {
	state := &cp.TournamentState{
		CurrentRound:  4,
		ScoringConfig: cp.ScoringConfig{Options: map[string]any{"pointWin": 0.3, "pointDraw": 0.1}},
		Rounds: []cp.RoundData{
			{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultDraw}}},
			{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultDraw}}},
			{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultDraw}}},
		},
	}
	index := ComputeOppositionIndex(&swisslib.PlayerState{ID: "a"}, state)
	if index.Buchholz != 0.9 || index.SonnebornBerger != 0.09 {
		t.Fatalf("decimal opposition index = %+v, want normalized BH=0.9 SB=0.09", index)
	}
}
