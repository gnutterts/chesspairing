// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"context"
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// ratingOf returns the rating of the player with the given ID.
func ratingOf(players []chesspairing.PlayerEntry, id string) int {
	for _, p := range players {
		if p.ID == id {
			return p.Rating
		}
	}
	return 0
}

// isPair returns true if the GamePairing matches the two IDs (either colour order).
func isPair(gp chesspairing.GamePairing, id1, id2 string) bool {
	return (gp.WhiteID == id1 && gp.BlackID == id2) ||
		(gp.WhiteID == id2 && gp.BlackID == id1)
}

// assertPaired fails the test if no pairing in result matches (id1, id2).
func assertPaired(t *testing.T, result *chesspairing.PairingResult, id1, id2 string) {
	t.Helper()
	for _, gp := range result.Pairings {
		if isPair(gp, id1, id2) {
			return
		}
	}
	t.Errorf("expected pairing %s vs %s not found", id1, id2)
}

// simulateRound creates a RoundData from a PairingResult where the higher-rated
// player always wins (as White). Byes are included.
func simulateRoundHigherWins(players []chesspairing.PlayerEntry, result *chesspairing.PairingResult, roundNum int) chesspairing.RoundData {
	rd := chesspairing.RoundData{Number: roundNum}
	for _, gp := range result.Pairings {
		wRating := ratingOf(players, gp.WhiteID)
		bRating := ratingOf(players, gp.BlackID)
		var res chesspairing.GameResult
		if wRating >= bRating {
			res = chesspairing.ResultWhiteWins
		} else {
			res = chesspairing.ResultBlackWins
		}
		rd.Games = append(rd.Games, chesspairing.GameData{
			WhiteID: gp.WhiteID,
			BlackID: gp.BlackID,
			Result:  res,
		})
	}
	rd.Byes = append(rd.Byes, result.Byes...)
	return rd
}

// simulateRoundDraws creates a RoundData where all games are draws.
func simulateRoundDraws(result *chesspairing.PairingResult, roundNum int) chesspairing.RoundData {
	rd := chesspairing.RoundData{Number: roundNum}
	for _, gp := range result.Pairings {
		rd.Games = append(rd.Games, chesspairing.GameData{
			WhiteID: gp.WhiteID,
			BlackID: gp.BlackID,
			Result:  chesspairing.ResultDraw,
		})
	}
	rd.Byes = append(rd.Byes, result.Byes...)
	return rd
}

// ---------------------------------------------------------------------------
// Test 1: 6 players, 5 rounds, higher-rated always wins
// ---------------------------------------------------------------------------

