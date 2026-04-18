// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "testing"

func TestC1_NoRematches(t *testing.T) {
	ctx := &CriteriaContext{
		Players: map[string]*PlayerState{
			"p1": {ID: "p1", Opponents: []string{"p2"}},
			"p2": {ID: "p2", Opponents: []string{"p1"}},
			"p3": {ID: "p3", Opponents: nil},
		},
	}

	// p1 vs p2: already played → fails C1
	pair := &ProposedPairing{
		White: ctx.Players["p1"],
		Black: ctx.Players["p2"],
	}
	if C1NoRematches(pair, ctx) {
		t.Error("p1 vs p2 should fail C1 (already played)")
	}

	// p1 vs p3: never played → passes C1
	pair2 := &ProposedPairing{
		White: ctx.Players["p1"],
		Black: ctx.Players["p3"],
	}
	if !C1NoRematches(pair2, ctx) {
		t.Error("p1 vs p3 should pass C1 (never played)")
	}
}

func TestC2_NoSecondPAB(t *testing.T) {
	ctx := &CriteriaContext{}

	player := &PlayerState{ID: "p1", ByeReceived: false}
	if !C2NoSecondPAB(player, ctx) {
		t.Error("player without bye should pass C2")
	}

	player2 := &PlayerState{ID: "p2", ByeReceived: true}
	if C2NoSecondPAB(player2, ctx) {
		t.Error("player with prior bye should fail C2")
	}
}

func TestC3_AbsoluteColorConflict(t *testing.T) {
	ctx := &CriteriaContext{
		TopScorers: map[string]bool{},
	}

	// Both have absolute White → fails C3
	p1 := &PlayerState{ID: "p1", ColorHistory: []Color{ColorWhite, ColorWhite, ColorWhite}}
	p2 := &PlayerState{ID: "p2", ColorHistory: []Color{ColorWhite, ColorWhite, ColorWhite}}
	pair := &ProposedPairing{White: p1, Black: p2}
	if C3AbsoluteColorConflict(pair, ctx) {
		t.Error("both absolute White should fail C3")
	}

	// Opposite absolutes → passes C3
	p3 := &PlayerState{ID: "p3", ColorHistory: []Color{ColorBlack, ColorBlack, ColorBlack}}
	pair2 := &ProposedPairing{White: p1, Black: p3}
	if !C3AbsoluteColorConflict(pair2, ctx) {
		t.Error("opposite absolute preferences should pass C3")
	}

	// One absolute, other non-absolute → passes C3
	p4 := &PlayerState{ID: "p4", ColorHistory: []Color{ColorWhite}}
	pair3 := &ProposedPairing{White: p1, Black: p4}
	if !C3AbsoluteColorConflict(pair3, ctx) {
		t.Error("one absolute + one non-absolute should pass C3")
	}
}

func TestC3_TopScorerException(t *testing.T) {
	ctx := &CriteriaContext{
		IsLastRound: true,
		TopScorers:  map[string]bool{"p1": true, "p2": true},
	}

	p1 := &PlayerState{ID: "p1", ColorHistory: []Color{ColorWhite, ColorWhite, ColorWhite}}
	p2 := &PlayerState{ID: "p2", ColorHistory: []Color{ColorWhite, ColorWhite, ColorWhite}}
	pair := &ProposedPairing{White: p1, Black: p2}
	if !C3AbsoluteColorConflict(pair, ctx) {
		t.Error("both topscorers in final round should pass C3 (exception)")
	}
}

func TestCanonicalPairKey(t *testing.T) {
	tests := []struct {
		a, b string
		want [2]string
	}{
		{"p1", "p2", [2]string{"p1", "p2"}},
		{"p2", "p1", [2]string{"p1", "p2"}},
		{"abc", "abc", [2]string{"abc", "abc"}},
	}
	for _, tt := range tests {
		got := CanonicalPairKey(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("CanonicalPairKey(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsForbiddenPair(t *testing.T) {
	forbidden := map[[2]string]bool{
		{"p1", "p3"}: true,
		{"p2", "p4"}: true,
	}
	ctx := &CriteriaContext{
		ForbiddenPairs: forbidden,
	}
	tests := []struct {
		whiteID, blackID string
		wantForbidden    bool
	}{
		{"p1", "p3", true},
		{"p3", "p1", true},
		{"p2", "p4", true},
		{"p1", "p2", false},
		{"p3", "p4", false},
	}
	for _, tt := range tests {
		pair := &ProposedPairing{
			White: &PlayerState{ID: tt.whiteID},
			Black: &PlayerState{ID: tt.blackID},
		}
		got := IsForbiddenPair(pair, ctx)
		if got != tt.wantForbidden {
			t.Errorf("IsForbiddenPair(%s vs %s) = %v, want %v",
				tt.whiteID, tt.blackID, got, tt.wantForbidden)
		}
	}
}

func TestIsForbiddenPair_NilMap(t *testing.T) {
	ctx := &CriteriaContext{}
	pair := &ProposedPairing{
		White: &PlayerState{ID: "p1"},
		Black: &PlayerState{ID: "p2"},
	}
	if IsForbiddenPair(pair, ctx) {
		t.Error("expected false when ForbiddenPairs is nil")
	}
}

func TestIsPairForbiddenByID(t *testing.T) {
	forbidden := map[[2]string]bool{
		{"p1", "p3"}: true,
		{"p2", "p4"}: true,
	}
	ctx := &CriteriaContext{
		ForbiddenPairs: forbidden,
	}

	tests := []struct {
		aID, bID      string
		wantForbidden bool
	}{
		{"p1", "p3", true},
		{"p3", "p1", true}, // reversed order
		{"p2", "p4", true},
		{"p1", "p2", false}, // not forbidden
		{"p3", "p4", false},
	}

	for _, tt := range tests {
		got := IsPairForbiddenByID(tt.aID, tt.bID, ctx)
		if got != tt.wantForbidden {
			t.Errorf("IsPairForbiddenByID(%q, %q) = %v, want %v",
				tt.aID, tt.bID, got, tt.wantForbidden)
		}
	}
}

func TestIsPairForbiddenByID_NilMap(t *testing.T) {
	ctx := &CriteriaContext{}
	if IsPairForbiddenByID("p1", "p2", ctx) {
		t.Error("expected false when ForbiddenPairs is nil")
	}
}

func TestC4CompleteBracket(t *testing.T) {
	bp := &BracketPairing{
		Pairs:    []ProposedPairing{{}, {}},
		Floaters: nil,
	}
	if !C4CompleteBracket(bp, 4) {
		t.Error("2 pairs from 4 players should pass C4")
	}

	bp2 := &BracketPairing{
		Pairs:    []ProposedPairing{{}},
		Floaters: []*PlayerState{{}},
	}
	if !C4CompleteBracket(bp2, 3) {
		t.Error("1 pair + 1 floater from 3 players should pass C4")
	}

	bp3 := &BracketPairing{
		Pairs: []ProposedPairing{{}},
	}
	if C4CompleteBracket(bp3, 4) {
		t.Error("1 pair from 4 players should fail C4")
	}
}
