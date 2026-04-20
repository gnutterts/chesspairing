// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// pairing/dubov/dubov_test.go
package dubov

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestPair_Round1_FourPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000},
			{ID: "b", DisplayName: "Bob", Rating: 1800},
			{ID: "c", DisplayName: "Carol", Rating: 1600},
			{ID: "d", DisplayName: "Dave", Rating: 1400},
		},
		CurrentRound: 1,
		Rounds:       nil,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}

	if len(result.Byes) != 0 {
		t.Errorf("expected no byes, got %d", len(result.Byes))
	}

	// Validate: every player appears exactly once.
	seen := make(map[string]bool)
	for _, gp := range result.Pairings {
		seen[gp.WhiteID] = true
		seen[gp.BlackID] = true
	}
	for _, pe := range state.Players {
		if !seen[pe.ID] {
			t.Errorf("player %s not in any pairing", pe.ID)
		}
	}

	// Validate structural integrity.
	players := swisslib.BuildPlayerStates(state)
	if err := swisslib.ValidatePairing(players, result); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}

func TestPair_Round1_OddPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000},
			{ID: "b", DisplayName: "Bob", Rating: 1800},
			{ID: "c", DisplayName: "Carol", Rating: 1600},
		},
		CurrentRound: 1,
		Rounds:       nil,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 {
		t.Fatalf("expected 1 bye, got %d", len(result.Byes))
	}
	if result.Byes[0].Type != chesspairing.ByePAB {
		t.Errorf("expected PAB bye, got %v", result.Byes[0].Type)
	}
}

func TestPair_Round1_TwoPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000},
			{ID: "b", DisplayName: "Bob", Rating: 1800},
		},
		CurrentRound: 1,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
}

func TestPair_SinglePlayer(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000},
		},
		CurrentRound: 1,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Pairings) != 0 {
		t.Errorf("expected 0 pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 {
		t.Errorf("expected 1 bye, got %d", len(result.Byes))
	}
}

func TestPair_TooFewPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:      nil,
		CurrentRound: 1,
	}

	p := New(Options{})
	_, err := p.Pair(context.Background(), state)
	if err == nil {
		t.Error("expected error for no players")
	}
}

func TestPair_MultiRound_NoRematches(t *testing.T) {
	// Round 1 results: a beat c (white), b beat d (white).
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000},
			{ID: "b", DisplayName: "Bob", Rating: 1800},
			{ID: "c", DisplayName: "Carol", Rating: 1600},
			{ID: "d", DisplayName: "Dave", Rating: 1400},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "a", BlackID: "c", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "b", BlackID: "d", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}

	// Verify no rematches.
	for _, gp := range result.Pairings {
		if (gp.WhiteID == "a" && gp.BlackID == "c") || (gp.WhiteID == "c" && gp.BlackID == "a") {
			t.Error("a vs c is a rematch")
		}
		if (gp.WhiteID == "b" && gp.BlackID == "d") || (gp.WhiteID == "d" && gp.BlackID == "b") {
			t.Error("b vs d is a rematch")
		}
	}

	// Validate structural integrity.
	players := swisslib.BuildPlayerStates(state)
	if err := swisslib.ValidatePairing(players, result); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}

func TestPair_ForbiddenPairs(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000},
			{ID: "b", DisplayName: "Bob", Rating: 1800},
			{ID: "c", DisplayName: "Carol", Rating: 1600},
			{ID: "d", DisplayName: "Dave", Rating: 1400},
		},
		CurrentRound: 1,
	}

	p := New(Options{
		ForbiddenPairs: [][]string{{"a", "c"}},
	})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, gp := range result.Pairings {
		if (gp.WhiteID == "a" && gp.BlackID == "c") || (gp.WhiteID == "c" && gp.BlackID == "a") {
			t.Error("a vs c pairing should be forbidden")
		}
	}
}

func TestPair_InactivePlayers(t *testing.T) {
	// A player who never joined is simply absent from state.Players in the
	// new model.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000},
			{ID: "b", DisplayName: "Bob", Rating: 1800},
			{ID: "d", DisplayName: "Dave", Rating: 1400},
		},
		CurrentRound: 1,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 active → 1 pairing + 1 bye.
	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 {
		t.Fatalf("expected 1 bye, got %d", len(result.Byes))
	}
}

