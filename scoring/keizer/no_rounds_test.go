package keizer

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestScoreNoRoundsAddsSelfVictory(t *testing.T) {
	state := &chesspairing.TournamentState{Players: []chesspairing.PlayerEntry{{ID: "A", Rating: 2}, {ID: "B", Rating: 1}}}
	winFraction := 0.5
	got, err := New(Options{WinFraction: &winFraction}).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].PlayerID != "A" || got[0].Score != 1 || got[1].PlayerID != "B" || got[1].Score != 0.5 {
		t.Fatalf("scores = %+v, want A=1 and B=0.5", got)
	}
}
