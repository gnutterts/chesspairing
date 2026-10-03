package team

import (
	"context"
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func m(h, a string, hg, ag float64) cp.MatchData {
	return cp.MatchData{HomeID: h, AwayID: a, Result: &cp.TeamMatchResult{HomeGame: hg, AwayGame: ag}}
}

// C2: a team that received a PAB recorded in TeamByes must not get it again.
func TestPairingAllocatedByeIsNotRepeated(t *testing.T) {
	players := []cp.PlayerEntry{
		{ID: "A", Rating: 2500}, {ID: "B", Rating: 2400}, {ID: "C", Rating: 2300}, {ID: "D", Rating: 2200}, {ID: "E", Rating: 2100},
	}
	tr := 5
	st := &cp.TournamentState{
		Players: players,
		Rounds: []cp.RoundData{
			{Number: 1, Matches: []cp.MatchData{m("A", "B", 3, 1), m("C", "D", 3, 1)}, TeamByes: []cp.ByeEntry{{PlayerID: "E", Type: cp.ByePAB}}},
			{Number: 2, Matches: []cp.MatchData{m("B", "E", 3, 1), m("A", "C", 2, 2)}, TeamByes: []cp.ByeEntry{{PlayerID: "D", Type: cp.ByePAB}}},
		},
		CurrentRound:  3,
		PairingConfig: cp.PairingConfig{System: cp.PairingTeam},
	}
	res, err := New(Options{TotalRounds: &tr}).Pair(context.Background(), st)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("pairings=%v teamByes=%v byes=%v", res.Pairings, res.TeamByes, res.Byes)
	for _, b := range res.TeamByes {
		if b.PlayerID == "E" || b.PlayerID == "D" {
			t.Errorf("C2 violated: %s gets a second PAB", b.PlayerID)
		}
	}
}

// A forfeit win also prevents a later pairing-allocated bye.
func TestForfeitWinMakesTeamIneligibleForPairingAllocatedBye(t *testing.T) {
	st := &cp.TournamentState{
		Players: []cp.PlayerEntry{{ID: "A"}, {ID: "B"}},
		Rounds: []cp.RoundData{{Number: 1, Matches: []cp.MatchData{{
			HomeID: "A", AwayID: "B",
			Boards: []cp.GameData{{WhiteID: "A", BlackID: "B", Result: cp.ResultForfeitWhiteWins, IsForfeit: true}},
		}}}},
		CurrentRound: 2,
	}
	participants, err := buildParticipantStates(st, "match")
	if err != nil {
		t.Fatal(err)
	}
	for _, participant := range participants {
		if participant.ID == "A" {
			if !participant.PABIneligible.FullPointUnplayed {
				t.Error("forfeit winner remains eligible for a pairing-allocated bye")
			}
			return
		}
	}
	t.Fatal("forfeit winner is not a pairing participant")
}

// Legacy recording through Byes remains accepted for team states.
func TestLegacyByesTeamByeScores(t *testing.T) {
	players := []cp.PlayerEntry{{ID: "A", Rating: 2500}, {ID: "B", Rating: 2400}, {ID: "C", Rating: 2300}}
	st := &cp.TournamentState{
		Players: players,
		Rounds: []cp.RoundData{{
			Number: 1, Games: []cp.GameData{{WhiteID: "A", BlackID: "B", Result: cp.ResultWhiteWins}},
			Byes: []cp.ByeEntry{{PlayerID: "C", Type: cp.ByePAB}},
		}},
		CurrentRound:  2,
		PairingConfig: cp.PairingConfig{System: cp.PairingTeam},
	}
	ps, err := buildParticipantStates(st, "match")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ps {
		t.Logf("%s score=%v", p.ID, p.Score)
		if p.ID == "C" && p.Score == 0 {
			t.Errorf("legacy Byes PAB scored 0 for C")
		}
	}
}
