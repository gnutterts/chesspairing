package team

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestConfiguredTeamAbsencePoints(t *testing.T) {
	state := &cp.TournamentState{
		CurrentRound: 2, Players: []cp.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
		ScoringConfig: cp.ScoringConfig{Options: map[string]any{"pointAbsent": 0.7, "pointExcused": 0.4, "pointClubCommitment": 0.3, "pointLoss": 0.2}},
		Rounds:        []cp.RoundData{{TeamByes: []cp.ByeEntry{{PlayerID: "a", Type: cp.ByeAbsent}, {PlayerID: "b", Type: cp.ByeExcused}, {PlayerID: "c", Type: cp.ByeClubCommitment}, {PlayerID: "d", Type: cp.ByeZero}}}},
	}
	participants, err := buildParticipantStates(state, "game")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]float64{"a": .7, "b": .4, "c": .3, "d": .2}
	for _, p := range participants {
		if p.Score != want[p.ID] {
			t.Errorf("%s score = %v, want %v", p.ID, p.Score, want[p.ID])
		}
	}
}
