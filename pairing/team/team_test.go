// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestPair_PlayerEntryTeamID(t *testing.T) {
	// Finding 2: "test with a TeamID state of 2 teams x 4 players asserting Pair returns the pairing and no error"
	state := &chesspairing.TournamentState{
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", TeamID: "t1"},
			{ID: "p2", TeamID: "t1"},
			{ID: "p3", TeamID: "t1"},
			{ID: "p4", TeamID: "t1"},
			{ID: "p5", TeamID: "t2"},
			{ID: "p6", TeamID: "t2"},
			{ID: "p7", TeamID: "t2"},
			{ID: "p8", TeamID: "t2"},
		},
	}
	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair returned error: %v", err)
	}
	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
	pair := result.Pairings[0]
	if (pair.WhiteID != "t1" || pair.BlackID != "t2") && (pair.WhiteID != "t2" || pair.BlackID != "t1") {
		t.Errorf("expected pairing between t1 and t2, got %s-%s", pair.WhiteID, pair.BlackID)
	}
}

func TestPair_Round1_FourTeams(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400},
			{ID: "t2", DisplayName: "Beta", Rating: 2300},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200},
			{ID: "t4", DisplayName: "Delta", Rating: 2100},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 2 {
		t.Errorf("expected 2 pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 0 {
		t.Errorf("expected 0 byes, got %d", len(result.Byes))
	}

	// The first Article 3.6 identifier is (1,2,3,4): t1-t3, t2-t4.
	checkPairing(t, result.Pairings, "t1", "t3")
	checkPairing(t, result.Pairings, "t2", "t4")
}

func TestPair_Round1_IdentifierOrder(t *testing.T) {
	for _, teamCount := range []int{6, 8} {
		teams := make([]chesspairing.PlayerEntry, teamCount)
		for i := range teams {
			teams[i] = chesspairing.PlayerEntry{ID: "t" + strconv.Itoa(i+1), Rating: 3000 - i}
		}
		state := &chesspairing.TournamentState{
			Players:      teams,
			CurrentRound: 1,
			PairingConfig: chesspairing.PairingConfig{
				System: chesspairing.PairingTeam,
			},
		}
		result, err := New(Options{}).Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("Pair() with %d teams error: %v", teamCount, err)
		}
		for i := 1; i <= teamCount/2; i++ {
			checkPairing(t, result.Pairings, "t"+strconv.Itoa(i), "t"+strconv.Itoa(i+teamCount/2))
		}
	}
}

func TestPair_AvoidsPlayedFirstIdentifierPair(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", Rating: 2600},
			{ID: "t2", Rating: 2500},
			{ID: "t3", Rating: 2400},
			{ID: "t4", Rating: 2300},
			{ID: "t5", Rating: 2200},
			{ID: "t6", Rating: 2100},
		},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games:  []chesspairing.GameData{{WhiteID: "t1", BlackID: "t4", Result: chesspairing.ResultDraw}},
			TeamByes: []chesspairing.ByeEntry{
				{PlayerID: "t2", Type: chesspairing.ByeHalf},
				{PlayerID: "t3", Type: chesspairing.ByeHalf},
				{PlayerID: "t5", Type: chesspairing.ByeHalf},
				{PlayerID: "t6", Type: chesspairing.ByeHalf},
			},
		}},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}
	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	checkPairing(t, result.Pairings, "t1", "t5")
	checkPairing(t, result.Pairings, "t2", "t4")
	checkPairing(t, result.Pairings, "t3", "t6")
}

func TestPair_Round1_FiveTeams_PAB(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400},
			{ID: "t2", DisplayName: "Beta", Rating: 2300},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200},
			{ID: "t4", DisplayName: "Delta", Rating: 2100},
			{ID: "t5", DisplayName: "Epsilon", Rating: 2000},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 2 {
		t.Errorf("expected 2 pairings, got %d", len(result.Pairings))
	}
	if len(result.TeamByes) != 1 {
		t.Errorf("expected 1 bye, got %d", len(result.TeamByes))
	} else {
		// PAB to team with lowest score (all 0), most matches (all 0), largest TPN (t5).
		if result.TeamByes[0].PlayerID != "t5" {
			t.Errorf("expected t5 to get bye, got %s", result.TeamByes[0].PlayerID)
		}
		if result.TeamByes[0].Type != chesspairing.ByePAB {
			t.Errorf("expected ByePAB, got %v", result.TeamByes[0].Type)
		}
	}
}

