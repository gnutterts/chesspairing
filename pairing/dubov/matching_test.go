// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestSplitG1G2_Round1(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, ColorHistory: nil},
		{ID: "b", TPN: 2, ColorHistory: nil},
		{ID: "c", TPN: 3, ColorHistory: nil},
		{ID: "d", TPN: 4, ColorHistory: nil},
	}

	g1, g2 := SplitG1G2(players, true)

	if len(g1) != 2 || len(g2) != 2 {
		t.Fatalf("expected 2+2 split, got %d+%d", len(g1), len(g2))
	}
	if g1[0].ID != "a" || g1[1].ID != "b" {
		t.Errorf("G1 should be first half: got %s, %s", g1[0].ID, g1[1].ID)
	}
	if g2[0].ID != "c" || g2[1].ID != "d" {
		t.Errorf("G2 should be second half: got %s, %s", g2[0].ID, g2[1].ID)
	}
}

func TestSplitG1G2_ByColorPreference(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, ColorHistory: []swisslib.Color{swisslib.ColorBlack}}, // prefers white -> G1
		{ID: "b", TPN: 2, ColorHistory: []swisslib.Color{swisslib.ColorWhite}}, // prefers black -> G2
		{ID: "c", TPN: 3, ColorHistory: []swisslib.Color{swisslib.ColorBlack}}, // prefers white -> G1
		{ID: "d", TPN: 4, ColorHistory: []swisslib.Color{swisslib.ColorWhite}}, // prefers black -> G2
	}

	g1, g2 := SplitG1G2(players, false)

	if len(g1) != 2 || len(g2) != 2 {
		t.Fatalf("expected 2+2 split, got %d+%d", len(g1), len(g2))
	}
	// G1 = white-preference players.
	for _, p := range g1 {
		if p.ID != "a" && p.ID != "c" {
			t.Errorf("unexpected player %s in G1", p.ID)
		}
	}
	// G2 = black-preference players.
	for _, p := range g2 {
		if p.ID != "b" && p.ID != "d" {
			t.Errorf("unexpected player %s in G2", p.ID)
		}
	}
}

func TestSplitG1G2_OddCount(t *testing.T) {
	// With 5 players in round 1: G1=2, G2=3.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, ColorHistory: nil},
		{ID: "b", TPN: 2, ColorHistory: nil},
		{ID: "c", TPN: 3, ColorHistory: nil},
		{ID: "d", TPN: 4, ColorHistory: nil},
		{ID: "e", TPN: 5, ColorHistory: nil},
	}

	g1, g2 := SplitG1G2(players, true)

	if len(g1)+len(g2) != 5 {
		t.Fatalf("total should be 5, got %d+%d", len(g1), len(g2))
	}
	if len(g1) != 2 || len(g2) != 3 {
		t.Errorf("expected 2+3 split for round 1 with 5 players, got %d+%d", len(g1), len(g2))
	}
}

func TestBalanceG1G2(t *testing.T) {
	ratings := map[string]int{"a": 2000, "b": 1800, "c": 1600, "d": 1400}

	t.Run("already balanced", func(t *testing.T) {
		g1 := []*swisslib.PlayerState{
			{ID: "a", TPN: 1, Rating: 2000},
			{ID: "b", TPN: 2, Rating: 1800},
		}
		g2 := []*swisslib.PlayerState{
			{ID: "c", TPN: 3, Rating: 1600},
			{ID: "d", TPN: 4, Rating: 1400},
		}

		bg1, bg2 := BalanceG1G2(g1, g2, ratings)
		if len(bg1) != 2 || len(bg2) != 2 {
			t.Errorf("expected 2+2, got %d+%d", len(bg1), len(bg2))
		}
	})

	t.Run("G1 too large", func(t *testing.T) {
		g1 := []*swisslib.PlayerState{
			{ID: "a", TPN: 1, Rating: 2000},
			{ID: "b", TPN: 2, Rating: 1800},
			{ID: "c", TPN: 3, Rating: 1600},
		}
		g2 := []*swisslib.PlayerState{
			{ID: "d", TPN: 4, Rating: 1400},
		}

		bg1, bg2 := BalanceG1G2(g1, g2, ratings)
		if len(bg1) != 2 || len(bg2) != 2 {
			t.Errorf("expected 2+2 after balancing, got %d+%d", len(bg1), len(bg2))
		}
	})
}