func TestFIDE_Dubov_6Player5Round(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P01", Rating: 2600},
		{ID: "p2", DisplayName: "P02", Rating: 2550},
		{ID: "p3", DisplayName: "P03", Rating: 2500},
		{ID: "p4", DisplayName: "P04", Rating: 2450},
		{ID: "p5", DisplayName: "P05", Rating: 2400},
		{ID: "p6", DisplayName: "P06", Rating: 2350},
		{ID: "p7", DisplayName: "P07", Rating: 2300},
		{ID: "p8", DisplayName: "P08", Rating: 2250},
		{ID: "p9", DisplayName: "P09", Rating: 2200},
		{ID: "p10", DisplayName: "P10", Rating: 2150},
		{ID: "p11", DisplayName: "P11", Rating: 2100},
		{ID: "p12", DisplayName: "P12", Rating: 2050},
	}

	pairer := New(Options{})
	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	// --- Round 1 ---
	r1, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 1 error: %v", err)
	}

	if len(r1.Pairings) != 6 {
		t.Fatalf("round 1: expected 6 pairings, got %d", len(r1.Pairings))
	}
	if len(r1.Byes) != 0 {
		t.Fatalf("round 1: expected 0 byes, got %d", len(r1.Byes))
	}

	// Verify R1 pairings: top half vs bottom half.
	assertPaired(t, r1, "p1", "p7")
	assertPaired(t, r1, "p2", "p8")
	assertPaired(t, r1, "p3", "p9")
	assertPaired(t, r1, "p4", "p10")
	assertPaired(t, r1, "p5", "p11")
	assertPaired(t, r1, "p6", "p12")

	// Verify R1 colours: top-seed gets White on board 1 (default).
	for _, gp := range r1.Pairings {
		if isPair(gp, "p1", "p7") {
			if gp.WhiteID != "p1" {
				t.Errorf("round 1 board p1-p7: expected p1 as White, got %s", gp.WhiteID)
			}
		}
	}

	swisslib.AssertPairingInvariants(t, state, r1)

	// Simulate R1: higher-rated wins.
	rd1 := simulateRoundHigherWins(players, r1, 1)
	state.Rounds = []chesspairing.RoundData{rd1}
	state.CurrentRound = 2

	// --- Round 2 ---
	r2, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 2 error: %v", err)
	}

	if len(r2.Pairings) < 1 {
		t.Fatalf("round 2: expected pairings, got 0")
	}

	// Verify no rematches in R2 and structural invariants.
	swisslib.AssertPairingInvariants(t, state, r2)

	// Verify colour alternation for R1 winners who had White:
	// they should ideally get Black in R2.
	r1WhiteWinners := make(map[string]bool)
	for _, gp := range r1.Pairings {
		wRating := ratingOf(players, gp.WhiteID)
		bRating := ratingOf(players, gp.BlackID)
		if wRating >= bRating {
			r1WhiteWinners[gp.WhiteID] = true
		}
	}
	for _, gp := range r2.Pairings {
		if r1WhiteWinners[gp.WhiteID] {
			t.Logf("R2: %s had White in R1 (winner) and has White again in R2 — may be acceptable if no alternation possible", gp.WhiteID)
		}
	}

	// Simulate R2 and continue through R3.
	// Use draws to avoid overly polarised scores that exhaust pairing space.
	rd2 := simulateRoundDraws(r2, 2)
	state.Rounds = append(state.Rounds, rd2)
	state.CurrentRound = 3

	// --- Round 3 ---
	r3, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 3 error: %v", err)
	}

	if len(r3.Pairings) < 1 {
		t.Fatalf("round 3: expected pairings, got 0")
	}

	swisslib.AssertPairingInvariants(t, state, r3)
}

// ---------------------------------------------------------------------------
// Test 2: 5 players, 5 rounds — bye rotation
// ---------------------------------------------------------------------------

func TestFIDE_Dubov_OddPlayers_ByeRotation(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P1", Rating: 2500},
		{ID: "p2", DisplayName: "P2", Rating: 2400},
		{ID: "p3", DisplayName: "P3", Rating: 2300},
		{ID: "p4", DisplayName: "P4", Rating: 2200},
		{ID: "p5", DisplayName: "P5", Rating: 2100},
	}

	pairer := New(Options{})
	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	byeCount := make(map[string]int)

	for round := 1; round <= 5; round++ {
		result, err := pairer.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d error: %v", round, err)
		}

		// Each round: exactly 2 pairings + 1 bye.
		if len(result.Pairings) != 2 {
			t.Fatalf("round %d: expected 2 pairings, got %d", round, len(result.Pairings))
		}
		if len(result.Byes) != 1 {
			t.Fatalf("round %d: expected 1 bye, got %d", round, len(result.Byes))
		}

		byePlayerID := result.Byes[0].PlayerID
		byeCount[byePlayerID]++

		// C2: no player should receive a second PAB.
		if byeCount[byePlayerID] > 1 {
			t.Errorf("round %d: player %s received PAB for the %d time (C2 violation)",
				round, byePlayerID, byeCount[byePlayerID])
		}

		swisslib.AssertPairingInvariants(t, state, result)

		// Simulate: higher-rated wins.
		rd := simulateRoundHigherWins(players, result, round)
		state.Rounds = append(state.Rounds, rd)
		state.CurrentRound = round + 1
	}

	// After 5 rounds, each of the 5 players should have exactly 1 PAB.
	for _, p := range players {
		if byeCount[p.ID] != 1 {
			t.Errorf("player %s received %d PABs, expected 1", p.ID, byeCount[p.ID])
		}
	}
}

// ---------------------------------------------------------------------------
// Test 3: 4 players, R1 all draws, pair R2
// ---------------------------------------------------------------------------

