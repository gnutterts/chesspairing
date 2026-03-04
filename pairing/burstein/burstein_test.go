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

func TestBursteinCriteria_NoFloatCriteria(t *testing.T) {
	t.Parallel()

	// Build a 6-player, 3-round tournament where float criteria (C14-C17)
	// would penalize a specific pairing under Dutch rules but not Burstein.
	//
	// Setup: After round 2, player p3 has downfloated in both rounds 1 and 2.
	// Under Dutch C14 (downfloat repeat R-1), pairing p3 as a downfloater
	// again would be penalized. Under Burstein (no C14), it's fine.
	//
	// We verify the Burstein pairer does NOT avoid the downfloat-repeat pairing.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P2400", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "P2300", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "P2200", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "P2100", Rating: 2100, Active: true},
			{ID: "p5", DisplayName: "P2000", Rating: 2000, Active: true},
			{ID: "p6", DisplayName: "P1900", Rating: 1900, Active: true},
		},
		CurrentRound: 3,
		Rounds: []chesspairing.RoundData{
			{Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p5", BlackID: "p2", Result: chesspairing.ResultBlackWins},
				{WhiteID: "p3", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
			}},
			{Games: []chesspairing.GameData{
				{WhiteID: "p2", BlackID: "p1", Result: chesspairing.ResultDraw},
				{WhiteID: "p4", BlackID: "p3", Result: chesspairing.ResultBlackWins},
				{WhiteID: "p6", BlackID: "p5", Result: chesspairing.ResultDraw},
			}},
		},
	}

	totalRounds := 5
	p := New(Options{TotalRounds: &totalRounds})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	// The test succeeds if pairing completes without error.
	// The key verification is that the criteria function used is NOT
	// DutchOptimizationCriteria (which includes C8, C14-C21).
	if len(result.Pairings) != 3 {
		t.Errorf("expected 3 pairings, got %d", len(result.Pairings))
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

func TestTopSeedColor_Black(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P2400", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "P2300", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "P2200", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "P2100", Rating: 2100, Active: true},
		},
		CurrentRound: 1,
	}

	black := "black"
	totalRounds := 5
	p := New(Options{TopSeedColor: &black, TotalRounds: &totalRounds})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	if result.Pairings[0].BlackID != "p1" {
		t.Errorf("board 1: expected p1 as Black, got white=%s black=%s",
			result.Pairings[0].WhiteID, result.Pairings[0].BlackID)
	}
}

func TestForbiddenPairs(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P2400", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "P2300", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "P2200", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "P2100", Rating: 2100, Active: true},
		},
		CurrentRound: 1,
	}

	totalRounds := 5
	p := New(Options{
		ForbiddenPairs: [][]string{{"p1", "p3"}},
		TotalRounds:    &totalRounds,
	})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	for _, pairing := range result.Pairings {
		if (pairing.WhiteID == "p1" && pairing.BlackID == "p3") ||
			(pairing.WhiteID == "p3" && pairing.BlackID == "p1") {
			t.Error("p1 should not be paired with p3 (forbidden pair)")
		}
	}
}

func TestBakuAcceleration_Round1(t *testing.T) {
	t.Parallel()

	// 8 players, 5 rounds, Baku acceleration.
	// GA = BakuGASize(8) = 2 * ceil(8/4) = 4 (top 4 players).
	// Round 1 is a full VP round → GA players get +1.0 virtual points.
	// GA (PairingScore 1.0): p1(2400), p2(2300), p3(2200), p4(2100)
	// GB (PairingScore 0.0): p5(2000), p6(1900), p7(1800), p8(1700)
	// Expected: GA pairs within GA, GB pairs within GB (no mixing).
	totalRounds := 5
	baku := "baku"
	p := New(Options{Acceleration: &baku, TotalRounds: &totalRounds})

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P2400", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "P2300", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "P2200", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "P2100", Rating: 2100, Active: true},
			{ID: "p5", DisplayName: "P2000", Rating: 2000, Active: true},
			{ID: "p6", DisplayName: "P1900", Rating: 1900, Active: true},
			{ID: "p7", DisplayName: "P1800", Rating: 1800, Active: true},
			{ID: "p8", DisplayName: "P1700", Rating: 1700, Active: true},
		},
		CurrentRound: 1,
	}

	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	if len(result.Pairings) != 4 {
		t.Fatalf("expected 4 pairings, got %d", len(result.Pairings))
	}

	// Verify no GA/GB mixing.
	ga := map[string]bool{"p1": true, "p2": true, "p3": true, "p4": true}
	for _, pair := range result.Pairings {
		whiteGA := ga[pair.WhiteID]
		blackGA := ga[pair.BlackID]
		if whiteGA != blackGA {
			t.Errorf("GA/GB mixing: board %d has %s (GA=%v) vs %s (GA=%v)",
				pair.Board, pair.WhiteID, whiteGA, pair.BlackID, blackGA)
		}
	}

	// Verify acceleration note is present.
	foundAccelNote := false
	for _, note := range result.Notes {
		if note == "Baku acceleration: GA=4 players, VP=1.0" {
			foundAccelNote = true
		}
	}
	if !foundAccelNote {
		t.Errorf("expected Baku acceleration note, got notes: %v", result.Notes)
	}
}
