package burstein

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPreassignedPABUsesBursteinScale(t *testing.T) {
	state := &chesspairing.TournamentState{ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: map[string]any{"pointForfeitWin": 0.5}}, Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}}, Rounds: []chesspairing.RoundData{{Games: []chesspairing.GameData{{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultForfeitWhiteWins}}}}, CurrentRound: 2, PreAssignedByes: []chesspairing.ByeEntry{{PlayerID: "a", Type: chesspairing.ByePAB}}}
	if _, err := New(Options{}).Pair(context.Background(), state); err == nil || !strings.Contains(err.Error(), "C2") {
		t.Fatalf("expected C2 rejection, got %v", err)
	}
}
