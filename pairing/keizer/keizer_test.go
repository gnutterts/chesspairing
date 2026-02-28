package keizer

import (
	"context"
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

func TestNew(t *testing.T) {
	p := New(Options{})
	if p == nil {
		t.Fatal("New returned nil")
	}
}

func TestNewFromMap(t *testing.T) {
	p := NewFromMap(map[string]any{
		"allowRepeatPairings":     false,
		"minRoundsBetweenRepeats": 5,
	})
	if p == nil {
		t.Fatal("NewFromMap returned nil")
	}
	if *p.opts.AllowRepeatPairings != false {
		t.Errorf("AllowRepeatPairings = %v, want false", *p.opts.AllowRepeatPairings)
	}
	if *p.opts.MinRoundsBetweenRepeats != 5 {
		t.Errorf("MinRoundsBetweenRepeats = %v, want 5", *p.opts.MinRoundsBetweenRepeats)
	}
}

func TestPairNoPlayers(t *testing.T) {
	p := New(Options{})
	result, err := p.Pair(context.Background(), &chesspairing.TournamentState{
		CurrentRound: 1,
	})
	if err != nil {
		t.Fatalf("Pair error: %v", err)
	}
	if len(result.Pairings) != 0 {
		t.Errorf("expected no pairings, got %d", len(result.Pairings))
	}
}

func TestPairOnePlayer(t *testing.T) {
	p := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
		CurrentRound: 1,
	}
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair error: %v", err)
	}
	if len(result.Pairings) != 0 {
		t.Errorf("expected no pairings with 1 player, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 || result.Byes[0] != "p1" {
		t.Errorf("expected bye for p1, got %v", result.Byes)
	}
}

func TestPairTwoPlayers(t *testing.T) {
	p := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		CurrentRound: 1,
	}
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair error: %v", err)
	}
	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 0 {
		t.Errorf("expected no byes, got %v", result.Byes)
	}
	// Higher rated gets white (first round, no history).
	pair := result.Pairings[0]
	if pair.WhiteID != "p1" || pair.BlackID != "p2" {
		t.Errorf("expected p1 vs p2, got %s vs %s", pair.WhiteID, pair.BlackID)
	}
	if pair.Board != 1 {
		t.Errorf("board = %d, want 1", pair.Board)
	}
}

func TestPairFourPlayersFirstRound(t *testing.T) {
	p := New(Options{})
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
		t.Fatalf("Pair error: %v", err)
	}
	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 0 {
		t.Errorf("expected no byes, got %v", result.Byes)
	}

	// Outside-in pairing: rank 1 (p1) vs rank 4 (p4), rank 2 (p2) vs rank 3 (p3).
	pair1 := result.Pairings[0]
	pair2 := result.Pairings[1]

	if pair1.WhiteID != "p1" || pair1.BlackID != "p4" {
		t.Errorf("board 1: expected p1 vs p4, got %s vs %s", pair1.WhiteID, pair1.BlackID)
	}
	if pair2.WhiteID != "p2" || pair2.BlackID != "p3" {
		t.Errorf("board 2: expected p2 vs p3, got %s vs %s", pair2.WhiteID, pair2.BlackID)
	}
}

func TestPairOddNumberOfPlayers(t *testing.T) {
	p := New(Options{})
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
		t.Fatalf("Pair error: %v", err)
	}
	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 {
		t.Fatalf("expected 1 bye, got %d", len(result.Byes))
	}
	// Middle player (rank 2 = p2) gets the bye.
	if result.Byes[0] != "p2" {
		t.Errorf("bye player = %s, want p2 (middle rank)", result.Byes[0])
	}
	// Remaining: p1 vs p3.
	pair := result.Pairings[0]
	if pair.WhiteID != "p1" || pair.BlackID != "p3" {
		t.Errorf("expected p1 vs p3, got %s vs %s", pair.WhiteID, pair.BlackID)
	}
}

func TestPairColorBalance(t *testing.T) {
	p := New(Options{})
	// Round 1: p1 (white) beat p2 (black).
	// Round 2: p1 should get black (they had white last).
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair error: %v", err)
	}
	pair := result.Pairings[0]
	// p1 had white → should now get black.
	if pair.WhiteID != "p2" || pair.BlackID != "p1" {
		t.Errorf("color balance: expected p2(W) vs p1(B), got %s(W) vs %s(B)",
			pair.WhiteID, pair.BlackID)
	}
}

func TestPairRepeatAvoidance(t *testing.T) {
	// 4 players. Round 1: p1 vs p4, p2 vs p3.
	// Round 2 (currentRound=2, minRoundsBetweenRepeats=3):
	// Default outside-in would pair p1 vs p4 again, but repeat avoidance
	// should swap to avoid it.
	p := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair error: %v", err)
	}
	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}

	// p1 should NOT be paired with p4 again (repeat avoidance).
	for _, pair := range result.Pairings {
		if (pair.WhiteID == "p1" && pair.BlackID == "p4") ||
			(pair.WhiteID == "p4" && pair.BlackID == "p1") {
			t.Errorf("p1 vs p4 repeated — repeat avoidance failed")
		}
	}
}

func TestPairRepeatAllowedAfterGap(t *testing.T) {
	// After enough rounds pass (>= minRoundsBetweenRepeats), repeats are OK.
	p := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
			}},
			{Number: 2, Games: []chesspairing.GameData{}},
			{Number: 3, Games: []chesspairing.GameData{}},
		},
		CurrentRound: 4, // 4 - 1 = 3 rounds gap, >= minRoundsBetweenRepeats
	}
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair error: %v", err)
	}
	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
	// Should pair them — enough gap has passed.
	pair := result.Pairings[0]
	ids := map[string]bool{pair.WhiteID: true, pair.BlackID: true}
	if !ids["p1"] || !ids["p2"] {
		t.Errorf("expected p1 vs p2 (repeat allowed after gap), got %s vs %s",
			pair.WhiteID, pair.BlackID)
	}
}

