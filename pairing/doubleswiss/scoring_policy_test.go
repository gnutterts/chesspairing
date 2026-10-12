package doubleswiss

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestFootballDefaultPABAndOverride(t *testing.T) {
	cfg := cp.ScoringConfig{System: cp.ScoringFootball}
	if got := defaultPABPoints(cfg); got != 4 {
		t.Fatalf("football PAB = %v, want 4", got)
	}
	if got := defaultPABPoints(cp.ScoringConfig{}); got != 1.5 {
		t.Fatalf("standard PAB = %v, want 1.5", got)
	}
	state := &cp.TournamentState{ScoringConfig: cfg, CurrentRound: 2, Players: []cp.PlayerEntry{{ID: "a"}}, Rounds: []cp.RoundData{{Byes: []cp.ByeEntry{{PlayerID: "a", Type: cp.ByePAB}}}}}
	for _, want := range []float64{defaultPABPoints(cfg), 2.5} {
		participants, err := lexswiss.BuildParticipantStatesWithPAB(state, want)
		if err != nil {
			t.Fatal(err)
		}
		if len(participants) != 1 || participants[0].Score != want {
			t.Fatalf("PAB override %v: %v", want, participants)
		}
	}
}
