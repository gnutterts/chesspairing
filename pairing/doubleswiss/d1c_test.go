// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestD1C_CrossBracketPairingCompletesRound(t *testing.T) {
	totalRounds := 3
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1", Rating: 2500}, {ID: "p2", Rating: 2400}, {ID: "p3", Rating: 2300}, {ID: "p4", Rating: 2200}, {ID: "p5", Rating: 2100}, {ID: "p6", Rating: 2000}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p5", BlackID: "p6", Result: chesspairing.ResultWhiteWins}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p5", Result: chesspairing.ResultDraw}, {WhiteID: "p3", BlackID: "p6", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins}}},
		},
		CurrentRound:  3,
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDoubleSwiss},
	}
	result, err := New(Options{TotalRounds: &totalRounds}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error = %v", err)
	}
	seen := make(map[string]bool)
	for _, pairing := range result.Pairings {
		seen[pairing.WhiteID] = true
		seen[pairing.BlackID] = true
	}
	for _, bye := range result.Byes {
		seen[bye.PlayerID] = true
	}
	for _, player := range state.Players {
		if !seen[player.ID] {
			t.Errorf("player %s is not paired or assigned a bye", player.ID)
		}
	}
}
