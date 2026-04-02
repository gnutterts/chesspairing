package roundrobin

import (
	"context"
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestNew(t *testing.T) {
	p := New(Options{})
	if p == nil {
		t.Fatal("New returned nil")
	}
}

func TestNewFromMap(t *testing.T) {
	p := NewFromMap(map[string]any{
		"cycles":       2,
		"colorBalance": false,
	})
	if p == nil {
		t.Fatal("NewFromMap returned nil")
	}
	if *p.opts.Cycles != 2 {
		t.Errorf("Cycles = %v, want 2", *p.opts.Cycles)
	}
	if *p.opts.ColorBalance != false {
		t.Errorf("ColorBalance = %v, want false", *p.opts.ColorBalance)
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
		t.Errorf("expected no pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p1" {
		t.Errorf("expected bye for p1, got %v", result.Byes)
	}
}

func TestPairTwoPlayersRound1(t *testing.T) {
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
		t.Errorf("expected no byes with 2 players, got %v", result.Byes)
	}
	// Board 1.
	if result.Pairings[0].Board != 1 {
		t.Errorf("board = %d, want 1", result.Pairings[0].Board)
	}
}

func TestPairFourPlayersAllRounds(t *testing.T) {
	// 4 players → 3 rounds (single RR). Every player plays every other exactly once.
	p := New(Options{})
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
	}

	// Track all pairings across rounds.
	pairSet := make(map[string]bool)
	for round := 1; round <= 3; round++ {
		state := &chesspairing.TournamentState{
			Players:      players,
			CurrentRound: round,
		}
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair error: %v", round, err)
		}
		if len(result.Pairings) != 2 {
			t.Fatalf("round %d: expected 2 pairings, got %d", round, len(result.Pairings))
		}
		if len(result.Byes) != 0 {
			t.Errorf("round %d: expected no byes, got %v", round, result.Byes)
		}
		for _, pair := range result.Pairings {
			key := pairKey(pair.WhiteID, pair.BlackID)
			if pairSet[key] {
				t.Errorf("round %d: duplicate pairing %s", round, key)
			}
			pairSet[key] = true
		}
	}

	// Verify all 6 pairs played (4 choose 2 = 6).
	if len(pairSet) != 6 {
		t.Errorf("expected 6 unique pairings, got %d", len(pairSet))
	}
}

func TestPairOddPlayers(t *testing.T) {
	// 3 players → 3 rounds. One bye per round, each pair plays once.
	p := New(Options{})
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
	}

	// n=3, table size=4 (with dummy), rounds per cycle = 3.
	pairSet := make(map[string]bool)
	byeSet := make(map[string]int) // player → bye count
	for round := 1; round <= 3; round++ {
		state := &chesspairing.TournamentState{
			Players:      players,
			CurrentRound: round,
		}
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair error: %v", round, err)
		}
		if len(result.Pairings) != 1 {
			t.Fatalf("round %d: expected 1 pairing, got %d", round, len(result.Pairings))
		}
		if len(result.Byes) != 1 {
			t.Fatalf("round %d: expected 1 bye, got %d", round, len(result.Byes))
		}
		byeSet[result.Byes[0].PlayerID]++
		for _, pair := range result.Pairings {
			key := pairKey(pair.WhiteID, pair.BlackID)
			pairSet[key] = true
		}
	}

	// All 3 unique pairs should be covered.
	if len(pairSet) != 3 {
		t.Errorf("expected 3 unique pairings, got %d", len(pairSet))
	}
	// Each player should get exactly 1 bye.
	for _, pl := range players {
		if byeSet[pl.ID] != 1 {
			t.Errorf("player %s had %d byes, want 1", pl.ID, byeSet[pl.ID])
		}
	}
}

func TestPairDoubleRoundRobin(t *testing.T) {
	// 4 players, 2 cycles → 6 rounds. Each pair plays twice.
	cycles := 2
	p := New(Options{Cycles: &cycles})
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
	}

	pairCount := make(map[string]int) // pair → count
	for round := 1; round <= 6; round++ {
		state := &chesspairing.TournamentState{
			Players:      players,
			CurrentRound: round,
		}
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair error: %v", round, err)
		}
		if len(result.Pairings) != 2 {
			t.Fatalf("round %d: expected 2 pairings, got %d", round, len(result.Pairings))
		}
		for _, pair := range result.Pairings {
			key := pairKey(pair.WhiteID, pair.BlackID)
			pairCount[key]++
		}
	}

	// Each pair should play exactly twice.
	for key, count := range pairCount {
		if count != 2 {
			t.Errorf("pair %s played %d times, want 2", key, count)
		}
	}
	if len(pairCount) != 6 {
		t.Errorf("expected 6 unique pairs, got %d", len(pairCount))
	}
}

