// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestFloaterType_Order(t *testing.T) {
	if FloaterTypeA >= FloaterTypeB {
		t.Error("TypeA should be worse (lower value) than TypeB")
	}
	if FloaterTypeB >= FloaterTypeC {
		t.Error("TypeB should be worse than TypeC")
	}
	if FloaterTypeC >= FloaterTypeD {
		t.Error("TypeC should be worse than TypeD")
	}
}

func TestClassifyFloater(t *testing.T) {
	// Player who already floated + no compatible opponent in adjacent group.
	p := &swisslib.PlayerState{ID: "p1", Opponents: []string{"adj1"}}
	adjacent := []*swisslib.PlayerState{
		{ID: "adj1", Opponents: []string{"p1"}}, // already played p1
	}

	ft := ClassifyFloater(p, true, adjacent, nil)
	if ft != FloaterTypeA {
		t.Errorf("already floated + no compatible opponent = TypeA, got %v", ft)
	}
}

func TestClassifyFloater_AlreadyFloatedWithCompatible(t *testing.T) {
	p := &swisslib.PlayerState{ID: "p1"}
	adjacent := []*swisslib.PlayerState{
		{ID: "adj1"}, // hasn't played p1
	}

	ft := ClassifyFloater(p, true, adjacent, nil)
	if ft != FloaterTypeB {
		t.Errorf("already floated + has compatible opponent = TypeB, got %v", ft)
	}
}

func TestClassifyFloater_NotFloatedNoCompatible(t *testing.T) {
	p := &swisslib.PlayerState{ID: "p1", Opponents: []string{"adj1"}}
	adjacent := []*swisslib.PlayerState{
		{ID: "adj1", Opponents: []string{"p1"}},
	}

	ft := ClassifyFloater(p, false, adjacent, nil)
	if ft != FloaterTypeC {
		t.Errorf("not floated + no compatible = TypeC, got %v", ft)
	}
}

func TestClassifyFloater_NotFloatedWithCompatible(t *testing.T) {
	p := &swisslib.PlayerState{ID: "p1"}
	adjacent := []*swisslib.PlayerState{
		{ID: "adj1"},
	}

	ft := ClassifyFloater(p, false, adjacent, nil)
	if ft != FloaterTypeD {
		t.Errorf("not floated + has compatible = TypeD, got %v", ft)
	}
}

func TestSelectDownFloater_LowestPairingNumber(t *testing.T) {
	// Art. 3.2.4: if equal due colours, lowest TPN floats down.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 2.0, ColorHistory: nil},
		{ID: "b", TPN: 2, Score: 2.0, ColorHistory: nil},
		{ID: "c", TPN: 3, Score: 2.0, ColorHistory: nil},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	floater := SelectDownFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "a" {
		t.Errorf("expected 'a' (lowest TPN), got %v", floater)
	}
}

func TestSelectUpFloater_HighestPairingNumber(t *testing.T) {
	// Art. 3.2.4: highest TPN floats up.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 0.0, ColorHistory: nil},
		{ID: "b", TPN: 2, Score: 0.0, ColorHistory: nil},
		{ID: "c", TPN: 3, Score: 0.0, ColorHistory: nil},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	floater := SelectUpFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "c" {
		t.Errorf("expected 'c' (highest TPN), got %v", floater)
	}
}

func TestSelectDownFloater_ColorBalance(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Art. 3.2.2: select floater to equalise due colours.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 2.0, ColorHistory: []swisslib.Color{W}},       // due B
		{ID: "b", TPN: 2, Score: 2.0, ColorHistory: []swisslib.Color{W}},       // due B
		{ID: "c", TPN: 3, Score: 2.0, ColorHistory: []swisslib.Color{B}},       // due W
		{ID: "d", TPN: 4, Score: 2.0, ColorHistory: []swisslib.Color{B}},       // due W
		{ID: "e", TPN: 5, Score: 2.0, ColorHistory: []swisslib.Color{W, B, W}}, // due B
	}
	// 3 due B, 2 due W — float someone due B to balance.
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	floater := SelectDownFloater(players, adjacent, nil, nil, false)
	if floater == nil {
		t.Fatal("expected a floater")
	}
	// Floater should be due Black (to equalise)
	pref := swisslib.ComputeColorPreference(floater.ColorHistory)
	if pref.Color == nil || *pref.Color != B {
		t.Errorf("expected floater due Black for colour balance, got %v", floater.ID)
	}
}

func TestSelectDownFloater_AlreadyFloatedPreferred(t *testing.T) {
	// Art. 3.9: a player that has already floated (multi-hop) should be classified
	// as Type A/B (more disadvantaged) and thus preferred for further floating over
	// a player that has not floated yet, to avoid accumulating disadvantage.
	//
	// Setup: 3 players in a group, all with equal colour balance and no history.
	// Player "a" was already floated from a higher group (tracked in floatedIDs).
	// Player "b" and "c" are native to this group.
	// Adjacent group has compatible opponents for all.
	// Without floatedIDs tracking, all would be TypeD and "a" would win on lowest TPN.
	// With floatedIDs tracking, "a" is TypeB (already floated + compatible),
	// while "b" and "c" are TypeD (not floated + compatible).
	// TypeD > TypeB, so "b" (lowest TPN among TypeD) should be selected, not "a".
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 2.0},
		{ID: "b", TPN: 2, Score: 2.0},
		{ID: "c", TPN: 3, Score: 2.0},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}
	floatedIDs := map[string]bool{"a": true}

	floater := SelectDownFloater(players, adjacent, nil, floatedIDs, false)
	if floater == nil {
		t.Fatal("expected a floater")
	}
	// "a" is TypeB (already floated), "b" and "c" are TypeD (not floated).
	// TypeD is preferred (least disadvantage), so "b" should be selected (lowest TPN among TypeD).
	if floater.ID != "b" {
		t.Errorf("expected 'b' (TypeD, lowest TPN), got %v — already-floated player should not be preferred", floater.ID)
	}
}