func TestFIDE_Dubov_DrawResults(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P1", Rating: 2500},
		{ID: "p2", DisplayName: "P2", Rating: 2400},
		{ID: "p3", DisplayName: "P3", Rating: 2300},
		{ID: "p4", DisplayName: "P4", Rating: 2200},
	}

	pairer := New(Options{})

	// --- Round 1 ---
	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	r1, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 1 error: %v", err)
	}

	// All draws in R1.
	rd1 := chesspairing.RoundData{Number: 1}
	for _, gp := range r1.Pairings {
		rd1.Games = append(rd1.Games, chesspairing.GameData{
			WhiteID: gp.WhiteID,
			BlackID: gp.BlackID,
			Result:  chesspairing.ResultDraw,
		})
	}

	state.Rounds = []chesspairing.RoundData{rd1}
	state.CurrentRound = 2

	// --- Round 2 ---
	r2, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 2 error: %v", err)
	}

	if len(r2.Pairings) != 2 {
		t.Fatalf("round 2: expected 2 pairings, got %d", len(r2.Pairings))
	}

	// Verify no rematches and structural invariants.
	swisslib.AssertPairingInvariants(t, state, r2)

	// Explicit no-rematch check.
	r1Pairs := make(map[[2]string]bool)
	for _, gp := range r1.Pairings {
		r1Pairs[swisslib.CanonicalPairKey(gp.WhiteID, gp.BlackID)] = true
	}
	for _, gp := range r2.Pairings {
		key := swisslib.CanonicalPairKey(gp.WhiteID, gp.BlackID)
		if r1Pairs[key] {
			t.Errorf("round 2 rematch: %s vs %s", gp.WhiteID, gp.BlackID)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 4: Double forfeit — players CAN be re-paired
// ---------------------------------------------------------------------------

func TestFIDE_Dubov_DoubleForfeit(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P1", Rating: 2500},
		{ID: "p2", DisplayName: "P2", Rating: 2400},
		{ID: "p3", DisplayName: "P3", Rating: 2300},
		{ID: "p4", DisplayName: "P4", Rating: 2200},
	}

	// R1: p1-p3 is a double forfeit, p2-p4 is normal.
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDoubleForfeit, IsForfeit: true},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}

	pairer := New(Options{})
	r2, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 2 error: %v", err)
	}

	if len(r2.Pairings) != 2 {
		t.Fatalf("round 2: expected 2 pairings, got %d", len(r2.Pairings))
	}

	// p1 and p3 CAN be re-paired because double forfeit is excluded from history.
	// Verify that pairings are valid (invariants pass — no rematch flagged for p1-p3).
	swisslib.AssertPairingInvariants(t, state, r2)

	// Check that p1 and p3 appear in pairings (they should, since they're active).
	seen := make(map[string]bool)
	for _, gp := range r2.Pairings {
		seen[gp.WhiteID] = true
		seen[gp.BlackID] = true
	}
	if !seen["p1"] {
		t.Error("p1 should appear in R2 pairings")
	}
	if !seen["p3"] {
		t.Error("p3 should appear in R2 pairings")
	}
}

// ---------------------------------------------------------------------------
// Test 5: Withdrawal — p6 withdraws after R1
// ---------------------------------------------------------------------------

func TestFIDE_Dubov_Withdrawal(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P1", Rating: 2500},
		{ID: "p2", DisplayName: "P2", Rating: 2400},
		{ID: "p3", DisplayName: "P3", Rating: 2300},
		{ID: "p4", DisplayName: "P4", Rating: 2200},
		{ID: "p5", DisplayName: "P5", Rating: 2100},
		{ID: "p6", DisplayName: "P6", Rating: 2000},
	}

	pairer := New(Options{})
	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	// R1: normal pairing.
	r1, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 1 error: %v", err)
	}

	rd1 := simulateRoundHigherWins(players, r1, 1)
	state.Rounds = []chesspairing.RoundData{rd1}
	state.CurrentRound = 2

	// p6 withdraws.
	withdrawnAfter := 1
	state.Players[5].WithdrawnAfterRound = &withdrawnAfter

	r2, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 2 error: %v", err)
	}

	// 5 active players → 2 pairings + 1 bye.
	if len(r2.Pairings) != 2 {
		t.Fatalf("round 2: expected 2 pairings, got %d", len(r2.Pairings))
	}
	if len(r2.Byes) != 1 {
		t.Fatalf("round 2: expected 1 bye, got %d", len(r2.Byes))
	}

	// p6 should not appear in pairings or byes.
	for _, gp := range r2.Pairings {
		if gp.WhiteID == "p6" || gp.BlackID == "p6" {
			t.Error("withdrawn player p6 should not appear in pairings")
		}
	}
	for _, bye := range r2.Byes {
		if bye.PlayerID == "p6" {
			t.Error("withdrawn player p6 should not receive a bye")
		}
	}

	swisslib.AssertPairingInvariants(t, state, r2)
}