func TestPairColorReversalInDoubleRR(t *testing.T) {
	// In double RR with color balance, cycle 2 reverses colors.
	// 2 players: round 1 (cycle 1), round 2 (cycle 2).
	cycles := 2
	p := New(Options{Cycles: &cycles})
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
	}

	state1 := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}
	result1, err := p.Pair(context.Background(), state1)
	if err != nil {
		t.Fatalf("round 1: Pair error: %v", err)
	}

	state2 := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 2,
	}
	result2, err := p.Pair(context.Background(), state2)
	if err != nil {
		t.Fatalf("round 2: Pair error: %v", err)
	}

	// Colors should be reversed.
	if result1.Pairings[0].WhiteID == result2.Pairings[0].WhiteID {
		t.Errorf("colors not reversed: round 1 white=%s, round 2 white=%s",
			result1.Pairings[0].WhiteID, result2.Pairings[0].WhiteID)
	}
}

func TestPairInvalidRound(t *testing.T) {
	p := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		CurrentRound: 5, // Only 1 round for 2 players
	}
	_, err := p.Pair(context.Background(), state)
	if err == nil {
		t.Error("expected error for round > total rounds")
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
	// Only 2 active players → 1 pairing, no bye.
	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
	// Verify p2 is not in any pairing.
	for _, pair := range result.Pairings {
		if pair.WhiteID == "p2" || pair.BlackID == "p2" {
			t.Error("inactive player p2 should not be paired")
		}
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

func TestPairSixPlayersAllRounds(t *testing.T) {
	// 6 players → 5 rounds. All 15 pairs should play.
	p := New(Options{})
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "A", Rating: 2000, Active: true},
		{ID: "p2", DisplayName: "B", Rating: 1900, Active: true},
		{ID: "p3", DisplayName: "C", Rating: 1800, Active: true},
		{ID: "p4", DisplayName: "D", Rating: 1700, Active: true},
		{ID: "p5", DisplayName: "E", Rating: 1600, Active: true},
		{ID: "p6", DisplayName: "F", Rating: 1500, Active: true},
	}

	pairSet := make(map[string]bool)
	for round := 1; round <= 5; round++ {
		state := &chesspairing.TournamentState{
			Players:      players,
			CurrentRound: round,
		}
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair error: %v", round, err)
		}
		if len(result.Pairings) != 3 {
			t.Fatalf("round %d: expected 3 pairings, got %d", round, len(result.Pairings))
		}
		for _, pair := range result.Pairings {
			key := pairKey(pair.WhiteID, pair.BlackID)
			if pairSet[key] {
				t.Errorf("round %d: duplicate pairing %s", round, key)
			}
			pairSet[key] = true
		}
	}

	// 6 choose 2 = 15 unique pairs.
	if len(pairSet) != 15 {
		t.Errorf("expected 15 unique pairings, got %d", len(pairSet))
	}
}

// Options tests

func TestOptionsWithDefaults(t *testing.T) {
	o := Options{}.WithDefaults()
	if *o.Cycles != 1 {
		t.Errorf("Cycles = %v, want 1", *o.Cycles)
	}
	if *o.ColorBalance != true {
		t.Errorf("ColorBalance = %v, want true", *o.ColorBalance)
	}
}

func TestOptionsPreservesExplicit(t *testing.T) {
	cycles := 3
	balance := false
	o := Options{
		Cycles:       &cycles,
		ColorBalance: &balance,
	}.WithDefaults()
	if *o.Cycles != 3 {
		t.Errorf("Cycles = %v, want 3", *o.Cycles)
	}
	if *o.ColorBalance != false {
		t.Errorf("ColorBalance = %v, want false", *o.ColorBalance)
	}
}

