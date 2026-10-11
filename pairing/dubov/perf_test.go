// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// TestDubov_40Players_9Rounds_AllComplete is the Article 3.2.4 regression
// guard: the first stage (shifting players out of the smaller subgroup for
// unavoidable same-subgroup pairs) lets every round of a 40-player, 9-round
// tournament complete. Before the first stage this failed in round 7.
func TestDubov_40Players_9Rounds_AllComplete(t *testing.T) {
	runDubovTournament(t, 40, 9, 9, 9010)
}

// TestDubov_100Players_11Rounds_CompletesQuickly guards the transposition
// depth-first search: every round of a 100-player, 11-round tournament pairs
// within a few seconds instead of hanging or returning ErrBracketTooLarge.
func TestDubov_100Players_11Rounds_CompletesQuickly(t *testing.T) {
	runDubovTournamentWithLimit(t, 100, 11, 42, 11001, 5*time.Second)
}

// TestDubov_200Players_11Rounds_CompletesQuickly is the 200-player guard from
// DUBOV.md. It runs in the normal suite and is skipped under -short.
func TestDubov_200Players_11Rounds_CompletesQuickly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 200-player timing guard in -short mode")
	}
	runDubovTournamentWithLimit(t, 200, 11, 43, 12001, 5*time.Second)
}

func runDubovTournament(t *testing.T, playerCount, roundCount int, seed1, seed2 uint64) {
	t.Helper()
	runDubovTournamentWithLimit(t, playerCount, roundCount, seed1, seed2, 0)
}

func runDubovTournamentWithLimit(t *testing.T, playerCount, roundCount int, seed1, seed2 uint64, roundLimit time.Duration) {
	t.Helper()
	entries := make([]chesspairing.PlayerEntry, playerCount)
	for i := range entries {
		entries[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("P%03d", i+1),
			DisplayName: fmt.Sprintf("Player %03d", i+1),
			Rating:      2500 - i*5,
		}
	}
	totalRounds := roundCount
	pairer := New(Options{TotalRounds: &totalRounds})
	state := &chesspairing.TournamentState{Players: entries, CurrentRound: 1}
	rng := rand.New(rand.NewPCG(seed1, seed2))
	for round := 1; round <= roundCount; round++ {
		state.CurrentRound = round
		start := time.Now()
		result, err := pairer.Pair(context.Background(), state)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("round %d: Pair() error = %v", round, err)
		}
		// The race detector slows the search several times; the tournament
		// still runs in full, but the per-round limit only holds without it.
		if roundLimit > 0 && !raceEnabled && elapsed >= roundLimit {
			t.Fatalf("round %d took %v, want < %v", round, elapsed, roundLimit)
		}
		if len(result.Pairings) != playerCount/2 {
			t.Fatalf("round %d: got %d pairings, want %d", round, len(result.Pairings), playerCount/2)
		}
		if len(result.Byes) != 0 {
			t.Fatalf("round %d: got %d byes, want 0", round, len(result.Byes))
		}
		swisslib.AssertPairingInvariants(t, state, result)

		rd := chesspairing.RoundData{Number: round, Byes: append([]chesspairing.ByeEntry(nil), result.Byes...)}
		for _, gp := range result.Pairings {
			var res chesspairing.GameResult
			switch rng.IntN(3) {
			case 0:
				res = chesspairing.ResultWhiteWins
			case 1:
				res = chesspairing.ResultBlackWins
			default:
				res = chesspairing.ResultDraw
			}
			rd.Games = append(rd.Games, chesspairing.GameData{WhiteID: gp.WhiteID, BlackID: gp.BlackID, Result: res})
		}
		state.Rounds = append(state.Rounds, rd)
	}
}
