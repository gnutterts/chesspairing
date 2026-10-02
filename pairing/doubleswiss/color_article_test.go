// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestAllocateColor_SymmetricWhiteHistories(t *testing.T) {
	white := lexswiss.ColorWhite
	hrp := &lexswiss.ParticipantState{
		ID:            "p1",
		PairingNumber: 1,
		Score:         2,
		ColorHistory:  []lexswiss.Color{white, white},
	}
	opponent := &lexswiss.ParticipantState{
		ID:            "p2",
		PairingNumber: 2,
		Score:         2,
		ColorHistory:  []lexswiss.Color{white, white},
	}

	whiteID, blackID := AllocateColor(hrp, opponent, nil)
	if whiteID != "p2" || blackID != "p1" {
		t.Errorf("AllocateColor() = White %s, Black %s; want White p2, Black p1", whiteID, blackID)
	}
}

func TestPair_LastRoundGivesWhiteToPlayerWithFewerWhites(t *testing.T) {
	totalRounds := 4
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "A", Rating: 2500},
			{ID: "B", Rating: 2400},
			{ID: "C", Rating: 2300},
			{ID: "D", Rating: 2200},
			{ID: "E", Rating: 2100},
			{ID: "F", Rating: 2000},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "C", BlackID: "A", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "B", BlackID: "D", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "E", BlackID: "F", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "A", BlackID: "F", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "B", BlackID: "E", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "C", BlackID: "D", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "D", BlackID: "A", Result: chesspairing.ResultBlackWins},
					{WhiteID: "B", BlackID: "F", Result: chesspairing.ResultBlackWins},
					{WhiteID: "C", BlackID: "E", Result: chesspairing.ResultBlackWins},
				},
			},
		},
		CurrentRound: 4,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	result, err := New(Options{TotalRounds: &totalRounds}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	for _, pairing := range result.Pairings {
		if pairing.WhiteID == "A" && pairing.BlackID == "B" {
			return
		}
		if pairing.WhiteID == "B" && pairing.BlackID == "A" {
			t.Errorf("A has fewer Whites than B, but B received White")
			return
		}
	}
	t.Errorf("A and B were not paired: %v", result.Pairings)
}