func TestSelectUpFloater_AlreadyFloatedPreferred(t *testing.T) {
	// Same principle as above but for upward floating.
	// "c" was already floated, so it gets TypeB. "a" and "b" are TypeD.
	// Among TypeD, highest TPN wins for upward floating → "b" selected.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 0.0},
		{ID: "b", TPN: 2, Score: 0.0},
		{ID: "c", TPN: 3, Score: 0.0},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}
	floatedIDs := map[string]bool{"c": true}

	floater := SelectUpFloater(players, adjacent, nil, floatedIDs, false)
	if floater == nil {
		t.Fatal("expected a floater")
	}
	if floater.ID != "b" {
		t.Errorf("expected 'b' (TypeD, highest TPN among non-floated), got %v", floater.ID)
	}
}

func TestSelectDownFloater_MaxiRatingCap(t *testing.T) {
	// Art. 3.2.3: In maxi-tournament, if chosen floater's rating differs
	// from the lowest TPN player by >100, override to lowest TPN.
	// Setup: 3 players. Normal selection picks from majority due-colour group.
	// Make "b" the natural selection via colour balance, then verify override.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2100, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "b", TPN: 2, Rating: 2400, Score: 2.0, ColorHistory: []swisslib.Color{W}}, // due B
		{ID: "c", TPN: 3, Rating: 2350, Score: 2.0, ColorHistory: []swisslib.Color{W}}, // due B
	}
	// 1 due W, 2 due B → float someone due B to balance → "b" (lowest TPN among due-B).
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	// Without maxi: "b" selected (due B, lowest TPN in majority group).
	floater := SelectDownFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "b" {
		t.Fatalf("without maxi: expected 'b', got %v", floater)
	}

	// With maxi: "b" (rating 2400) vs reference "a" (TPN 1, rating 2100) → diff = 300 > 100 → override to "a".
	floater = SelectDownFloater(players, adjacent, nil, nil, true)
	if floater == nil || floater.ID != "a" {
		t.Errorf("with maxi: expected 'a' (override to lowest TPN due to >100 rating diff), got %v", floater)
	}
}

func TestSelectUpFloater_MaxiRatingCap(t *testing.T) {
	// Art. 3.2.3: When pairing upward in maxi, if chosen floater's rating
	// differs from the highest TPN player by >100, override to highest TPN.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2100, Score: 0.0, ColorHistory: []swisslib.Color{W}}, // due B
		{ID: "b", TPN: 2, Rating: 2400, Score: 0.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "c", TPN: 3, Rating: 2050, Score: 0.0, ColorHistory: []swisslib.Color{B}}, // due W
	}
	// 1 due B, 2 due W → float someone due W → highest TPN among due-W = "c".
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	// Without maxi: "c" selected (due W, highest TPN).
	floater := SelectUpFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "c" {
		t.Fatalf("without maxi: expected 'c', got %v", floater)
	}

	// With maxi: "c" (rating 2050) vs reference "c" (TPN 3 = highest, rating 2050) → same player, no override.
	// This case: the natural selection IS the reference. No override. Confirm "c" still selected.
	floater = SelectUpFloater(players, adjacent, nil, nil, true)
	if floater == nil || floater.ID != "c" {
		t.Errorf("with maxi: expected 'c' (no override needed, is highest TPN), got %v", floater)
	}
}

func TestSelectDownFloater_MaxiWithinCap(t *testing.T) {
	// Art. 3.2.3: If rating diff is within 100, no override.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2300, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "b", TPN: 2, Rating: 2350, Score: 2.0, ColorHistory: []swisslib.Color{W}}, // due B
		{ID: "c", TPN: 3, Rating: 2200, Score: 2.0, ColorHistory: []swisslib.Color{W}}, // due B
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	// With maxi: "b" (rating 2350) vs reference "a" (TPN 1, rating 2300) → diff = 50 <= 100 → no override.
	floater := SelectDownFloater(players, adjacent, nil, nil, true)
	if floater == nil || floater.ID != "b" {
		t.Errorf("with maxi within cap: expected 'b' (diff 50 <= 100, no override), got %v", floater)
	}
}

func TestSelectUpFloater_MaxiRatingCapOverride(t *testing.T) {
	// Art. 3.2.3: up-floater override fires when chosen player differs
	// from highest TPN player by >100 rating points.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2100, Score: 0.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "b", TPN: 2, Rating: 2400, Score: 0.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "c", TPN: 3, Rating: 2050, Score: 0.0, ColorHistory: []swisslib.Color{W}}, // due B
	}
	// 2 due W, 1 due B → float someone due W → highest TPN among due-W = "b" (TPN 2).
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	// Without maxi: "b" selected (due W, highest TPN in majority group).
	floater := SelectUpFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "b" {
		t.Fatalf("without maxi: expected 'b', got %v", floater)
	}

	// With maxi: "b" (rating 2400) vs reference "c" (TPN 3, rating 2050) → diff = 350 > 100 → override to "c".
	floater = SelectUpFloater(players, adjacent, nil, nil, true)
	if floater == nil || floater.ID != "c" {
		t.Errorf("with maxi: expected 'c' (override to highest TPN due to >100 rating diff), got %v", floater)
	}
}
