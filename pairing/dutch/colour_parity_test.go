// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestPair_ColourParityIgnoresPlayersWithRequestedBye covers C.04.3 5.2.5: the
// parity of a player's number among the players taking part in the round
// decides the colour when nobody has a preference, so the player with a
// requested bye (5) does not count. Reference: the pairing of this field in
// round 1 by bbpPairings 6.0 and JaVaFo 2.2, which give the same colours.
func TestPair_ColourParityIgnoresPlayersWithRequestedBye(t *testing.T) {
	ratings := []int{2588, 2578, 2195, 2045, 2034, 1989, 1977, 1952, 1937, 1873, 1815, 1807, 1653, 1573, 1541, 1536}
	var players []chesspairing.PlayerEntry
	for i, rating := range ratings {
		players = append(players, chesspairing.PlayerEntry{ID: fmt.Sprintf("%d", i+1), Rating: rating, PairingNumber: i + 1})
	}
	state := &chesspairing.TournamentState{
		Players:         players,
		PreAssignedByes: []chesspairing.ByeEntry{{PlayerID: "5", Type: chesspairing.ByeHalf}},
		CurrentRound:    1,
	}
	totalRounds := 7

	result, err := New(Options{TotalRounds: &totalRounds}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	want := [][2]string{{"1", "9"}, {"10", "2"}, {"3", "11"}, {"12", "4"}, {"6", "13"}, {"14", "7"}, {"8", "15"}}
	if len(result.Pairings) != len(want) {
		t.Fatalf("got %d pairings, want %d", len(result.Pairings), len(want))
	}
	for i, pairing := range result.Pairings {
		if pairing.WhiteID != want[i][0] || pairing.BlackID != want[i][1] {
			t.Errorf("board %d = %s-%s, want %s-%s", i+1, pairing.WhiteID, pairing.BlackID, want[i][0], want[i][1])
		}
	}
	var pab []string
	for _, bye := range result.Byes {
		if bye.Type == chesspairing.ByePAB {
			pab = append(pab, bye.PlayerID)
		}
	}
	if len(pab) != 1 || pab[0] != "16" {
		t.Errorf("pairing-allocated bye = %v, want [16]", pab)
	}
}
