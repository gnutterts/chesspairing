package doubleswiss

import (
	"context"
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

func TestPair_Round1_FourPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 2100, Active: true},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
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

	// Round 1 lexicographic: p1 vs p2, p3 vs p4.
	checkPairing(t, result.Pairings, "p1", "p2")
	checkPairing(t, result.Pairings, "p3", "p4")
}

func TestPair_Round1_FivePlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 2100, Active: true},
			{ID: "p5", DisplayName: "Eve", Rating: 2000, Active: true},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
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
	// PAB to lowest-ranked (highest TPN = p5).
	if result.Byes[0].PlayerID != "p5" {
		t.Errorf("expected p5 to get bye, got %s", result.Byes[0].PlayerID)
	}
	if result.Byes[0].Type != chesspairing.ByePAB {
		t.Errorf("expected ByePAB, got %v", result.Byes[0].Type)
	}
}

func TestPair_Round2_WithHistory(t *testing.T) {
	// After round 1: p1 beats p2 (score 1), p3 beats p4 (score 1).
	// Scores: p1=1, p3=1, p2=0, p4=0.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 2100, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
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

	// Scoregroup 1.0: p1, p3 → lexicographic p1 vs p3.
	// Scoregroup 0.0: p2, p4 → lexicographic p2 vs p4.
	checkPairing(t, result.Pairings, "p1", "p3")
	checkPairing(t, result.Pairings, "p2", "p4")
}

func TestPair_Round2_AvoidRepeat(t *testing.T) {
	// After round 1: p1 beats p2, p3 draws p4.
	// Scores: p1=1, p2=0, p3=0.5, p4=0.5.
	// Scoregroups: [1.0: p1], [0.5: p3, p4], [0.0: p2].
	// p1 alone → needs upfloater from [0.5]. Lexicographic: p4 floats up.
	// Bracket [1.0: p1, p4]: p1 vs p4.
	// Bracket [0.5: p3] + [0.0: p2] → p3 vs p2.
	// But wait — p1 hasn't played p4, and p3 hasn't played p2. Both are fine.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 2100, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
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

	// p1 should not be paired with p2 again (C1: no repeat).
	for _, p := range result.Pairings {
		if (p.WhiteID == "p1" && p.BlackID == "p2") || (p.WhiteID == "p2" && p.BlackID == "p1") {
			t.Error("p1 should not be paired with p2 again")
		}
	}
}

func TestPair_NoActivePlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:      nil,
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
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

func TestPair_SinglePlayer(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
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
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p1" {
		t.Error("single player should get PAB")
	}
}

func TestPair_BoardOrdering(t *testing.T) {
	// Board ordering: max score of pair desc, then min TPN of pair asc.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 2100, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	// Board 1 should have the highest-scoring pair (p1 vs p3, both 1.0).
	if result.Pairings[0].Board != 1 {
		t.Errorf("board 1: expected Board=1, got %d", result.Pairings[0].Board)
	}
}

// checkPairing verifies that a pairing exists between two participants (in either colour order).
func checkPairing(t *testing.T, pairings []chesspairing.GamePairing, id1, id2 string) {
	t.Helper()
	for _, p := range pairings {
		if (p.WhiteID == id1 && p.BlackID == id2) || (p.WhiteID == id2 && p.BlackID == id1) {
			return
		}
	}
	t.Errorf("expected pairing between %s and %s", id1, id2)
}