func TestMatchBracketDubov_Simple(t *testing.T) {
	ratings := map[string]int{
		"a": 2000, "b": 1800, "c": 1600, "d": 1400,
	}

	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2000, ColorHistory: nil, Score: 0},
		{ID: "b", TPN: 2, Rating: 1800, ColorHistory: nil, Score: 0},
		{ID: "c", TPN: 3, Rating: 1600, ColorHistory: nil, Score: 0},
		{ID: "d", TPN: 4, Rating: 1400, ColorHistory: nil, Score: 0},
	}

	bracket := swisslib.Bracket{
		Players:       players,
		Homogeneous:   true,
		OriginalScore: 0,
	}

	ctx := &matchContext{
		ratings:         ratings,
		isRound1:        true,
		forbidden:       nil,
		completedRounds: 0,
	}

	result, err := MatchBracketDubov(bracket, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(result.pairs))
	}
	if len(result.floaters) != 0 {
		t.Errorf("expected 0 floaters, got %d", len(result.floaters))
	}
}

func TestMatchBracketDubov_NeedsTransposition(t *testing.T) {
	// Setup where the identity transposition violates C1 (rematch).
	// a has played c, b has played d -> identity pairing (a-c, b-d) fails.
	// Transposition should pair a-d, b-c.
	players := []*swisslib.PlayerState{
		{
			ID: "a", TPN: 1, Rating: 2000, Score: 1, Opponents: []string{"c"},
			ColorHistory: []swisslib.Color{swisslib.ColorWhite},
		},
		{
			ID: "b", TPN: 2, Rating: 1800, Score: 1, Opponents: []string{"d"},
			ColorHistory: []swisslib.Color{swisslib.ColorWhite},
		},
		{
			ID: "c", TPN: 3, Rating: 1600, Score: 1, Opponents: []string{"a"},
			ColorHistory: []swisslib.Color{swisslib.ColorBlack},
		},
		{
			ID: "d", TPN: 4, Rating: 1400, Score: 1, Opponents: []string{"b"},
			ColorHistory: []swisslib.Color{swisslib.ColorBlack},
		},
	}

	ratings := map[string]int{"a": 2000, "b": 1800, "c": 1600, "d": 1400}

	bracket := swisslib.Bracket{
		Players:       players,
		Homogeneous:   true,
		OriginalScore: 1,
	}

	ctx := &matchContext{
		ratings:         ratings,
		isRound1:        false,
		forbidden:       nil,
		completedRounds: 1,
	}

	result, err := MatchBracketDubov(bracket, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(result.pairs))
	}

	// Verify no rematches.
	for _, pair := range result.pairs {
		if swisslib.HasPlayed(pair.white, pair.black) {
			t.Errorf("rematch detected: %s vs %s", pair.white.ID, pair.black.ID)
		}
	}
}

func TestMatchBracketDubov_SinglePlayer(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2000, Score: 3},
	}

	bracket := swisslib.Bracket{
		Players:       players,
		Homogeneous:   true,
		OriginalScore: 3,
	}

	ctx := &matchContext{
		ratings:         map[string]int{"a": 2000},
		isRound1:        false,
		completedRounds: 3,
	}

	result, err := MatchBracketDubov(bracket, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.pairs) != 0 {
		t.Errorf("expected 0 pairs for single player, got %d", len(result.pairs))
	}
	if len(result.floaters) != 1 {
		t.Errorf("expected 1 floater, got %d", len(result.floaters))
	}
}

func TestGenerateDubovTranspositions(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1},
		{ID: "b", TPN: 2},
		{ID: "c", TPN: 3},
	}

	perms := generateDubovTranspositions(players, 100)

	// 3! = 6 permutations.
	if len(perms) != 6 {
		t.Errorf("expected 6 permutations, got %d", len(perms))
	}

	// First permutation should be identity.
	if perms[0][0].ID != "a" || perms[0][1].ID != "b" || perms[0][2].ID != "c" {
		t.Errorf("first permutation should be identity, got %s,%s,%s",
			perms[0][0].ID, perms[0][1].ID, perms[0][2].ID)
	}
}

func TestGenerateDubovTranspositions_Capped(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1},
		{ID: "b", TPN: 2},
		{ID: "c", TPN: 3},
	}

	perms := generateDubovTranspositions(players, 3)

	if len(perms) != 3 {
		t.Errorf("expected 3 permutations (capped), got %d", len(perms))
	}
}

func TestGenerateDubovTranspositions_Empty(t *testing.T) {
	perms := generateDubovTranspositions(nil, 100)
	if len(perms) != 0 {
		t.Errorf("expected 0 permutations for nil, got %d", len(perms))
	}
}
