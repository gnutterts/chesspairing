package tiebreaker

import (
	"context"
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func regressionValues(t *testing.T, id string, state *cp.TournamentState, scores []cp.PlayerScore) map[string]float64 {
	t.Helper()
	tb, err := Get(id)
	if err != nil {
		t.Fatal(err)
	}
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]float64, len(values))
	for _, value := range values {
		result[value.PlayerID] = value.Value
	}
	return result
}

func TestC07Article12BoardOrderAndTeamBye(t *testing.T) {
	state := &cp.TournamentState{
		Players:       []cp.PlayerEntry{{ID: "a1", TeamID: "A"}, {ID: "a2", TeamID: "A"}, {ID: "b1", TeamID: "B"}, {ID: "b2", TeamID: "B"}},
		ScoringConfig: cp.ScoringConfig{System: cp.ScoringTeam},
		Rounds:        []cp.RoundData{{Number: 1, Matches: []cp.MatchData{{HomeID: "A", AwayID: "B", Boards: []cp.GameData{{WhiteID: "a1", BlackID: "b1", Result: cp.ResultWhiteWins}, {WhiteID: "b2", BlackID: "a2", Result: cp.ResultWhiteWins}}}}}, {Number: 2, TeamByes: []cp.ByeEntry{{PlayerID: "A", Type: cp.ByePAB}}}},
	}
	scores := []cp.PlayerScore{{PlayerID: "A", Team: &cp.TeamPoints{}}, {PlayerID: "B", Team: &cp.TeamPoints{}}}
	if got := regressionValues(t, "board-count", state, scores)["A"]; got != -4 {
		t.Fatalf("BC for team bye = %v, want -4", got)
	}
	if got := regressionValues(t, "top-board-results", state, scores)["A"]; got != 22 {
		t.Fatalf("TBR = %v, want board-order encoding 22", got)
	}
	if got := regressionValues(t, "bottom-board-elimination", state, scores)["A"]; got != 4 {
		t.Fatalf("BBE = %v, want 4 (2 points in half points)", got)
	}
}

func TestC07Article16TeamDummyCaps(t *testing.T) {
	state := &cp.TournamentState{
		Players:       []cp.PlayerEntry{{ID: "a1", TeamID: "A"}, {ID: "a2", TeamID: "A"}},
		ScoringConfig: cp.ScoringConfig{System: cp.ScoringTeam},
		Rounds:        []cp.RoundData{{Number: 1, TeamByes: []cp.ByeEntry{{PlayerID: "A", Type: cp.ByeFullPoint}}}},
	}
	scores := []cp.PlayerScore{{PlayerID: "A", Score: 6, Team: &cp.TeamPoints{Match: 6, Game: 6}}}
	if got := regressionValues(t, "buchholz-mp", state, scores)["A"]; got != 1 {
		t.Fatalf("MP dummy = %v, want 1", got)
	}
	if got := regressionValues(t, "eggsb", state, scores)["A"]; got != 2 {
		t.Fatalf("GP dummy contribution = %v, want 2", got)
	}
}

func TestC07Article15RoundRobinForfeits(t *testing.T) {
	state := &cp.TournamentState{
		Players:       []cp.PlayerEntry{{ID: "A"}, {ID: "B"}},
		PairingConfig: cp.PairingConfig{System: cp.PairingRoundRobin},
		Rounds:        []cp.RoundData{{Number: 1, Games: []cp.GameData{{WhiteID: "A", BlackID: "B", Result: cp.ResultForfeitBlackWins, IsForfeit: true}}}},
	}
	scores := []cp.PlayerScore{{PlayerID: "A", Score: 0.5}, {PlayerID: "B", Score: 0.5}}
	for _, id := range []string{"wins", "black-wins", "black-games", "koya"} {
		if got := regressionValues(t, id, state, scores)["B"]; got != 1 {
			t.Errorf("%s for round-robin forfeit win = %v, want 1", id, got)
		}
	}
}

func TestC07BPGForfeitWinExcludedInSwiss(t *testing.T) {
	state := &cp.TournamentState{
		Players:       []cp.PlayerEntry{{ID: "A"}, {ID: "B"}},
		PairingConfig: cp.PairingConfig{System: cp.PairingDutch},
		Rounds:        []cp.RoundData{{Number: 1, Games: []cp.GameData{{WhiteID: "A", BlackID: "B", Result: cp.ResultForfeitBlackWins, IsForfeit: true}}}},
	}
	scores := []cp.PlayerScore{{PlayerID: "A"}, {PlayerID: "B"}}
	if got := regressionValues(t, "black-games", state, scores)["B"]; got != 0 {
		t.Errorf("black-games for Swiss forfeit win = %v, want 0", got)
	}
}

func TestC07BottomBoardEliminationOneBoard(t *testing.T) {
	state := &cp.TournamentState{
		Players:       []cp.PlayerEntry{{ID: "a1", TeamID: "A"}, {ID: "b1", TeamID: "B"}},
		ScoringConfig: cp.ScoringConfig{System: cp.ScoringTeam},
		Rounds:        []cp.RoundData{{Number: 1, Matches: []cp.MatchData{{HomeID: "A", AwayID: "B", Boards: []cp.GameData{{WhiteID: "a1", BlackID: "b1", Result: cp.ResultWhiteWins}}}}}},
	}
	scores := []cp.PlayerScore{{PlayerID: "A", Team: &cp.TeamPoints{}}, {PlayerID: "B", Team: &cp.TeamPoints{}}}
	for _, id := range []string{"A", "B"} {
		if got := regressionValues(t, "bottom-board-elimination", state, scores)[id]; got != 0 {
			t.Errorf("BBE with one board for %s = %v, want 0", id, got)
		}
	}
}

func TestC07BPGExcludesPendingAndREPSubtractsWithdrawal(t *testing.T) {
	withdrawn := 1
	state := &cp.TournamentState{
		Players: []cp.PlayerEntry{{ID: "A", WithdrawnAfterRound: &withdrawn}, {ID: "B"}},
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{{WhiteID: "A", BlackID: "B", Result: cp.ResultDraw}}},
			{Number: 2, Games: []cp.GameData{{WhiteID: "B", BlackID: "A", Result: cp.ResultPending}}},
			{Number: 3},
		},
	}
	scores := []cp.PlayerScore{{PlayerID: "A"}, {PlayerID: "B"}}
	if got := regressionValues(t, "black-games", state, scores)["A"]; got != 0 {
		t.Errorf("BPG with pending game = %v, want 0", got)
	}
	if got := regressionValues(t, "rounds-played", state, scores)["A"]; got != 1 {
		t.Errorf("REP after withdrawal = %v, want 1", got)
	}
}
