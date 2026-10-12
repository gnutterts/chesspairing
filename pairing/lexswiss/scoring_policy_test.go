package lexswiss

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestFootballGameByeAndPABPoints(t *testing.T) {
	state := &cp.TournamentState{
		ScoringConfig: cp.ScoringConfig{System: cp.ScoringFootball}, CurrentRound: 2,
		Players: []cp.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"}, {ID: "f"}},
		Rounds: []cp.RoundData{{Games: []cp.GameData{
			{WhiteID: "a", BlackID: "b", Result: cp.ResultWhiteWins},
			{WhiteID: "c", BlackID: "d", Result: cp.ResultDraw},
		}, Byes: []cp.ByeEntry{{PlayerID: "e", Type: cp.ByePAB}, {PlayerID: "f", Type: cp.ByeAbsent}}}},
	}
	state.ScoringConfig.Options = map[string]any{"pointAbsent": 0.2}
	for _, tt := range []struct{ pab, want float64 }{{4, 4}, {2.5, 2.5}} {
		participants, err := BuildParticipantStatesWithPAB(state, tt.pab)
		if err != nil {
			t.Fatal(err)
		}
		scores := map[string]float64{}
		for _, p := range participants {
			scores[p.ID] = p.Score
		}
		if scores["a"] != 3 || scores["b"] != 0 || scores["c"] != 1 || scores["d"] != 1 || scores["e"] != tt.want || scores["f"] != .2 {
			t.Fatalf("PAB %v: scores %v", tt.pab, scores)
		}
	}
}

func TestDecimalScoreGroupsNormalize(t *testing.T) {
	first, second := 0.1, 0.2
	participants := []ParticipantState{{ID: "a", Score: first + second}, {ID: "b", Score: .3}}
	groups := BuildScoreGroups(participants)
	if len(groups) != 1 || len(groups[0].Participants) != 2 || groups[0].Score != .3 {
		t.Fatalf("decimal score groups = %v", groups)
	}
	if participants[0].Score != first+second {
		t.Fatalf("grouping changed caller's score: %v", participants[0].Score)
	}
}

func TestUnrecordedAbsenceUsesConfiguredPoints(t *testing.T) {
	state := &cp.TournamentState{
		CurrentRound:  2,
		Players:       []cp.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		ScoringConfig: cp.ScoringConfig{Options: map[string]any{"pointAbsent": 0.5}},
		Rounds:        []cp.RoundData{{Games: []cp.GameData{{WhiteID: "a", BlackID: "b", Result: cp.ResultDraw}}}},
	}
	participants, err := BuildParticipantStates(state)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range participants {
		if p.ID == "c" && p.Score != 0.5 {
			t.Fatalf("unrecorded absence score = %v, want 0.5", p.Score)
		}
	}
}
