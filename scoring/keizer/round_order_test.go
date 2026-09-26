package keizer

import (
	"context"
	"reflect"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestScoreUsesRoundNumberOrder(t *testing.T) {
	no := false
	r1 := chesspairing.RoundData{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "D", Result: chesspairing.ResultWhiteWins}}}
	r2 := chesspairing.RoundData{Number: 2, Games: []chesspairing.GameData{{WhiteID: "D", BlackID: "A", Result: chesspairing.ResultWhiteWins}}}
	players := []chesspairing.PlayerEntry{{ID: "A", Rating: 4}, {ID: "B", Rating: 3}, {ID: "C", Rating: 2}, {ID: "D", Rating: 1}}
	score := func(rounds []chesspairing.RoundData) []chesspairing.PlayerScore {
		got, err := New(Options{Frozen: chesspairing.BoolPtr(true), SelfVictory: &no}).Score(context.Background(), &chesspairing.TournamentState{Players: players, Rounds: rounds, CurrentRound: 3})
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	chronological := score([]chesspairing.RoundData{r1, r2})
	reversed := score([]chesspairing.RoundData{r2, r1})
	if !reflect.DeepEqual(chronological, reversed) {
		t.Fatalf("scores in round-number order = %+v, reversed slice = %+v", chronological, reversed)
	}
}

func TestScoreRejectsNonContiguousRoundNumbers(t *testing.T) {
	_, err := New(Options{}).Score(context.Background(), &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}},
		Rounds:  []chesspairing.RoundData{{Number: 2}},
	})
	if err == nil {
		t.Fatal("Score accepted non-contiguous round numbers")
	}
}
