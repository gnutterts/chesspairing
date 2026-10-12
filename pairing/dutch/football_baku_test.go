package dutch

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestFootballDefaultWinBaku(t *testing.T) {
	acceleration := "baku"
	total := 9
	state := &chesspairing.TournamentState{ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringFootball, Options: map[string]any{"pointDraw": 1.5}}, CurrentRound: 1, Players: []chesspairing.PlayerEntry{{ID: "a", Rating: 2400}, {ID: "b", Rating: 2300}, {ID: "c", Rating: 2200}, {ID: "d", Rating: 2100}}}
	result, err := New(Options{Acceleration: &acceleration, TotalRounds: &total}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(result.Notes, " "), "VP=3.0") {
		t.Fatalf("notes: %v", result.Notes)
	}
	players, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		t.Fatal(err)
	}
	groupA, err := swisslib.BakuGroupA(state.Players)
	if err != nil {
		t.Fatal(err)
	}
	swisslib.AddVirtualPoints(players, func(id string) float64 {
		return swisslib.BakuVirtualPoints(swisslib.PairingWinPoints(state.ScoringConfig), total, 1, groupA[id])
	})
	for _, p := range players {
		if groupA[p.ID] && p.PairingScore != 3 {
			t.Errorf("%s pairing score = %g", p.ID, p.PairingScore)
		}
	}
}
