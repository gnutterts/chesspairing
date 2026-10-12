package lim

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPreassignedPABUsesLimScale(t *testing.T) {
	state := &chesspairing.TournamentState{ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: map[string]any{"pointForfeitWin": 0.5}}, Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}}, Rounds: []chesspairing.RoundData{{Games: []chesspairing.GameData{{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultForfeitWhiteWins}}}}, CurrentRound: 2, PreAssignedByes: []chesspairing.ByeEntry{{PlayerID: "a", Type: chesspairing.ByePAB}}}
	if _, err := New(Options{}).Pair(context.Background(), state); err == nil {
		t.Fatal("Lim accepted PAB awarded after a full unplayed win on its 1-point scale")
	}
}
