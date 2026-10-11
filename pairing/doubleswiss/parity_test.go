// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestPair_InitialColourParityUsesFixedTPN covers C.04.5 Article 4.3.1:
// colour parity uses the fixed TPN. Players 3 to 9 have not played a match yet,
// player 2 gets the bye in round 2, and the pairs among players 3 to 8 get the
// initial colour for the odd-TPN player.
func TestPair_InitialColourParityUsesFixedTPN(t *testing.T) {
	var players []chesspairing.PlayerEntry
	for i := 1; i <= 9; i++ {
		players = append(players, chesspairing.PlayerEntry{ID: fmt.Sprintf("p%d", i), Rating: 2600 - i*10})
	}
	byes := make([]chesspairing.ByeEntry, 0, 7)
	for i := 3; i <= 9; i++ {
		byes = append(byes, chesspairing.ByeEntry{PlayerID: fmt.Sprintf("p%d", i), Type: chesspairing.ByeHalf})
	}
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games:  []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins}},
			Byes:   byes,
		}},
		CurrentRound: 2,
	}

	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	white := make(map[string]bool, len(result.Pairings))
	for _, pairing := range result.Pairings {
		white[pairing.WhiteID] = true
	}
	for _, id := range []string{"p3", "p5", "p7"} {
		if !white[id] {
			t.Errorf("%s should have White (odd TPN, initial colour White); pairings = %v", id, result.Pairings)
		}
	}
}
