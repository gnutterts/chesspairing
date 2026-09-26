// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// FIDE C.04.4.1 Dubov System Art. 1.7.1: "ARO is defined for each player who
// has played at least one game. It is given by the sum of the ratings of the
// opponents the player met over-the-board, divided by the number of such
// opponents, and rounded to the nearest integer number."
//
// A withdrawn opponent still counts with its rating. The rating map must be
// built from the full tournament state, not only from the active players.
func TestRepro_D11_AROWithdrawnOpponentCountsWithRating(t *testing.T) {
	withdrawnAfterRound1 := 1
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P1", Rating: 2000},
			{ID: "p2", DisplayName: "P2", Rating: 2500, WithdrawnAfterRound: &withdrawnAfterRound1},
			{ID: "p3", DisplayName: "P3", Rating: 2200},
			{ID: "p4", DisplayName: "P4", Rating: 2300},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}

	playerStates, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		t.Fatal(err)
	}
	ratings := BuildRatingMapFromState(state)

	var p1 *swisslib.PlayerState
	for i := range playerStates {
		if playerStates[i].ID == "p1" {
			p1 = &playerStates[i]
		}
	}
	if p1 == nil {
		t.Fatal("p1 not found in active players")
	}

	gotARO := ComputeARO(p1, ratings)
	if gotARO != 2500 {
		t.Fatalf("ARO(p1) = %v, want 2500 (withdrawn opponent p2 keeps its rating)", gotARO)
	}
}

// TestRepro_D11_AROSortOrder asserts that the ARO-based ascending order
// includes the withdrawn opponent's rating, so p3 (ARO 2300) sorts before
// p1 (ARO 2500).
func TestRepro_D11_AROSortOrder(t *testing.T) {
	withdrawnAfterRound1 := 1
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P1", Rating: 2000},
			{ID: "p2", DisplayName: "P2", Rating: 2500, WithdrawnAfterRound: &withdrawnAfterRound1},
			{ID: "p3", DisplayName: "P3", Rating: 2200},
			{ID: "p4", DisplayName: "P4", Rating: 2300},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}

	playerStates, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		t.Fatal(err)
	}
	ratings := BuildRatingMapFromState(state)

	var p1, p3 *swisslib.PlayerState
	for i := range playerStates {
		switch playerStates[i].ID {
		case "p1":
			p1 = &playerStates[i]
		case "p3":
			p3 = &playerStates[i]
		}
	}
	if p1 == nil || p3 == nil {
		t.Fatal("p1 or p3 not found in active players")
	}

	order := []*swisslib.PlayerState{p1, p3}
	SortByAROAscending(order, ratings)
	if order[0].ID != "p3" || order[1].ID != "p1" {
		t.Fatalf("ARO order = %s, %s; want p3 (ARO 2300), p1 (ARO 2500)", order[0].ID, order[1].ID)
	}
}
