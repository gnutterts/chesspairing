package burstein

import (
	"context"
	"errors"
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

func TestPair_SeedingRound(t *testing.T) {
	t.Parallel()

	// 4 players, round 1 of 9 (seeding round).
	totalRounds := 9
	p := New(Options{TotalRounds: &totalRounds})

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		CurrentRound: 1,
	}

	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	if len(result.Pairings) != 2 {
		t.Errorf("expected 2 pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 0 {
		t.Errorf("expected 0 byes, got %d", len(result.Byes))
	}

	// Verify all 4 players are paired.
	paired := make(map[string]bool)
	for _, pair := range result.Pairings {
		paired[pair.WhiteID] = true
		paired[pair.BlackID] = true
	}
	for _, id := range []string{"p1", "p2", "p3", "p4"} {
		if !paired[id] {
			t.Errorf("player %s not paired", id)
		}
	}

	// Check seeding round note.
	foundSeedingNote := false
	for _, note := range result.Notes {
		if note == "Seeding round 1 of 4" {
			foundSeedingNote = true
		}
	}
	if !foundSeedingNote {
		t.Errorf("expected seeding round note, got notes: %v", result.Notes)
	}
}

func TestPair_PostSeedingRound(t *testing.T) {
	t.Parallel()

	// 6 players, round 5 of 9 (post-seeding). Uses a 1-factorization of K_6
	// so after 4 rounds each player has played exactly 4 of 5 opponents,
	// leaving exactly one valid perfect matching for round 5.
	//
	// R1: p1-p6, p2-p5, p3-p4
	// R2: p2-p6, p3-p1, p4-p5
	// R3: p3-p6, p4-p2, p5-p1
	// R4: p4-p6, p5-p3, p1-p2
	// Remaining for R5: p5-p6, p1-p4, p2-p3
	totalRounds := 9
	p := New(Options{TotalRounds: &totalRounds})

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
			{ID: "p5", DisplayName: "Eve", Rating: 1200, Active: true},
			{ID: "p6", DisplayName: "Frank", Rating: 1000, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p1", Result: chesspairing.ResultBlackWins},
					{WhiteID: "p4", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p3", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultBlackWins},
					{WhiteID: "p5", BlackID: "p1", Result: chesspairing.ResultBlackWins},
				},
			},
			{
				Number: 4,
				Games: []chesspairing.GameData{
					{WhiteID: "p4", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p5", BlackID: "p3", Result: chesspairing.ResultBlackWins},
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
				},
			},
		},
		CurrentRound: 5,
	}

	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	if len(result.Pairings) != 3 {
		t.Errorf("expected 3 pairings, got %d", len(result.Pairings))
	}

	// Check post-seeding note.
	foundPostSeedingNote := false
	for _, note := range result.Notes {
		if note == "Post-seeding round 5 (opposition index ranking)" {
			foundPostSeedingNote = true
		}
	}
	if !foundPostSeedingNote {
		t.Errorf("expected post-seeding note, got notes: %v", result.Notes)
	}
}

func TestPair_SinglePlayer(t *testing.T) {
	t.Parallel()

	p := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
		CurrentRound: 1,
	}

	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	if len(result.Pairings) != 0 {
		t.Errorf("expected 0 pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 || result.Byes[0] != "p1" {
		t.Errorf("expected bye for p1, got %v", result.Byes)
	}
}

func TestPair_NoPlayers(t *testing.T) {
	t.Parallel()

	p := New(Options{})
	state := &chesspairing.TournamentState{
		CurrentRound: 1,
	}

	_, err := p.Pair(context.Background(), state)
	if err == nil {
		t.Fatal("expected error for no players")
	}
	if !errors.Is(err, ErrTooFewPlayers) {
		t.Errorf("expected ErrTooFewPlayers, got %v", err)
	}
}

func TestPair_OddPlayers_Bye(t *testing.T) {
	t.Parallel()

	totalRounds := 9
	p := New(Options{TotalRounds: &totalRounds})

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		CurrentRound: 1,
	}

	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	if len(result.Pairings) != 1 {
		t.Errorf("expected 1 pairing, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 {
		t.Errorf("expected 1 bye, got %d", len(result.Byes))
	}

	// The bye should go to the lowest-scored player with most games played
	// (Burstein rule). In round 1, all have 0 score and 0 games,
	// so the lowest ranking (highest TPN = p3) gets the bye.
	if result.Byes[0] != "p3" {
		t.Errorf("expected p3 to get bye, got %s", result.Byes[0])
	}
}

func TestPair_BursteinNote(t *testing.T) {
	t.Parallel()

	totalRounds := 9
	p := New(Options{TotalRounds: &totalRounds})

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		CurrentRound: 1,
	}

	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	foundSystemNote := false
	for _, note := range result.Notes {
		if note == "Pairings generated by Burstein Swiss system (C.04.4.2)" {
			foundSystemNote = true
		}
	}
	if !foundSystemNote {
		t.Errorf("expected Burstein system note, got notes: %v", result.Notes)
	}
}

func TestPair_BoardNumbers(t *testing.T) {
	t.Parallel()

	totalRounds := 9
	p := New(Options{TotalRounds: &totalRounds})

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		CurrentRound: 1,
	}

	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	for i, pair := range result.Pairings {
		if pair.Board != i+1 {
			t.Errorf("pair %d: board=%d, want %d", i, pair.Board, i+1)
		}
	}
}
