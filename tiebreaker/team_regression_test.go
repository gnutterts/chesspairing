package tiebreaker

import (
	"context"
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/team"
)

func TestTiedAllForfeitMatchIsPlayed(t *testing.T) {
	st := &cp.TournamentState{
		Players:       []cp.PlayerEntry{{ID: "a1", TeamID: "A"}, {ID: "a2", TeamID: "A"}, {ID: "b1", TeamID: "B"}, {ID: "b2", TeamID: "B"}},
		CurrentRound:  2,
		ScoringConfig: cp.ScoringConfig{System: cp.ScoringTeam},
		Rounds: []cp.RoundData{{Number: 1, Matches: []cp.MatchData{{
			HomeID: "A", AwayID: "B",
			Boards: []cp.GameData{
				{WhiteID: "a1", BlackID: "b1", Result: cp.ResultForfeitWhiteWins, IsForfeit: true},
				{WhiteID: "b2", BlackID: "a2", Result: cp.ResultForfeitWhiteWins, IsForfeit: true},
			},
		}}}},
	}
	scores, err := team.New(team.Options{}).Score(context.Background(), st)
	if err != nil {
		t.Fatal(err)
	}
	records := buildTeamOpponentRecords(st, scores)
	if !records.records["A"][0].Played || !records.records["B"][0].Played {
		t.Errorf("tied all-forfeit match is not played: %+v", records.records)
	}
}

func TestMatchWithOneForfeitBoardIsPlayed(t *testing.T) {
	ps := []cp.PlayerEntry{}
	for _, tm := range []string{"A", "B", "C", "D"} {
		ps = append(ps, cp.PlayerEntry{ID: tm + "1", TeamID: tm}, cp.PlayerEntry{ID: tm + "2", TeamID: tm})
	}
	st := &cp.TournamentState{
		Players: ps, CurrentRound: 3, ScoringConfig: cp.ScoringConfig{System: cp.ScoringTeam},
		PairingConfig: cp.PairingConfig{System: cp.PairingTeam},
		Rounds: []cp.RoundData{
			{Number: 1, Matches: []cp.MatchData{
				{HomeID: "A", AwayID: "B", Boards: []cp.GameData{{WhiteID: "A1", BlackID: "B1", Result: cp.ResultWhiteWins}, {WhiteID: "B2", BlackID: "A2", Result: cp.ResultForfeitBlackWins, IsForfeit: true}}},
				{HomeID: "C", AwayID: "D", Boards: []cp.GameData{{WhiteID: "C1", BlackID: "D1", Result: cp.ResultDraw}, {WhiteID: "D2", BlackID: "C2", Result: cp.ResultDraw}}},
			}},
			{Number: 2, Matches: []cp.MatchData{
				{HomeID: "A", AwayID: "C", Boards: []cp.GameData{{WhiteID: "A1", BlackID: "C1", Result: cp.ResultDraw}, {WhiteID: "C2", BlackID: "A2", Result: cp.ResultDraw}}},
				{HomeID: "B", AwayID: "D", Boards: []cp.GameData{{WhiteID: "B1", BlackID: "D1", Result: cp.ResultWhiteWins}, {WhiteID: "D2", BlackID: "B2", Result: cp.ResultBlackWins}}},
			}},
		},
	}
	sc, err := team.New(team.Options{}).Score(context.Background(), st)
	if err != nil {
		t.Fatal(err)
	}
	tab := buildTeamOpponentRecords(st, sc)
	for _, r := range tab.records["A"] {
		t.Logf("A round %d played=%v cat=%v opp=%s", r.Round, r.Played, r.Category, r.OpponentID)
	}
	if !tab.records["A"][0].Played {
		t.Errorf("A-B round 1 had one game played but is treated as unplayed")
	}
	bh, err := (&TeamBuchholz{}).Compute(context.Background(), st, sc)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range bh {
		if value.PlayerID == "A" && value.Value != 4 {
			t.Errorf("A Buchholz-MP = %v, want 4", value.Value)
		}
	}
}
