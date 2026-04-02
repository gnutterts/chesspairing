package varma_test

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/algorithm/varma"
	"github.com/gnutterts/chesspairing/pairing/roundrobin"
)

func TestVarmaRoundRobinFederationSeparation(t *testing.T) {
	// 12 players from 3 federations (4 each).
	// Varma should assign them so same-federation players are spread across groups,
	// minimizing early-round same-federation pairings.
	players := []chesspairing.PlayerEntry{
		{ID: "ned1", DisplayName: "De Groot", Rating: 2400, Federation: "NED", Active: true},
		{ID: "ned2", DisplayName: "De Vries", Rating: 2300, Federation: "NED", Active: true},
		{ID: "ned3", DisplayName: "Jansen", Rating: 2200, Federation: "NED", Active: true},
		{ID: "ned4", DisplayName: "Van Dijk", Rating: 2100, Federation: "NED", Active: true},
		{ID: "usa1", DisplayName: "Adams", Rating: 2350, Federation: "USA", Active: true},
		{ID: "usa2", DisplayName: "Baker", Rating: 2250, Federation: "USA", Active: true},
		{ID: "usa3", DisplayName: "Clark", Rating: 2150, Federation: "USA", Active: true},
		{ID: "usa4", DisplayName: "Davis", Rating: 2050, Federation: "USA", Active: true},
		{ID: "ind1", DisplayName: "Gupta", Rating: 2380, Federation: "IND", Active: true},
		{ID: "ind2", DisplayName: "Kumar", Rating: 2280, Federation: "IND", Active: true},
		{ID: "ind3", DisplayName: "Patel", Rating: 2180, Federation: "IND", Active: true},
		{ID: "ind4", DisplayName: "Sharma", Rating: 2080, Federation: "IND", Active: true},
	}

	// Step 1: Assign pairing numbers via Varma.
	assigned, err := varma.Assign(players)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if len(assigned) != 12 {
		t.Fatalf("Assign returned %d players, want 12", len(assigned))
	}

	// Log assignments.
	for i, p := range assigned {
		t.Logf("pairing number %d: %s (%s)", i+1, p.DisplayName, p.Federation)
	}

	// Step 2: Use Varma-assigned ordering as player list for round-robin.
	state := &chesspairing.TournamentState{
		Players: assigned,
		PairingConfig: chesspairing.PairingConfig{
			System:  chesspairing.PairingRoundRobin,
			Options: map[string]any{},
		},
	}

	pairer := roundrobin.New(roundrobin.Options{})
	totalRounds := 11 // 12-1 = 11

	// Count same-federation pairings per round.
	sameFedByRound := make([]int, totalRounds)
	fedOf := make(map[string]string)
	for _, p := range assigned {
		fedOf[p.ID] = p.Federation
	}

	for round := 1; round <= totalRounds; round++ {
		state.CurrentRound = round
		result, err := pairer.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: %v", round, err)
		}

		for _, p := range result.Pairings {
			if fedOf[p.WhiteID] == fedOf[p.BlackID] {
				sameFedByRound[round-1]++
			}
		}

		games := make([]chesspairing.GameData, len(result.Pairings))
		for i, p := range result.Pairings {
			games[i] = chesspairing.GameData{
				WhiteID: p.WhiteID,
				BlackID: p.BlackID,
				Result:  chesspairing.ResultDraw,
			}
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{
			Number: round,
			Games:  games,
		})
	}

	// Log all rounds.
	for round := 0; round < totalRounds; round++ {
		t.Logf("round %d: %d same-federation pairings", round+1, sameFedByRound[round])
	}

	// Total same-federation pairings across 11 rounds must equal C(4,2)*3 = 18
	// (each pair of same-federation players meets exactly once).
	totalSameFed := 0
	for _, count := range sameFedByRound {
		totalSameFed += count
	}
	if totalSameFed != 18 {
		t.Errorf("total same-federation pairings: %d, want 18", totalSameFed)
	}

	// Early rounds (1-3) should have significantly fewer same-federation pairings
	// than a random distribution would give. With Varma separation,
	// we expect the early rounds to be lower. Assert total across first 3 rounds <= 7.
	earlyRoundSameFed := 0
	for round := 0; round < 3; round++ {
		earlyRoundSameFed += sameFedByRound[round]
	}
	if earlyRoundSameFed > 7 {
		t.Errorf("first 3 rounds have %d same-federation pairings, want <= 7 (Varma separation failed)", earlyRoundSameFed)
	}
}