func TestParseOptions(t *testing.T) {
	m := map[string]any{
		"cycles":       2,
		"colorBalance": false,
		"unknownField": "ignored",
	}
	o := ParseOptions(m)
	if o.Cycles == nil || *o.Cycles != 2 {
		t.Errorf("Cycles = %v, want 2", o.Cycles)
	}
	if o.ColorBalance == nil || *o.ColorBalance != false {
		t.Errorf("ColorBalance = %v, want false", o.ColorBalance)
	}
}

func TestBergerTableOdd5Players(t *testing.T) {
	p := New(Options{})
	players := make([]chesspairing.PlayerEntry, 5)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("p%d", i+1),
			DisplayName: fmt.Sprintf("Player %d", i+1),
			Rating:      2000 - i*100,
			Active:      true,
		}
	}

	state := &chesspairing.TournamentState{
		Players: players,
		PairingConfig: chesspairing.PairingConfig{
			System:  chesspairing.PairingRoundRobin,
			Options: map[string]any{},
		},
	}

	pairSet := make(map[string]bool)
	byeCount := make(map[string]int)

	for round := 1; round <= 5; round++ {
		state.CurrentRound = round
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair error: %v", round, err)
		}

		// 5 players → 6 table size → 3 positions per side → 2 pairings + 1 bye.
		if len(result.Pairings) != 2 {
			t.Fatalf("round %d: expected 2 pairings, got %d", round, len(result.Pairings))
		}
		if len(result.Byes) != 1 {
			t.Fatalf("round %d: expected 1 bye, got %d", round, len(result.Byes))
		}

		byeCount[result.Byes[0].PlayerID]++

		for _, pair := range result.Pairings {
			key := pairKey(pair.WhiteID, pair.BlackID)
			if pairSet[key] {
				t.Errorf("round %d: duplicate pairing %s", round, key)
			}
			pairSet[key] = true
		}

		// Append results including byes.
		games := make([]chesspairing.GameData, len(result.Pairings))
		for i, pair := range result.Pairings {
			games[i] = chesspairing.GameData{
				WhiteID: pair.WhiteID,
				BlackID: pair.BlackID,
				Result:  chesspairing.ResultDraw,
			}
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{
			Number: round,
			Games:  games,
			Byes:   result.Byes,
		})
	}

	// Each player gets exactly 1 bye across all 5 rounds.
	for _, pl := range players {
		if byeCount[pl.ID] != 1 {
			t.Errorf("player %s: got %d byes, want 1", pl.ID, byeCount[pl.ID])
		}
	}

	// C(5,2) = 10 unique pairs.
	if len(pairSet) != 10 {
		t.Errorf("expected 10 unique pairings, got %d", len(pairSet))
	}
}

func TestPairRoundZeroError(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Player 1", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Player 2", Rating: 1900, Active: true},
		},
		CurrentRound: 0,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingRoundRobin,
		},
	}

	p := New(Options{})
	_, err := p.Pair(context.Background(), state)
	if err == nil {
		t.Error("expected error for round 0, got nil")
	}
}

func TestPairNegativeRoundError(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Player 1", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Player 2", Rating: 1900, Active: true},
		},
		CurrentRound: -1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingRoundRobin,
		},
	}

	p := New(Options{})
	_, err := p.Pair(context.Background(), state)
	if err == nil {
		t.Error("expected error for negative round, got nil")
	}
}

func TestPairTripleRoundRobin(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Player 1", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Player 2", Rating: 1900, Active: true},
			{ID: "p3", DisplayName: "Player 3", Rating: 1800, Active: true},
		},
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingRoundRobin,
		},
	}

	cycles := 3
	p := New(Options{Cycles: &cycles})
	totalRounds := 3 * 3 // 3 cycles * 3 rounds per cycle (odd 3 players -> table size 4 -> 3 rounds per cycle)

	pairCounts := make(map[string]int)
	for round := 1; round <= totalRounds; round++ {
		state.CurrentRound = round
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: %v", round, err)
		}

		for _, pr := range result.Pairings {
			a, b := pr.WhiteID, pr.BlackID
			if a > b {
				a, b = b, a
			}
			pairCounts[a+"-"+b]++
		}

		games := make([]chesspairing.GameData, len(result.Pairings))
		for i, pr := range result.Pairings {
			games[i] = chesspairing.GameData{
				WhiteID: pr.WhiteID,
				BlackID: pr.BlackID,
				Result:  chesspairing.ResultDraw,
			}
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{
			Number: round,
			Games:  games,
			Byes:   result.Byes,
		})
	}

	// C(3,2) = 3 unique pairs, each played 3 times.
	if len(pairCounts) != 3 {
		t.Errorf("got %d unique pairs, want 3", len(pairCounts))
	}
	for pair, count := range pairCounts {
		if count != 3 {
			t.Errorf("pair %s played %d times, want 3", pair, count)
		}
	}
}