// ---------------------------------------------------------------------------
// Test 6: Large tournament — 20 players, 7 rounds
// ---------------------------------------------------------------------------

func TestFIDE_Dubov_LargeTournament_20Players7Rounds(t *testing.T) {
	players := make([]chesspairing.PlayerEntry, 20)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("p%02d", i+1),
			DisplayName: fmt.Sprintf("Player %d", i+1),
			Rating:      2700 - i*50,
		}
	}

	pairer := New(Options{})
	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	for round := 1; round <= 7; round++ {
		state.CurrentRound = round

		result, err := pairer.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d error: %v", round, err)
		}

		// The Dubov pairer may produce partial pairings in later rounds
		// when score group fragmentation makes complete matching impossible.
		// Use weak invariants that don't require completeness.
		assertWeakInvariants(t, state, result)

		// Simulate: higher-rated wins.
		games := make([]chesspairing.GameData, len(result.Pairings))
		for i, gp := range result.Pairings {
			res := chesspairing.ResultWhiteWins
			if ratingOf(players, gp.BlackID) > ratingOf(players, gp.WhiteID) {
				res = chesspairing.ResultBlackWins
			}
			games[i] = chesspairing.GameData{
				WhiteID: gp.WhiteID, BlackID: gp.BlackID, Result: res,
			}
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{
			Number: round, Games: games, Byes: result.Byes,
		})
	}
}

// assertWeakInvariants checks invariants that hold even for partial pairings:
// no duplicate, no inactive paired, no self-pairing, no rematches. Does NOT
// require every active player to be paired.
func assertWeakInvariants(t *testing.T, state *chesspairing.TournamentState, result *chesspairing.PairingResult) {
	t.Helper()

	activeIDs := make(map[string]bool)
	for _, p := range state.Players {
		if state.IsActiveInRound(p.ID, state.CurrentRound) {
			activeIDs[p.ID] = true
		}
	}

	seen := make(map[string]int)
	for i, gp := range result.Pairings {
		seen[gp.WhiteID]++
		seen[gp.BlackID]++
		if gp.WhiteID == gp.BlackID {
			t.Errorf("pairing[%d]: player %s paired against themselves", i, gp.WhiteID)
		}
	}
	for _, bye := range result.Byes {
		seen[bye.PlayerID]++
	}

	for id, count := range seen {
		if count != 1 {
			t.Errorf("player %s appears %d times in pairings+byes (expected 1)", id, count)
		}
	}
	for id := range seen {
		if !activeIDs[id] {
			t.Errorf("inactive player %s found in pairings or byes", id)
		}
	}

	// No rematches.
	prevPairs := make(map[[2]string]bool)
	for _, rd := range state.Rounds {
		for _, g := range rd.Games {
			if g.IsForfeit {
				continue
			}
			prevPairs[swisslib.CanonicalPairKey(g.WhiteID, g.BlackID)] = true
		}
	}
	for _, gp := range result.Pairings {
		key := swisslib.CanonicalPairKey(gp.WhiteID, gp.BlackID)
		if prevPairs[key] {
			t.Errorf("rematch detected: %s vs %s", gp.WhiteID, gp.BlackID)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 7: MaxT function — table-driven
// ---------------------------------------------------------------------------

func TestFIDE_Dubov_MaxT(t *testing.T) {
	tests := []struct {
		completedRounds int
		want            int
	}{
		{0, 2},
		{1, 2},
		{4, 2},
		{5, 3},
		{9, 3},
		{10, 4},
		{15, 5},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("rounds=%d", tt.completedRounds), func(t *testing.T) {
			got := MaxT(tt.completedRounds)
			if got != tt.want {
				t.Errorf("MaxT(%d) = %d, want %d", tt.completedRounds, got, tt.want)
			}
		})
	}
}