func TestPair_NoPABCandidate(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "t1"}, {ID: "t2"}, {ID: "t3"}, {ID: "t4"}, {ID: "t5"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Byes: []chesspairing.ByeEntry{
				{PlayerID: "t1", Type: chesspairing.ByeFullPoint},
				{PlayerID: "t2", Type: chesspairing.ByeFullPoint},
				{PlayerID: "t3", Type: chesspairing.ByeFullPoint},
				{PlayerID: "t4", Type: chesspairing.ByeFullPoint},
				{PlayerID: "t5", Type: chesspairing.ByeFullPoint},
			},
		}},
		CurrentRound: 2,
	}
	_, err := New(Options{}).Pair(context.Background(), state)
	if !errors.Is(err, swisslib.ErrNoPABCandidate) {
		t.Fatalf("Pair() error = %v, want ErrNoPABCandidate", err)
	}
}

func TestPair_Round2_WithHistory(t *testing.T) {
	// After round 1: t1 beats t2 (score 2 match pts), t3 beats t4.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400},
			{ID: "t2", DisplayName: "Beta", Rating: 2300},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200},
			{ID: "t4", DisplayName: "Delta", Rating: 2100},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "t1", BlackID: "t2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "t3", BlackID: "t4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 2 {
		t.Errorf("expected 2 pairings, got %d", len(result.Pairings))
	}

	// Each two-team score group has only one legal pairing.
	checkPairing(t, result.Pairings, "t1", "t3")
	checkPairing(t, result.Pairings, "t2", "t4")
}

func TestPair_NoRepeatPairing(t *testing.T) {
	// t1 already played t2 in round 1. They should not be paired again.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400},
			{ID: "t2", DisplayName: "Beta", Rating: 2300},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200},
			{ID: "t4", DisplayName: "Delta", Rating: 2100},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "t1", BlackID: "t2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "t3", BlackID: "t4", Result: chesspairing.ResultDraw},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	for _, p := range result.Pairings {
		if (p.WhiteID == "t1" && p.BlackID == "t2") || (p.WhiteID == "t2" && p.BlackID == "t1") {
			t.Error("t1 should not be paired with t2 again")
		}
	}
}

func TestPair_NoActivePlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:      nil,
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 0 {
		t.Errorf("expected 0 pairings, got %d", len(result.Pairings))
	}
}

func TestPair_SingleTeam(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 0 {
		t.Errorf("expected 0 pairings, got %d", len(result.Pairings))
	}
	if len(result.TeamByes) != 1 || result.TeamByes[0].PlayerID != "t1" {
		t.Error("single team should get PAB")
	}
}

func TestPair_BoardOrdering(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400},
			{ID: "t2", DisplayName: "Beta", Rating: 2300},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200},
			{ID: "t4", DisplayName: "Delta", Rating: 2100},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "t1", BlackID: "t2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "t3", BlackID: "t4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	// Board 1 should have the highest-scoring pair (t1 vs t3, both 1.0).
	if result.Pairings[0].Board != 1 {
		t.Errorf("board ordering: expected Board=1, got %d", result.Pairings[0].Board)
	}
}

func TestPair_ForbiddenPairs(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400},
			{ID: "t2", DisplayName: "Beta", Rating: 2300},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200},
			{ID: "t4", DisplayName: "Delta", Rating: 2100},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	// Forbid t1 vs t2.
	pairer := New(Options{
		ForbiddenPairs: [][]string{{"t1", "t2"}},
	})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	for _, p := range result.Pairings {
		if (p.WhiteID == "t1" && p.BlackID == "t2") || (p.WhiteID == "t2" && p.BlackID == "t1") {
			t.Error("t1 and t2 should not be paired (forbidden)")
		}
	}
}