// pairKey creates a canonical key for a pair (order-independent).
func pairKey(a, b string) string {
	if a < b {
		return a + "-" + b
	}
	return b + "-" + a
}

// --- FIDE Berger table golden test helpers ---

type expectedPairing struct {
	board int
	white string
	black string
}

func makePlayers(n int) []chesspairing.PlayerEntry {
	players := make([]chesspairing.PlayerEntry, n)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("p%d", i+1),
			DisplayName: fmt.Sprintf("Player %d", i+1),
			Rating:      2000 - i*100,
			Active:      true,
		}
	}
	return players
}

func requireNilErr(t *testing.T, err error, round int) {
	t.Helper()
	if err != nil {
		t.Fatalf("round %d: Pair error: %v", round, err)
	}
}

func requirePairings(t *testing.T, result *chesspairing.PairingResult, round int, expected []expectedPairing) {
	t.Helper()
	if len(result.Pairings) != len(expected) {
		t.Fatalf("round %d: expected %d pairings, got %d", round, len(expected), len(result.Pairings))
	}
	for i, exp := range expected {
		got := result.Pairings[i]
		if got.Board != exp.board || got.WhiteID != exp.white || got.BlackID != exp.black {
			t.Errorf("round %d board %d: got %s(W)-%s(B), want %s(W)-%s(B)",
				round, exp.board, got.WhiteID, got.BlackID, exp.white, exp.black)
		}
	}
}

// --- FIDE Berger table golden tests ---
// Expected values come from the official FIDE C.05 Annex 1 Berger tables.
// In each pairing X-Y, X is White and Y is Black.

func TestFIDEBerger4Players(t *testing.T) {
	p := New(Options{})
	players := makePlayers(4)
	state := &chesspairing.TournamentState{Players: players}

	// Round 1: 1-4  2-3
	state.CurrentRound = 1
	result, err := p.Pair(context.Background(), state)
	requireNilErr(t, err, 1)
	requirePairings(t, result, 1, []expectedPairing{
		{1, "p1", "p4"},
		{2, "p2", "p3"},
	})

	// Round 2: 4-3  1-2
	state.CurrentRound = 2
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 2)
	requirePairings(t, result, 2, []expectedPairing{
		{1, "p4", "p3"},
		{2, "p1", "p2"},
	})

	// Round 3: 2-4  3-1
	state.CurrentRound = 3
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 3)
	requirePairings(t, result, 3, []expectedPairing{
		{1, "p2", "p4"},
		{2, "p3", "p1"},
	})
}

func TestFIDEBerger6Players(t *testing.T) {
	p := New(Options{})
	players := makePlayers(6)
	state := &chesspairing.TournamentState{Players: players}

	// Round 1: 1-6  2-5  3-4
	state.CurrentRound = 1
	result, err := p.Pair(context.Background(), state)
	requireNilErr(t, err, 1)
	requirePairings(t, result, 1, []expectedPairing{
		{1, "p1", "p6"},
		{2, "p2", "p5"},
		{3, "p3", "p4"},
	})

	// Round 2: 6-4  5-3  1-2
	state.CurrentRound = 2
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 2)
	requirePairings(t, result, 2, []expectedPairing{
		{1, "p6", "p4"},
		{2, "p5", "p3"},
		{3, "p1", "p2"},
	})

	// Round 3: 2-6  3-1  4-5
	state.CurrentRound = 3
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 3)
	requirePairings(t, result, 3, []expectedPairing{
		{1, "p2", "p6"},
		{2, "p3", "p1"},
		{3, "p4", "p5"},
	})

	// Round 4: 6-5  1-4  2-3
	state.CurrentRound = 4
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 4)
	requirePairings(t, result, 4, []expectedPairing{
		{1, "p6", "p5"},
		{2, "p1", "p4"},
		{3, "p2", "p3"},
	})

	// Round 5: 3-6  4-2  5-1
	state.CurrentRound = 5
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 5)
	requirePairings(t, result, 5, []expectedPairing{
		{1, "p3", "p6"},
		{2, "p4", "p2"},
		{3, "p5", "p1"},
	})
}

