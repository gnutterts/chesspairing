// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPair_NoMedianPairsArrivingFloaters(t *testing.T) {
	// After three rounds the scoregroups are 3, 2, 1 and 0; 1.5 is absent.
	// The only down-floater and up-floater reach the empty median position and
	// are paired there, as the interpretation of Arts. 2.2 and 3.6.3 requires.
	state := noMedianState()

	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	checkPairing(t, result.Pairings, "p1", "p3")
	checkPairing(t, result.Pairings, "p2", "p4")
	checkPairing(t, result.Pairings, "p5", "p6")
}

func TestPair_NoMedianBlockedFloatersCrack(t *testing.T) {
	// The arriving floaters cannot meet. Art. 2.6 first cracks the upper
	// pair, then the lower pair, producing a complete median pairing.
	result, err := New(Options{ForbiddenPairs: [][]string{{"p2", "p4"}}}).Pair(context.Background(), noMedianState())
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	checkPairing(t, result.Pairings, "p1", "p6")
	checkPairing(t, result.Pairings, "p2", "p3")
	checkPairing(t, result.Pairings, "p4", "p5")
}

func noMedianState() *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", Rating: 2600},
			{ID: "p2", Rating: 2500},
			{ID: "p3", Rating: 2400},
			{ID: "p4", Rating: 2300},
			{ID: "p5", Rating: 2200},
			{ID: "p6", Rating: 2100},
		},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p2", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p3", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
			}},
			{Number: 2, Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p2", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
			}},
			{Number: 3, Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p5", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p4", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
			}},
		},
		CurrentRound: 4,
	}
}

func TestPair_DownFloaterUsesCompatibleAdjacentGroup(t *testing.T) {
	// Art. 3.3 selects the lowest numbered player who has a compatible
	// opponent in the lower scoregroup. Only p1 can float here.
	players := []chesspairing.PlayerEntry{
		{ID: "p1", Rating: 2600},
		{ID: "p2", Rating: 2500},
		{ID: "p3", Rating: 2400},
		{ID: "p4", Rating: 2300},
		{ID: "p5", Rating: 2200},
		{ID: "p6", Rating: 2100},
	}
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p2", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p3", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
			},
		}},
		CurrentRound: 2,
	}
	forbidden := [][]string{
		{"p2", "p4"},
		{"p2", "p6"},
		{"p3", "p4"},
		{"p3", "p5"},
	}

	result, err := New(Options{ForbiddenPairs: forbidden}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	checkPairing(t, result.Pairings, "p1", "p5")
}
