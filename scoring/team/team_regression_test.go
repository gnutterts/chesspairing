package team

import (
	"context"
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func scores(t *testing.T, st *cp.TournamentState) map[string]cp.TeamPoints {
	out, err := New(Options{}).Score(context.Background(), st)
	if err != nil {
		t.Fatal(err)
	}
	r := map[string]cp.TeamPoints{}
	for _, s := range out {
		r[s.PlayerID] = *s.Team
	}
	return r
}

// Team-as-player model (no TeamID), boards stored under team IDs with alternating colours.
// A scores only the first board; B scores the other three by board side.
func TestBoardAttributionWithoutTeamID(t *testing.T) {
	st := &cp.TournamentState{
		Players: []cp.PlayerEntry{{ID: "A"}, {ID: "B"}},
		Rounds: []cp.RoundData{{Number: 1, Matches: []cp.MatchData{{HomeID: "A", AwayID: "B", Boards: []cp.GameData{
			{WhiteID: "A", BlackID: "B", Result: cp.ResultWhiteWins},
			{WhiteID: "B", BlackID: "A", Result: cp.ResultWhiteWins},
			{WhiteID: "A", BlackID: "B", Result: cp.ResultBlackWins},
			{WhiteID: "B", BlackID: "A", Result: cp.ResultWhiteWins},
		}}}}},
		CurrentRound: 2,
	}
	got := scores(t, st)
	t.Logf("%+v", got)
	if got["A"].Game != 1 || got["B"].Game != 3 {
		t.Errorf("want A=1 B=3 GP, got %+v", got)
	}
}

// One double-forfeit board in an otherwise decided match.
func TestMixedRosterDoesNotCreateTeamForUnassignedPlayer(t *testing.T) {
	st := &cp.TournamentState{
		Players: []cp.PlayerEntry{
			{ID: "a1", TeamID: "A"}, {ID: "b1", TeamID: "B"}, {ID: "unassigned"},
		},
		Rounds: []cp.RoundData{{Number: 1, Matches: []cp.MatchData{{
			HomeID: "A", AwayID: "B", Boards: []cp.GameData{{WhiteID: "a1", BlackID: "b1", Result: cp.ResultWhiteWins}},
		}}}},
		CurrentRound: 2,
	}
	got := scores(t, st)
	if len(got) != 2 || got["A"].Game != 1 || got["B"].Game != 0 {
		t.Errorf("scores = %+v, want only teams A and B", got)
	}
	if _, ok := got["unassigned"]; ok {
		t.Errorf("unassigned player was incorrectly made a team: %+v", got)
	}
}

func TestDoubleForfeitBoardDoesNotCancelMatch(t *testing.T) {
	st := &cp.TournamentState{
		Players: []cp.PlayerEntry{{ID: "a1", TeamID: "A"}, {ID: "a2", TeamID: "A"}, {ID: "b1", TeamID: "B"}, {ID: "b2", TeamID: "B"}},
		Rounds: []cp.RoundData{{Number: 1, Matches: []cp.MatchData{{HomeID: "A", AwayID: "B", Boards: []cp.GameData{
			{WhiteID: "a1", BlackID: "b1", Result: cp.ResultWhiteWins},
			{WhiteID: "b2", BlackID: "a2", Result: cp.ResultDoubleForfeit, IsForfeit: true},
		}}}}},
		CurrentRound: 2,
	}
	got := scores(t, st)
	t.Logf("%+v", got)
	if got["A"].Match != 2 {
		t.Errorf("A won 1-0 on boards but MP=%v", got["A"].Match)
	}
}
