package tiebreaker_test

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/team"
	"github.com/gnutterts/chesspairing/standings"
	"github.com/gnutterts/chesspairing/tiebreaker"
)

func TestFIDEExercise_34_MPVGPStand_Order(t *testing.T) {
	state := tiebreaker.ExportO5TeamState()
	// FideAOrderForExternalTest is populated when TestFIDEExercise_34_MPVGPStand runs,
	// but we can just redefine it here to be safe, or call the exported function.
	fideA := map[string]int{
		"T5": 1, "T1": 2, "T2": 3, "T4": 4, "T3": 5, "T8": 6, "T6": 7,
		"T9": 8, "T13": 9, "T7": 10, "T10": 11, "T11": 12, "T14": 13, "T12": 14,
	}

	scorer := team.New(team.Options{PrimaryScore: "match"})
	tb, err := tiebreaker.Get("mpvgp")
	if err != nil {
		t.Fatal(err)
	}

	rows, err := standings.Build(context.Background(), state, scorer, []chesspairing.TieBreaker{tb})
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range rows {
		if fideA[row.PlayerID] != i+1 {
			t.Errorf("expected %s at place %d, got %d", row.PlayerID, fideA[row.PlayerID], i+1)
		}
	}
}
