package team

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestReportedTotalsOverrideIncompleteBoards(t *testing.T) {
	match := cp.MatchData{
		HomeID: "A",
		AwayID: "B",
		Boards: []cp.GameData{
			{WhiteID: "A", BlackID: "B", Result: cp.ResultWhiteWins},
			{WhiteID: "A", BlackID: "B", Result: cp.ResultPending},
		},
		Result: &cp.TeamMatchResult{HomeGame: 0.5, AwayGame: 1.5},
	}
	home, away := MatchGamePoints(match, Options{}.WithDefaults().Options, map[string]string{"A": "A", "B": "B"})
	if home != 0.5 || away != 1.5 {
		t.Errorf("reported totals = %v-%v, want 0.5-1.5", home, away)
	}
}

func TestUndecidedMatchDoesNotScore(t *testing.T) {
	st := &cp.TournamentState{
		Players:      []cp.PlayerEntry{{ID: "A"}, {ID: "B"}},
		Rounds:       []cp.RoundData{{Number: 1, Matches: []cp.MatchData{{HomeID: "A", AwayID: "B"}}}},
		CurrentRound: 1,
	}
	got := scores(t, st)
	if got["A"].Match != 0 {
		t.Errorf("unreported match scored as draw: %+v", got)
	}
}