func TestPair_ColorPreferenceTypeB(t *testing.T) {
	cpType := "B"
	pairer := New(Options{
		ColorPreferenceType: &cpType,
	})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400},
			{ID: "t2", DisplayName: "Beta", Rating: 2300},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 1 {
		t.Errorf("expected 1 pairing, got %d", len(result.Pairings))
	}
}

func TestPair_ColorPreferenceNone(t *testing.T) {
	cpType := "none"
	pairer := New(Options{
		ColorPreferenceType: &cpType,
	})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400},
			{ID: "t2", DisplayName: "Beta", Rating: 2300},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 1 {
		t.Errorf("expected 1 pairing, got %d", len(result.Pairings))
	}
}

// checkPairing verifies that a pairing exists between two teams (in either colour order).
func checkPairing(t *testing.T, pairings []chesspairing.GamePairing, id1, id2 string) {
	t.Helper()
	for _, p := range pairings {
		if (p.WhiteID == id1 && p.BlackID == id2) || (p.WhiteID == id2 && p.BlackID == id1) {
			return
		}
	}
	t.Errorf("expected pairing between %s and %s", id1, id2)
}

func TestPair_ContextCancelled(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", Rating: 2400},
			{ID: "p2", Rating: 2300},
		},
		CurrentRound: 1,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pairer := New(Options{})
	res, err := pairer.Pair(ctx, state)
	if err == nil || (err.Error() != context.Canceled.Error() && !errors.Is(err, context.Canceled)) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
	if res != nil {
		t.Errorf("expected nil result, got %v", res)
	}
}

func TestTeamScores_GamesAndMatches(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", TeamID: "t1", Rating: 2000},
			{ID: "p2", TeamID: "t2", Rating: 1900},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 2,
				Matches: []chesspairing.MatchData{
					{
						HomeID: "t2",
						AwayID: "t1",
						Result: &chesspairing.TeamMatchResult{HomeGame: 1, AwayGame: 0},
					},
				},
			},
		},
		CurrentRound: 3,
		ScoringConfig: chesspairing.ScoringConfig{
			System: chesspairing.ScoringTeam,
		},
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	primary := "match"
	participants, err := buildParticipantStates(state, primary)
	if err != nil {
		t.Fatalf("buildParticipantStates: %v", err)
	}

	for _, p := range participants {
		if p.Score != 2.0 {
			t.Errorf("team %s score = %v, want 2.0", p.ID, p.Score)
		}
	}
}

func TestTeamByes_ScoredCorrectly(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", TeamID: "t1"},
			{ID: "p2", TeamID: "t2"},
			{ID: "p3", TeamID: "t3"},
		},
		CurrentRound: 1,
		ScoringConfig: chesspairing.ScoringConfig{
			System: chesspairing.ScoringTeam,
		},
		PairingConfig: chesspairing.PairingConfig{
			System:  chesspairing.PairingTeam,
			Options: map[string]any{"primaryScore": "match"},
		},
	}
	pairer := New(Options{PrimaryScore: ptr("match")})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair: %v", err)
	}

	state.Rounds = append(state.Rounds, chesspairing.RoundData{
		Number:   1,
		TeamByes: result.TeamByes,
	})
	state.CurrentRound = 2

	participants, err := buildParticipantStates(state, "match")
	if err != nil {
		t.Fatalf("buildParticipantStates: %v", err)
	}

	var byedTeam string
	if len(result.TeamByes) > 0 {
		byedTeam = result.TeamByes[0].PlayerID
	}
	if byedTeam == "" {
		t.Fatalf("no team bye generated")
	}

	var score float64
	for _, p := range participants {
		if p.ID == byedTeam {
			score = p.Score
		}
	}
	// PAB should give MP 1.
	if score != 1.0 {
		t.Errorf("scored %v, want MP: 1", score)
	}
}
