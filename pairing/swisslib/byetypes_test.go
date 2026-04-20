// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestByePointsAllByeTypes locks in the contract that byePoints()
// returns exactly the same value the score loop in BuildPlayerStates
// adds for the same bye type. Float computation depends on this
// agreement: if byePoints disagreed with the score loop, a player's
// FloatDown flag would be based on a value their pairing score does
// not reflect.
func TestByePointsAllByeTypes(t *testing.T) {
	cases := []struct {
		bt   chesspairing.ByeType
		want float64
	}{
		{chesspairing.ByePAB, 1.0},
		{chesspairing.ByeHalf, 0.5},
		{chesspairing.ByeZero, 0.0},
		{chesspairing.ByeAbsent, 0.0},
		{chesspairing.ByeExcused, 0.0},
		{chesspairing.ByeClubCommitment, 0.0},
	}

	// Sentinel: every defined ByeType must be in the table above.
	count := 0
	for b := chesspairing.ByePAB; b.IsValid(); b++ {
		count++
	}
	if count != len(cases) {
		t.Fatalf("byePoints test table covers %d types but %d ByeType values are defined; add the missing case", len(cases), count)
	}

	for _, c := range cases {
		t.Run(c.bt.String(), func(t *testing.T) {
			if got := byePoints(c.bt); got != c.want {
				t.Errorf("byePoints(%v) = %v, want %v", c.bt, got, c.want)
			}
		})
	}
}

// TestByeReceivedOnlyPAB verifies BuildPlayerStates flags
// PlayerState.ByeReceived only when the player has actually received a
// PAB. Other bye types (half, zero, absent, excused, club commitment)
// must leave the player eligible for a future PAB.
func TestByeReceivedOnlyPAB(t *testing.T) {
	for b := chesspairing.ByePAB; b.IsValid(); b++ {
		t.Run(b.String(), func(t *testing.T) {
			state := &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "p1", DisplayName: "Alice", Rating: 2000},
					{ID: "p2", DisplayName: "Bob", Rating: 1800},
				},
				Rounds: []chesspairing.RoundData{
					{
						Number: 1,
						Byes: []chesspairing.ByeEntry{
							{PlayerID: "p1", Type: b},
						},
					},
				},
			}
			players := BuildPlayerStates(state)
			var p1 *PlayerState
			for i := range players {
				if players[i].ID == "p1" {
					p1 = &players[i]
				}
			}
			if p1 == nil {
				t.Fatal("p1 not found in player states")
			}
			wantReceived := b == chesspairing.ByePAB
			if p1.ByeReceived != wantReceived {
				t.Errorf("ByeReceived after %v bye = %v, want %v", b, p1.ByeReceived, wantReceived)
			}
		})
	}
}