func TestPair_EightPlayerTournament_ThreeRounds(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P1", Rating: 2400},
		{ID: "p2", DisplayName: "P2", Rating: 2300},
		{ID: "p3", DisplayName: "P3", Rating: 2200},
		{ID: "p4", DisplayName: "P4", Rating: 2100},
		{ID: "p5", DisplayName: "P5", Rating: 2000},
		{ID: "p6", DisplayName: "P6", Rating: 1900},
		{ID: "p7", DisplayName: "P7", Rating: 1800},
		{ID: "p8", DisplayName: "P8", Rating: 1700},
	}

	pairer := New(Options{})

	// --- Round 1 ---
	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	r1, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 1 error: %v", err)
	}
	if len(r1.Pairings) != 4 {
		t.Fatalf("round 1: expected 4 pairings, got %d", len(r1.Pairings))
	}

	// Simulate round 1: all white wins.
	round1 := chesspairing.RoundData{Number: 1}
	for _, gp := range r1.Pairings {
		round1.Games = append(round1.Games, chesspairing.GameData{
			WhiteID: gp.WhiteID,
			BlackID: gp.BlackID,
			Result:  chesspairing.ResultWhiteWins,
		})
	}

	// --- Round 2 ---
	state.Rounds = []chesspairing.RoundData{round1}
	state.CurrentRound = 2

	r2, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 2 error: %v", err)
	}
	if len(r2.Pairings) != 4 {
		t.Fatalf("round 2: expected 4 pairings, got %d", len(r2.Pairings))
	}

	// Verify no rematches in round 2.
	r1Pairs := make(map[[2]string]bool)
	for _, gp := range r1.Pairings {
		key := swisslib.CanonicalPairKey(gp.WhiteID, gp.BlackID)
		r1Pairs[key] = true
	}
	for _, gp := range r2.Pairings {
		key := swisslib.CanonicalPairKey(gp.WhiteID, gp.BlackID)
		if r1Pairs[key] {
			t.Errorf("round 2 rematch: %s vs %s", gp.WhiteID, gp.BlackID)
		}
	}

	// Simulate round 2: all white wins.
	round2 := chesspairing.RoundData{Number: 2}
	for _, gp := range r2.Pairings {
		round2.Games = append(round2.Games, chesspairing.GameData{
			WhiteID: gp.WhiteID,
			BlackID: gp.BlackID,
			Result:  chesspairing.ResultWhiteWins,
		})
	}

	// --- Round 3 ---
	state.Rounds = []chesspairing.RoundData{round1, round2}
	state.CurrentRound = 3

	r3, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("round 3 error: %v", err)
	}
	if len(r3.Pairings) != 4 {
		t.Fatalf("round 3: expected 4 pairings, got %d", len(r3.Pairings))
	}

	// Verify no rematches across all rounds.
	allPrevPairs := make(map[[2]string]bool)
	for _, gp := range r1.Pairings {
		allPrevPairs[swisslib.CanonicalPairKey(gp.WhiteID, gp.BlackID)] = true
	}
	for _, gp := range r2.Pairings {
		allPrevPairs[swisslib.CanonicalPairKey(gp.WhiteID, gp.BlackID)] = true
	}
	for _, gp := range r3.Pairings {
		key := swisslib.CanonicalPairKey(gp.WhiteID, gp.BlackID)
		if allPrevPairs[key] {
			t.Errorf("round 3 rematch: %s vs %s", gp.WhiteID, gp.BlackID)
		}
	}

	// Validate round 3 structural integrity.
	pStates := swisslib.BuildPlayerStates(state)
	if err := swisslib.ValidatePairing(pStates, r3); err != nil {
		t.Errorf("round 3 validation failed: %v", err)
	}
}

func TestPair_ForfeitsExcludedFromPairingHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000},
			{ID: "b", DisplayName: "Bob", Rating: 1800},
			{ID: "c", DisplayName: "Carol", Rating: 1600},
			{ID: "d", DisplayName: "Dave", Rating: 1400},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "a", BlackID: "c", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
					{WhiteID: "b", BlackID: "d", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Forfeit excluded from pairing history → a and c CAN be re-paired.
	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}
}

func TestPair_InactivePlayers_Excluded(t *testing.T) {
	// Equivalent to TestPair_InactivePlayers: a player who is not part of
	// the tournament is simply absent from state.Players.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000},
			{ID: "b", DisplayName: "Bob", Rating: 1800},
			{ID: "d", DisplayName: "Dave", Rating: 1400},
		},
		CurrentRound: 1,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 active → 1 pairing + 1 bye.
	if len(result.Pairings) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 {
		t.Fatalf("expected 1 bye, got %d", len(result.Byes))
	}
}
