package roundrobin

import (
	"context"
	"fmt"
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

func TestBergerTableGolden4Players(t *testing.T) {
	p := New(Options{})
	players := make([]chesspairing.PlayerEntry, 4)
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

	for round := 1; round <= 3; round++ {
		state.CurrentRound = round
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair error: %v", round, err)
		}

		// Each round must have exactly 2 pairings.
		if len(result.Pairings) != 2 {
			t.Fatalf("round %d: expected 2 pairings, got %d", round, len(result.Pairings))
		}

		// All 4 players must appear exactly once per round.
		seen := make(map[string]int)
		for _, pair := range result.Pairings {
			seen[pair.WhiteID]++
			seen[pair.BlackID]++
		}
		if len(seen) != 4 {
			t.Errorf("round %d: expected 4 unique players, got %d", round, len(seen))
		}
		for pid, count := range seen {
			if count != 1 {
				t.Errorf("round %d: player %s appeared %d times, want 1", round, pid, count)
			}
		}

		for _, pair := range result.Pairings {
			key := pairKey(pair.WhiteID, pair.BlackID)
			if pairSet[key] {
				t.Errorf("round %d: duplicate pairing %s", round, key)
			}
			pairSet[key] = true
		}

		// Append results so state accumulates history.
		games := make([]chesspairing.GameData, len(result.Pairings))
		for i, pair := range result.Pairings {
			games[i] = chesspairing.GameData{
				WhiteID: pair.WhiteID,
				BlackID: pair.BlackID,
				Result:  chesspairing.ResultDraw,
			}
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{Number: round, Games: games})
	}

	// C(4,2) = 6 unique pairs.
	if len(pairSet) != 6 {
		t.Errorf("expected 6 unique pairings, got %d", len(pairSet))
	}
}

func TestBergerTableGolden6Players(t *testing.T) {
	p := New(Options{})
	players := make([]chesspairing.PlayerEntry, 6)
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

	for round := 1; round <= 5; round++ {
		state.CurrentRound = round
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair error: %v", round, err)
		}

		// Each round must have exactly 3 pairings.
		if len(result.Pairings) != 3 {
			t.Fatalf("round %d: expected 3 pairings, got %d", round, len(result.Pairings))
		}

		// All 6 players must appear exactly once per round.
		seen := make(map[string]int)
		for _, pair := range result.Pairings {
			seen[pair.WhiteID]++
			seen[pair.BlackID]++
		}
		if len(seen) != 6 {
			t.Errorf("round %d: expected 6 unique players, got %d", round, len(seen))
		}
		for pid, count := range seen {
			if count != 1 {
				t.Errorf("round %d: player %s appeared %d times, want 1", round, pid, count)
			}
		}

		for _, pair := range result.Pairings {
			key := pairKey(pair.WhiteID, pair.BlackID)
			if pairSet[key] {
				t.Errorf("round %d: duplicate pairing %s", round, key)
			}
			pairSet[key] = true
		}

		// Append results so state accumulates history.
		games := make([]chesspairing.GameData, len(result.Pairings))
		for i, pair := range result.Pairings {
			games[i] = chesspairing.GameData{
				WhiteID: pair.WhiteID,
				BlackID: pair.BlackID,
				Result:  chesspairing.ResultDraw,
			}
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{Number: round, Games: games})
	}

	// C(6,2) = 15 unique pairs.
	if len(pairSet) != 15 {
		t.Errorf("expected 15 unique pairings, got %d", len(pairSet))
	}
}

func TestBergerTableColorBalance(t *testing.T) {
	for _, n := range []int{4, 6, 8, 10} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			p := New(Options{})
			players := make([]chesspairing.PlayerEntry, n)
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

			totalRounds := n - 1
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

				// Append results.
				games := make([]chesspairing.GameData, len(result.Pairings))
				for i, pair := range result.Pairings {
					games[i] = chesspairing.GameData{
						WhiteID: pair.WhiteID,
						BlackID: pair.BlackID,
						Result:  chesspairing.ResultDraw,
					}
				}
				state.Rounds = append(state.Rounds, chesspairing.RoundData{Number: round, Games: games})
			}

			// Verify each player plays N-1 total games with color imbalance at most 1.
			for i := 0; i < n; i++ {
				pid := fmt.Sprintf("p%d", i+1)
				total := whiteCount[pid] + blackCount[pid]
				if total != totalRounds {
					t.Errorf("player %s: played %d games, want %d", pid, total, totalRounds)
				}

				diff := whiteCount[pid] - blackCount[pid]
				if diff < 0 {
					diff = -diff
				}
				if diff > 1 {
					t.Errorf("player %s: color imbalance %d (white=%d, black=%d), want <= 1",
						pid, diff, whiteCount[pid], blackCount[pid])
				}
			}
		})
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