func TestPairNoRepeatsAllowed(t *testing.T) {
	noRepeat := false
	p := New(Options{AllowRepeatPairings: &noRepeat})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair error: %v", err)
	}

	// p1 can't play p4 (already played), p2 can't play p3 (already played).
	// Should pair: p1 vs p3, p2 vs p4 (or swapped colors).
	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}

	for _, pair := range result.Pairings {
		if (pair.WhiteID == "p1" && pair.BlackID == "p4") ||
			(pair.WhiteID == "p4" && pair.BlackID == "p1") {
			t.Error("p1 vs p4 repeated with noRepeats=true")
		}
		if (pair.WhiteID == "p2" && pair.BlackID == "p3") ||
			(pair.WhiteID == "p3" && pair.BlackID == "p2") {
			t.Error("p2 vs p3 repeated with noRepeats=true")
		}
	}
}

func TestPairInactivePlayers(t *testing.T) {
	p := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: false}, // withdrawn
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		CurrentRound: 1,
	}
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair error: %v", err)
	}
	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
	// Only p1 and p3 should be paired.
	pair := result.Pairings[0]
	ids := map[string]bool{pair.WhiteID: true, pair.BlackID: true}
	if ids["p2"] {
		t.Error("inactive player p2 should not be paired")
	}
}

func TestPairBoardNumbering(t *testing.T) {
	p := New(Options{})
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
		t.Fatalf("Pair error: %v", err)
	}
	for i, pair := range result.Pairings {
		if pair.Board != i+1 {
			t.Errorf("pairing %d board = %d, want %d", i, pair.Board, i+1)
		}
	}
}

func TestPairSortedByScore(t *testing.T) {
	// After round 1, players are sorted by score for round 2 pairing.
	// p4 wins and p1 loses → p4 should rank higher.
	p := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p4", BlackID: "p1", Result: chesspairing.ResultWhiteWins}, // p4 wins
					{WhiteID: "p3", BlackID: "p2", Result: chesspairing.ResultWhiteWins}, // p3 wins
				},
			},
		},
		CurrentRound: 2,
	}
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair error: %v", err)
	}

	// After round 1: p4=1.0, p3=1.0, p1=0, p2=0.
	// Ties broken by rating: p3(1600) < p4(1400) → actually p3 > p4 by rating at equal score? No:
	// p4=1.0, p3=1.0 tie → broken by rating: p3(1600) > p4(1400) → p3 ranks first among 1.0 group.
	// Wait: p4 and p3 both have 1.0. Tiebreak by rating: p3(1600) > p4(1400) → ranking: p3, p4, p1, p2.
	// Outside-in: p3 vs p2 (board 1), p4 vs p1 (board 2).
	// But repeat avoidance may kick in since p3 played p2 and p4 played p1 in round 1.
	// Default minRoundsBetweenRepeats=3, currentRound=2, lastPlayed=1 → gap=1 < 3 → can't repeat.
	// So: p3 can't play p2, p4 can't play p1 → swaps needed.
	// Result should be: p3 vs p1, p4 vs p2 (swapped).

	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}

	// Verify no repeats.
	for _, pair := range result.Pairings {
		if (pair.WhiteID == "p4" && pair.BlackID == "p1") ||
			(pair.WhiteID == "p1" && pair.BlackID == "p4") {
			t.Error("p4 vs p1 repeated")
		}
		if (pair.WhiteID == "p3" && pair.BlackID == "p2") ||
			(pair.WhiteID == "p2" && pair.BlackID == "p3") {
			t.Error("p3 vs p2 repeated")
		}
	}
}

// Options tests

func TestOptionsWithDefaults(t *testing.T) {
	o := Options{}.WithDefaults()
	if *o.AllowRepeatPairings != true {
		t.Errorf("AllowRepeatPairings = %v, want true", *o.AllowRepeatPairings)
	}
	if *o.MinRoundsBetweenRepeats != 3 {
		t.Errorf("MinRoundsBetweenRepeats = %v, want 3", *o.MinRoundsBetweenRepeats)
	}
}

func TestOptionsWithDefaultsPreservesExplicit(t *testing.T) {
	noRepeat := false
	gap := 5
	o := Options{
		AllowRepeatPairings:     &noRepeat,
		MinRoundsBetweenRepeats: &gap,
	}.WithDefaults()
	if *o.AllowRepeatPairings != false {
		t.Errorf("AllowRepeatPairings = %v, want false", *o.AllowRepeatPairings)
	}
	if *o.MinRoundsBetweenRepeats != 5 {
		t.Errorf("MinRoundsBetweenRepeats = %v, want 5", *o.MinRoundsBetweenRepeats)
	}
}

func TestParseOptions(t *testing.T) {
	m := map[string]any{
		"allowRepeatPairings":     false,
		"minRoundsBetweenRepeats": 7,
		"unknownField":            "ignored",
	}
	o := ParseOptions(m)
	if o.AllowRepeatPairings == nil || *o.AllowRepeatPairings != false {
		t.Errorf("AllowRepeatPairings = %v, want false", o.AllowRepeatPairings)
	}
	if o.MinRoundsBetweenRepeats == nil || *o.MinRoundsBetweenRepeats != 7 {
		t.Errorf("MinRoundsBetweenRepeats = %v, want 7", o.MinRoundsBetweenRepeats)
	}
}