func TestFIDEBerger8Players(t *testing.T) {
	p := New(Options{})
	players := makePlayers(8)
	state := &chesspairing.TournamentState{Players: players}

	// Round 1: 1-8  2-7  3-6  4-5
	state.CurrentRound = 1
	result, err := p.Pair(context.Background(), state)
	requireNilErr(t, err, 1)
	requirePairings(t, result, 1, []expectedPairing{
		{1, "p1", "p8"},
		{2, "p2", "p7"},
		{3, "p3", "p6"},
		{4, "p4", "p5"},
	})

	// Round 2: 8-5  6-4  7-3  1-2
	state.CurrentRound = 2
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 2)
	requirePairings(t, result, 2, []expectedPairing{
		{1, "p8", "p5"},
		{2, "p6", "p4"},
		{3, "p7", "p3"},
		{4, "p1", "p2"},
	})

	// Round 3: 2-8  3-1  4-7  5-6
	state.CurrentRound = 3
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 3)
	requirePairings(t, result, 3, []expectedPairing{
		{1, "p2", "p8"},
		{2, "p3", "p1"},
		{3, "p4", "p7"},
		{4, "p5", "p6"},
	})

	// Round 4: 8-6  7-5  1-4  2-3
	state.CurrentRound = 4
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 4)
	requirePairings(t, result, 4, []expectedPairing{
		{1, "p8", "p6"},
		{2, "p7", "p5"},
		{3, "p1", "p4"},
		{4, "p2", "p3"},
	})

	// Round 5: 3-8  4-2  5-1  6-7
	state.CurrentRound = 5
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 5)
	requirePairings(t, result, 5, []expectedPairing{
		{1, "p3", "p8"},
		{2, "p4", "p2"},
		{3, "p5", "p1"},
		{4, "p6", "p7"},
	})

	// Round 6: 8-7  1-6  2-5  3-4
	state.CurrentRound = 6
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 6)
	requirePairings(t, result, 6, []expectedPairing{
		{1, "p8", "p7"},
		{2, "p1", "p6"},
		{3, "p2", "p5"},
		{4, "p3", "p4"},
	})

	// Round 7: 4-8  5-3  6-2  7-1
	state.CurrentRound = 7
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 7)
	requirePairings(t, result, 7, []expectedPairing{
		{1, "p4", "p8"},
		{2, "p5", "p3"},
		{3, "p6", "p2"},
		{4, "p7", "p1"},
	})
}

