package burstein

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPreassignedPABUsesBursteinScale(t *testing.T) {
	state := &chesspairing.TournamentState{ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: map[string]any{"pointForfeitWin": 0.5}}, Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}}, Rounds: []chesspairing.RoundData{{Games: []chesspairing.GameData{{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultForfeitWhiteWins}}}}, CurrentRound: 2, PreAssignedByes: []chesspairing.ByeEntry{{PlayerID: "a", Type: chesspairing.ByePAB}}}
	if _, err := New(Options{}).Pair(context.Background(), state); err != nil {
		t.Fatalf("half-point forfeit win must not disqualify a PAB: %v", err)
	}
}
