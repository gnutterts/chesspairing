// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing_test

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/burstein"
	"github.com/gnutterts/chesspairing/pairing/doubleswiss"
	"github.com/gnutterts/chesspairing/pairing/dubov"
	"github.com/gnutterts/chesspairing/pairing/dutch"
	"github.com/gnutterts/chesspairing/pairing/keizer"
	"github.com/gnutterts/chesspairing/pairing/lim"
	"github.com/gnutterts/chesspairing/pairing/team"
)

func TestAssignPairingNumbers_Sorting(t *testing.T) {
	// Mixed criteria: rating, title, name, and original order.
	players := []chesspairing.PlayerEntry{
		{ID: "p1", Rating: 2000, Title: "", DisplayName: "Alice"},
		{ID: "p2", Rating: 2000, Title: "FM", DisplayName: "Bob"},
		{ID: "p3", Rating: 2200, Title: "IM", DisplayName: "Charlie"},
		{ID: "p4", Rating: 2000, Title: "FM", DisplayName: "Dave"},
		{ID: "p5", Rating: 2000, Title: "FM", DisplayName: "Bob"},
		{ID: "p6", Rating: 2000, Title: "WFM", DisplayName: "Eve"},
	}

	out, err := chesspairing.AssignPairingNumbers(players)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected order:
	// 1. Charlie (2200, IM)
	// 2. Bob (p2) (2000, FM, name=Bob, index=1)
	// 3. Bob (p5) (2000, FM, name=Bob, index=4)
	// 4. Dave (p4) (2000, FM, name=Dave)
	// 5. Eve (p6) (2000, WFM)
	// 6. Alice (p1) (2000, no title)

	expectedOrder := []string{"p3", "p2", "p5", "p4", "p6", "p1"}
	for i, id := range expectedOrder {
		var found *chesspairing.PlayerEntry
		for j := range out {
			if out[j].ID == id {
				found = &out[j]
				break
			}
		}
		if found == nil {
			t.Fatalf("missing %s", id)
		}
		if found.PairingNumber != i+1 {
			t.Errorf("player %s got PairingNumber %d, want %d", id, found.PairingNumber, i+1)
		}
	}

	// The output preserves entry order and the input is not mutated.
	for i, p := range out {
		if p.ID != players[i].ID {
			t.Errorf("expected out[%d] to be %s, got %s", i, players[i].ID, p.ID)
		}
		if players[i].PairingNumber != 0 {
			t.Errorf("input player %s was mutated", players[i].ID)
		}
	}
}

func TestAssignPairingNumbers_TitleOrder(t *testing.T) {
	titles := []string{"", "WCM", "WFM", "CM", "WIM", "FM", "WGM", "IM", "GM"}
	players := make([]chesspairing.PlayerEntry, len(titles))
	for i, title := range titles {
		players[i] = chesspairing.PlayerEntry{
			ID:          title,
			DisplayName: "Same Name",
			Rating:      2000,
			Title:       title,
		}
	}

	out, err := chesspairing.AssignPairingNumbers(players)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]int{"GM": 1, "IM": 2, "WGM": 3, "FM": 4, "WIM": 5, "CM": 6, "WFM": 7, "WCM": 8, "": 9}
	for _, player := range out {
		if player.PairingNumber != want[player.Title] {
			t.Errorf("title %q got pairing number %d, want %d", player.Title, player.PairingNumber, want[player.Title])
		}
	}
}

func TestAssignPairingNumbers_AllMissingIncludesFutureEntriesInRanking(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "late", DisplayName: "Late", Rating: 2400, JoinedRound: 3},
		{ID: "original", DisplayName: "Original", Rating: 2200},
	}

	out, err := chesspairing.AssignPairingNumbers(players)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out[0].PairingNumber != 1 || out[1].PairingNumber != 2 {
		t.Fatalf("pairing numbers = %d, %d, want 1, 2", out[0].PairingNumber, out[1].PairingNumber)
	}
}

func TestAssignPairingNumbers_PreservesCompleteAssignment(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "a", PairingNumber: 17},
		{ID: "b", PairingNumber: 3},
	}

	out, err := chesspairing.AssignPairingNumbers(players)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out[0].PairingNumber != 17 || out[1].PairingNumber != 3 {
		t.Fatalf("pairing numbers = %d, %d, want 17, 3", out[0].PairingNumber, out[1].PairingNumber)
	}
}

func TestAssignPairingNumbers_LateEntries(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", PairingNumber: 2},
		{ID: "late-a", Rating: 1500, JoinedRound: 2},
		{ID: "p2", PairingNumber: 5},
		{ID: "late-b", Rating: 3000, JoinedRound: 3},
	}

	out, err := chesspairing.AssignPairingNumbers(players)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]int{
		"p1":     2,
		"p2":     5,
		"late-b": 6,
		"late-a": 7,
	}

	for _, p := range out {
		if p.PairingNumber != expected[p.ID] {
			t.Errorf("player %s got PN %d, want %d", p.ID, p.PairingNumber, expected[p.ID])
		}
	}
}

func TestAssignPairingNumbers_Errors(t *testing.T) {
	t.Run("partial", func(t *testing.T) {
		players := []chesspairing.PlayerEntry{
			{ID: "p1", PairingNumber: 1},
			{ID: "p2", PairingNumber: 0},
		}
		_, err := chesspairing.AssignPairingNumbers(players)
		if err == nil || err.Error() != "pairing numbers partially set" {
			t.Errorf("expected partial error, got %v", err)
		}
	})

	t.Run("partial non-late entry", func(t *testing.T) {
		players := []chesspairing.PlayerEntry{
			{ID: "p1", PairingNumber: 1},
			{ID: "p2"},
		}
		_, err := chesspairing.AssignPairingNumbers(players)
		if err == nil || err.Error() != "pairing numbers partially set" {
			t.Errorf("expected partial error, got %v", err)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		players := []chesspairing.PlayerEntry{
			{ID: "p1", PairingNumber: 1},
			{ID: "p2", PairingNumber: 1},
		}
		_, err := chesspairing.AssignPairingNumbers(players)
		if err == nil || err.Error() != "duplicate pairing number 1" {
			t.Errorf("expected duplicate error, got %v", err)
		}
	})
}

func TestPairersRejectDuplicatePairingNumbers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", PairingNumber: 1},
			{ID: "b", PairingNumber: 1},
		},
		CurrentRound: 1,
	}
	pairers := map[string]chesspairing.Pairer{
		"Burstein":     burstein.New(burstein.Options{}),
		"Double-Swiss": doubleswiss.New(doubleswiss.Options{}),
		"Dubov":        dubov.New(dubov.Options{}),
		"Dutch":        dutch.New(dutch.Options{}),
		"Keizer":       keizer.New(keizer.Options{}),
		"Lim":          lim.New(lim.Options{}),
		"Team":         team.New(team.Options{}),
	}
	for name, pairer := range pairers {
		t.Run(name, func(t *testing.T) {
			_, err := pairer.Pair(context.Background(), state)
			if err == nil || err.Error() != "duplicate pairing number 1" {
				t.Errorf("Pair() error = %v, want duplicate pairing number error", err)
			}
		})
	}
}