func TestFIDEBerger5PlayersOdd(t *testing.T) {
	p := New(Options{})
	players := makePlayers(5)
	state := &chesspairing.TournamentState{Players: players}

	// 5 players → table size 6, player 6 = bye dummy.
	// Bye board is skipped, remaining boards renumbered from 1.

	// Round 1: 1-BYE  2-5  3-4  → bye=p1
	state.CurrentRound = 1
	result, err := p.Pair(context.Background(), state)
	requireNilErr(t, err, 1)
	requirePairings(t, result, 1, []expectedPairing{
		{1, "p2", "p5"},
		{2, "p3", "p4"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p1" {
		t.Errorf("round 1: bye want p1, got %v", result.Byes)
	}

	// Round 2: BYE-4  5-3  1-2  → bye=p4
	state.CurrentRound = 2
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 2)
	requirePairings(t, result, 2, []expectedPairing{
		{1, "p5", "p3"},
		{2, "p1", "p2"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p4" {
		t.Errorf("round 2: bye want p4, got %v", result.Byes)
	}

	// Round 3: 2-BYE  3-1  4-5  → bye=p2
	state.CurrentRound = 3
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 3)
	requirePairings(t, result, 3, []expectedPairing{
		{1, "p3", "p1"},
		{2, "p4", "p5"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p2" {
		t.Errorf("round 3: bye want p2, got %v", result.Byes)
	}

	// Round 4: BYE-5  1-4  2-3  → bye=p5
	state.CurrentRound = 4
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 4)
	requirePairings(t, result, 4, []expectedPairing{
		{1, "p1", "p4"},
		{2, "p2", "p3"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p5" {
		t.Errorf("round 4: bye want p5, got %v", result.Byes)
	}

	// Round 5: 3-BYE  4-2  5-1  → bye=p3
	state.CurrentRound = 5
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 5)
	requirePairings(t, result, 5, []expectedPairing{
		{1, "p4", "p2"},
		{2, "p5", "p1"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p3" {
		t.Errorf("round 5: bye want p3, got %v", result.Byes)
	}
}

func TestFIDEBerger7PlayersOdd(t *testing.T) {
	p := New(Options{})
	players := makePlayers(7)
	state := &chesspairing.TournamentState{Players: players}

	// 7 players → table size 8, player 8 = bye dummy.

	// Round 1: 1-BYE  2-7  3-6  4-5  → bye=p1
	state.CurrentRound = 1
	result, err := p.Pair(context.Background(), state)
	requireNilErr(t, err, 1)
	requirePairings(t, result, 1, []expectedPairing{
		{1, "p2", "p7"},
		{2, "p3", "p6"},
		{3, "p4", "p5"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p1" {
		t.Errorf("round 1: bye want p1, got %v", result.Byes)
	}

	// Round 2: BYE-5  6-4  7-3  1-2  → bye=p5
	state.CurrentRound = 2
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 2)
	requirePairings(t, result, 2, []expectedPairing{
		{1, "p6", "p4"},
		{2, "p7", "p3"},
		{3, "p1", "p2"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p5" {
		t.Errorf("round 2: bye want p5, got %v", result.Byes)
	}

	// Round 3: 2-BYE  3-1  4-7  5-6  → bye=p2
	state.CurrentRound = 3
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 3)
	requirePairings(t, result, 3, []expectedPairing{
		{1, "p3", "p1"},
		{2, "p4", "p7"},
		{3, "p5", "p6"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p2" {
		t.Errorf("round 3: bye want p2, got %v", result.Byes)
	}

	// Round 4: BYE-6  7-5  1-4  2-3  → bye=p6
	state.CurrentRound = 4
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 4)
	requirePairings(t, result, 4, []expectedPairing{
		{1, "p7", "p5"},
		{2, "p1", "p4"},
		{3, "p2", "p3"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p6" {
		t.Errorf("round 4: bye want p6, got %v", result.Byes)
	}

	// Round 5: 3-BYE  4-2  5-1  6-7  → bye=p3
	state.CurrentRound = 5
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 5)
	requirePairings(t, result, 5, []expectedPairing{
		{1, "p4", "p2"},
		{2, "p5", "p1"},
		{3, "p6", "p7"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p3" {
		t.Errorf("round 5: bye want p3, got %v", result.Byes)
	}

	// Round 6: BYE-7  1-6  2-5  3-4  → bye=p7
	state.CurrentRound = 6
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 6)
	requirePairings(t, result, 6, []expectedPairing{
		{1, "p1", "p6"},
		{2, "p2", "p5"},
		{3, "p3", "p4"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p7" {
		t.Errorf("round 6: bye want p7, got %v", result.Byes)
	}

	// Round 7: 4-BYE  5-3  6-2  7-1  → bye=p4
	state.CurrentRound = 7
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 7)
	requirePairings(t, result, 7, []expectedPairing{
		{1, "p5", "p3"},
		{2, "p6", "p2"},
		{3, "p7", "p1"},
	})
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p4" {
		t.Errorf("round 7: bye want p4, got %v", result.Byes)
	}
}

func TestSwapLastTwoRoundsDefaultTrue(t *testing.T) {
	o := Options{}.WithDefaults()
	if *o.SwapLastTwoRounds != true {
		t.Errorf("SwapLastTwoRounds = %v, want true", *o.SwapLastTwoRounds)
	}
}

func TestSwapLastTwoRoundsParseOptions(t *testing.T) {
	m := map[string]any{
		"swapLastTwoRounds": false,
	}
	o := ParseOptions(m)
	if o.SwapLastTwoRounds == nil || *o.SwapLastTwoRounds != false {
		t.Errorf("SwapLastTwoRounds = %v, want false", o.SwapLastTwoRounds)
	}
}

func TestSwapLastTwoRoundsExplicitPreserved(t *testing.T) {
	swap := false
	o := Options{SwapLastTwoRounds: &swap}.WithDefaults()
	if *o.SwapLastTwoRounds != false {
		t.Errorf("SwapLastTwoRounds = %v, want false", *o.SwapLastTwoRounds)
	}
}

func TestDoubleRRSwapLastTwoRounds4Players(t *testing.T) {
	// 4 players, double RR, swap enabled (default).
	// Cycle 1 unswapped schedule: R1, R2, R3 (roundInCycle 0,1,2).
	// With swap: R1, R3, R2 (roundInCycle 0, 2, 1).
	// Cycle 2: normal order with reversed colors.
	//
	// Unswapped cycle 1:
	//   R1 (ric=0): 1-4, 2-3
	//   R2 (ric=1): 4-3, 1-2
	//   R3 (ric=2): 2-4, 3-1
	//
	// Swapped cycle 1 (R2↔R3):
	//   R1 (ric=0): 1-4, 2-3
	//   R2 (ric=2): 2-4, 3-1    ← was round 3
	//   R3 (ric=1): 4-3, 1-2    ← was round 2
	//
	// Cycle 2 (normal order, colors reversed):
	//   R4 (ric=0): 4-1, 3-2
	//   R5 (ric=1): 3-4, 2-1
	//   R6 (ric=2): 4-2, 1-3

	cycles := 2
	p := New(Options{Cycles: &cycles})
	players := makePlayers(4)
	state := &chesspairing.TournamentState{Players: players}

	// Round 1 (cycle 1, ric=0): 1-4, 2-3
	state.CurrentRound = 1
	result, err := p.Pair(context.Background(), state)
	requireNilErr(t, err, 1)
	requirePairings(t, result, 1, []expectedPairing{
		{1, "p1", "p4"},
		{2, "p2", "p3"},
	})

	// Round 2 (cycle 1, swapped: ric=2): 2-4, 3-1
	state.CurrentRound = 2
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 2)
	requirePairings(t, result, 2, []expectedPairing{
		{1, "p2", "p4"},
		{2, "p3", "p1"},
	})

	// Round 3 (cycle 1, swapped: ric=1): 4-3, 1-2
	state.CurrentRound = 3
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 3)
	requirePairings(t, result, 3, []expectedPairing{
		{1, "p4", "p3"},
		{2, "p1", "p2"},
	})

	// Round 4 (cycle 2, ric=0, colors reversed): 4-1, 3-2
	state.CurrentRound = 4
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 4)
	requirePairings(t, result, 4, []expectedPairing{
		{1, "p4", "p1"},
		{2, "p3", "p2"},
	})

	// Round 5 (cycle 2, ric=1, colors reversed): 3-4, 2-1
	state.CurrentRound = 5
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 5)
	requirePairings(t, result, 5, []expectedPairing{
		{1, "p3", "p4"},
		{2, "p2", "p1"},
	})

	// Round 6 (cycle 2, ric=2, colors reversed): 4-2, 1-3
	state.CurrentRound = 6
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 6)
	requirePairings(t, result, 6, []expectedPairing{
		{1, "p4", "p2"},
		{2, "p1", "p3"},
	})
}

func TestDoubleRRNoThreeConsecutiveSameColor(t *testing.T) {
	for _, n := range []int{4, 6, 8} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			cycles := 2
			p := New(Options{Cycles: &cycles})
			players := makePlayers(n)
			state := &chesspairing.TournamentState{Players: players}

			tableSize := n
			if n%2 == 1 {
				tableSize = n + 1
			}
			totalRounds := (tableSize - 1) * 2

			// Track per-player color sequence across all rounds.
			colorSeq := make(map[string][]string) // playerID → ["W","B","W",...]

			for round := 1; round <= totalRounds; round++ {
				state.CurrentRound = round
				result, err := p.Pair(context.Background(), state)
				if err != nil {
					t.Fatalf("round %d: %v", round, err)
				}
				for _, pair := range result.Pairings {
					colorSeq[pair.WhiteID] = append(colorSeq[pair.WhiteID], "W")
					colorSeq[pair.BlackID] = append(colorSeq[pair.BlackID], "B")
				}
			}

			// Check no player has 3+ consecutive same color.
			for pid, seq := range colorSeq {
				for i := 2; i < len(seq); i++ {
					if seq[i] == seq[i-1] && seq[i] == seq[i-2] {
						t.Errorf("player %s has 3 consecutive %s starting at game %d: %v",
							pid, seq[i], i-1, seq)
					}
				}
			}
		})
	}
}

func TestDoubleRRSwapDisabledProducesOriginalSchedule(t *testing.T) {
	// With swap disabled, the schedule should match the unswapped Berger order.
	// 4 players: cycle 1 round 2 should be ric=1 (4-3, 1-2), not ric=2.
	cycles := 2
	swap := false
	p := New(Options{Cycles: &cycles, SwapLastTwoRounds: &swap})
	players := makePlayers(4)
	state := &chesspairing.TournamentState{Players: players}

	// Round 2 (cycle 1, ric=1, no swap): 4-3, 1-2
	state.CurrentRound = 2
	result, err := p.Pair(context.Background(), state)
	requireNilErr(t, err, 2)
	requirePairings(t, result, 2, []expectedPairing{
		{1, "p4", "p3"},
		{2, "p1", "p2"},
	})

	// Round 3 (cycle 1, ric=2, no swap): 2-4, 3-1
	state.CurrentRound = 3
	result, err = p.Pair(context.Background(), state)
	requireNilErr(t, err, 3)
	requirePairings(t, result, 3, []expectedPairing{
		{1, "p2", "p4"},
		{2, "p3", "p1"},
	})
}

func TestDoubleRRSwapNoOpForTwoPlayers(t *testing.T) {
	// 2 players: 1 round per cycle, swap has nothing to swap.
	// Should produce the same result whether swap is true or false.
	players := makePlayers(2)

	for _, swap := range []bool{true, false} {
		t.Run(fmt.Sprintf("swap=%v", swap), func(t *testing.T) {
			cycles := 2
			s := swap
			p := New(Options{Cycles: &cycles, SwapLastTwoRounds: &s})
			state := &chesspairing.TournamentState{Players: players}

			// Round 1
			state.CurrentRound = 1
			r1, err := p.Pair(context.Background(), state)
			requireNilErr(t, err, 1)
			if len(r1.Pairings) != 1 {
				t.Fatalf("expected 1 pairing, got %d", len(r1.Pairings))
			}

			// Round 2
			state.CurrentRound = 2
			r2, err := p.Pair(context.Background(), state)
			requireNilErr(t, err, 2)
			if len(r2.Pairings) != 1 {
				t.Fatalf("expected 1 pairing, got %d", len(r2.Pairings))
			}

			// Colors should be reversed between rounds.
			if r1.Pairings[0].WhiteID == r2.Pairings[0].WhiteID {
				t.Errorf("colors not reversed between cycles")
			}
		})
	}
}

func TestFIDEBergerColorBalance(t *testing.T) {
	for _, tc := range []struct {
		n            int
		maxImbalance int
	}{
		{3, 0},  // odd: perfect balance (each player plays n-1 = 2 games with bye)
		{4, 1},  // even
		{5, 0},  // odd
		{6, 1},  // even
		{7, 0},  // odd
		{8, 1},  // even
		{10, 1}, // even
	} {
		t.Run(fmt.Sprintf("N=%d", tc.n), func(t *testing.T) {
			p := New(Options{})
			players := makePlayers(tc.n)
			state := &chesspairing.TournamentState{Players: players}

			tableSize := tc.n
			if tc.n%2 == 1 {
				tableSize = tc.n + 1
			}
			totalRounds := tableSize - 1

			whiteCount := make(map[string]int)
			blackCount := make(map[string]int)

			for round := 1; round <= totalRounds; round++ {
				state.CurrentRound = round
				result, err := p.Pair(context.Background(), state)
				if err != nil {
					t.Fatalf("round %d: Pair error: %v", round, err)
				}

				for _, pair := range result.Pairings {
					whiteCount[pair.WhiteID]++
					blackCount[pair.BlackID]++
				}
			}

			for i := 0; i < tc.n; i++ {
				pid := fmt.Sprintf("p%d", i+1)
				diff := whiteCount[pid] - blackCount[pid]
				if diff < 0 {
					diff = -diff
				}
				if diff > tc.maxImbalance {
					t.Errorf("player %s: color imbalance %d (white=%d, black=%d), want <= %d",
						pid, diff, whiteCount[pid], blackCount[pid], tc.maxImbalance)
				}
			}
		})
	}
}
