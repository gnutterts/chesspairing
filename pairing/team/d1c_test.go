// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestD1C_CrossBracketPairingCompletesRound(t *testing.T) {
	totalRounds := 3
	preferenceNone := "none"
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "t1", Rating: 2500}, {ID: "t2", Rating: 2400}, {ID: "t3", Rating: 2300}, {ID: "t4", Rating: 2200}, {ID: "t5", Rating: 2100}, {ID: "t6", Rating: 2000}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "t1", BlackID: "t2", Result: chesspairing.ResultWhiteWins}, {WhiteID: "t3", BlackID: "t4", Result: chesspairing.ResultWhiteWins}, {WhiteID: "t5", BlackID: "t6", Result: chesspairing.ResultWhiteWins}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "t1", BlackID: "t5", Result: chesspairing.ResultDraw}, {WhiteID: "t3", BlackID: "t6", Result: chesspairing.ResultWhiteWins}, {WhiteID: "t2", BlackID: "t4", Result: chesspairing.ResultWhiteWins}}},
		},
		CurrentRound:  3,
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingTeam},
	}
	result, err := New(Options{TotalRounds: &totalRounds, ColorPreferenceType: &preferenceNone}).Pair(context.Background(), state)
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
			t.Errorf("team %s is not paired or assigned a bye", player.ID)
		}
	}
}
