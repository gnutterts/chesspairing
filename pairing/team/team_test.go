package team

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPair_Round1_FourTeams(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400, Active: true},
			{ID: "t2", DisplayName: "Beta", Rating: 2300, Active: true},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200, Active: true},
			{ID: "t4", DisplayName: "Delta", Rating: 2100, Active: true},
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

	// Round 1 lexicographic: t1 vs t2, t3 vs t4.
	checkPairing(t, result.Pairings, "t1", "t2")
	checkPairing(t, result.Pairings, "t3", "t4")
}

func TestPair_Round1_FiveTeams_PAB(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400, Active: true},
			{ID: "t2", DisplayName: "Beta", Rating: 2300, Active: true},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200, Active: true},
			{ID: "t4", DisplayName: "Delta", Rating: 2100, Active: true},
			{ID: "t5", DisplayName: "Epsilon", Rating: 2000, Active: true},
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
	if len(result.Byes) != 1 {
		t.Errorf("expected 1 bye, got %d", len(result.Byes))
	}
	// PAB to team with lowest score (all 0), most matches (all 0), largest TPN (t5).
	if result.Byes[0].PlayerID != "t5" {
		t.Errorf("expected t5 to get bye, got %s", result.Byes[0].PlayerID)
	}
	if result.Byes[0].Type != chesspairing.ByePAB {
		t.Errorf("expected ByePAB, got %v", result.Byes[0].Type)
	}
}

func TestPair_Round2_WithHistory(t *testing.T) {
	// After round 1: t1 beats t2 (score 2 match pts), t3 beats t4.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400, Active: true},
			{ID: "t2", DisplayName: "Beta", Rating: 2300, Active: true},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200, Active: true},
			{ID: "t4", DisplayName: "Delta", Rating: 2100, Active: true},
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

	// Scoregroup 1.0: t1, t3 → lexicographic: t1 vs t3.
	// Scoregroup 0.0: t2, t4 → lexicographic: t2 vs t4.
	checkPairing(t, result.Pairings, "t1", "t3")
	checkPairing(t, result.Pairings, "t2", "t4")
}

func TestPair_NoRepeatPairing(t *testing.T) {
	// t1 already played t2 in round 1. They should not be paired again.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400, Active: true},
			{ID: "t2", DisplayName: "Beta", Rating: 2300, Active: true},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200, Active: true},
			{ID: "t4", DisplayName: "Delta", Rating: 2100, Active: true},
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
			{ID: "t1", DisplayName: "Alpha", Rating: 2400, Active: true},
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
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "t1" {
		t.Error("single team should get PAB")
	}
}

func TestPair_BoardOrdering(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", DisplayName: "Alpha", Rating: 2400, Active: true},
			{ID: "t2", DisplayName: "Beta", Rating: 2300, Active: true},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200, Active: true},
			{ID: "t4", DisplayName: "Delta", Rating: 2100, Active: true},
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
			{ID: "t1", DisplayName: "Alpha", Rating: 2400, Active: true},
			{ID: "t2", DisplayName: "Beta", Rating: 2300, Active: true},
			{ID: "t3", DisplayName: "Gamma", Rating: 2200, Active: true},
			{ID: "t4", DisplayName: "Delta", Rating: 2100, Active: true},
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
			{ID: "t1", DisplayName: "Alpha", Rating: 2400, Active: true},
			{ID: "t2", DisplayName: "Beta", Rating: 2300, Active: true},
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
			{ID: "t1", DisplayName: "Alpha", Rating: 2400, Active: true},
			{ID: "t2", DisplayName: "Beta", Rating: 2300, Active: true},
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
