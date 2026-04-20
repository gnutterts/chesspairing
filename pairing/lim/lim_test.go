// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPair_Round1_EvenPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
			{ID: "p2", DisplayName: "Bob", Rating: 2300},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200},
			{ID: "p4", DisplayName: "Diana", Rating: 2100},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingLim,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 2 {
		t.Errorf("expected 2 pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 0 {
		t.Errorf("expected 0 byes, got %d", len(result.Byes))
	}

	// Round 1: 1v3, 2v4 (Art. 7.2: top half vs bottom half)
	checkPairing(t, result.Pairings, "p1", "p3")
	checkPairing(t, result.Pairings, "p2", "p4")
}

func TestPair_Round1_OddPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
			{ID: "p2", DisplayName: "Bob", Rating: 2300},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200},
			{ID: "p4", DisplayName: "Diana", Rating: 2100},
			{ID: "p5", DisplayName: "Eve", Rating: 2000},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingLim,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 2 {
		t.Errorf("expected 2 pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 {
		t.Errorf("expected 1 bye, got %d", len(result.Byes))
	}
	// Art. 1.1 + Art. 7.1: lowest rated player (p5) gets PAB.
	if result.Byes[0].PlayerID != "p5" {
		t.Errorf("expected p5 to get bye, got %s", result.Byes[0].PlayerID)
	}
	if result.Byes[0].Type != chesspairing.ByePAB {
		t.Errorf("expected ByePAB, got %v", result.Byes[0].Type)
	}
}

func TestPair_Round2_MedianFirstOrder(t *testing.T) {
	// After round 1: p1 beats p3, p2 beats p4.
	// Scores: p1=1, p2=1, p3=0, p4=0.
	// Median = 0.5 (1 round played / 2).
	// Scoregroups: [1.0: p1,p2] and [0.0: p3,p4].
	// Both groups are non-median. Process highest (1.0) first, then lowest (0.0).
	// No median group exists at 0.5.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
			{ID: "p2", DisplayName: "Bob", Rating: 2300},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200},
			{ID: "p4", DisplayName: "Diana", Rating: 2100},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultBlackWins},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingLim,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 2 {
		t.Errorf("expected 2 pairings, got %d", len(result.Pairings))
	}
	// p1 vs p2 (both scored 1.0, compatible, different colours in R1)
	// p3 vs p4 (both scored 0.0)
	checkPairing(t, result.Pairings, "p1", "p2")
	checkPairing(t, result.Pairings, "p3", "p4")
}

func TestPair_NoActivePlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:      nil,
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingLim,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 0 {
		t.Errorf("expected 0 pairings, got %d", len(result.Pairings))
	}
}

func TestPair_SinglePlayer(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingLim,
		},
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 0 {
		t.Errorf("expected 0 pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p1" {
		t.Error("single player should get PAB")
	}
}

// checkPairing verifies that a pairing exists between two players (in either colour order).
func checkPairing(t *testing.T, pairings []chesspairing.GamePairing, id1, id2 string) {
	t.Helper()
	for _, p := range pairings {
		if (p.WhiteID == id1 && p.BlackID == id2) || (p.WhiteID == id2 && p.BlackID == id1) {
			return
		}
	}
	t.Errorf("expected pairing between %s and %s", id1, id2)
}
