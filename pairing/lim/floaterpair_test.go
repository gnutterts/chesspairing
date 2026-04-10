// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestPairFloaters_DownFloaterHighestTarget(t *testing.T) {
	// Art. 3.8: down-floater paired with highest TPN available player.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack

	floater := &swisslib.PlayerState{
		ID: "f1", TPN: 1, Rating: 2400, Score: 3.0,
		ColorHistory: []swisslib.Color{W, B}, // due W
	}
	floaters := []FloaterEntry{
		{Player: floater, Direction: FloatDown, SourceScore: 3.0},
	}
	targets := []*swisslib.PlayerState{
		{ID: "t1", TPN: 4, Rating: 2200, Score: 2.0, ColorHistory: []swisslib.Color{B, W}}, // due B
		{ID: "t2", TPN: 5, Rating: 2100, Score: 2.0, ColorHistory: []swisslib.Color{W, B}}, // due W
		{ID: "t3", TPN: 6, Rating: 2000, Score: 2.0, ColorHistory: []swisslib.Color{B, W}}, // due B
	}

	pairs, remF, remT := PairFloaters(floaters, targets, true, false, nil)

	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	// Floater (due W) should pair with highest TPN target due alternate colour (B).
	// t3 (TPN 6, due B) is highest TPN due B.
	if pairs[0][1].ID != "t3" {
		t.Errorf("expected floater paired with t3 (highest TPN, due B), got %s", pairs[0][1].ID)
	}
	if len(remF) != 0 {
		t.Errorf("expected 0 remaining floaters, got %d", len(remF))
	}
	if len(remT) != 2 {
		t.Errorf("expected 2 remaining targets, got %d", len(remT))
	}
}

func TestPairFloaters_UpFloaterLowestTarget(t *testing.T) {
	// Art. 3.8: up-floater paired with lowest TPN available player.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack

	floater := &swisslib.PlayerState{
		ID: "f1", TPN: 6, Rating: 2000, Score: 1.0,
		ColorHistory: []swisslib.Color{B, W}, // due B
	}
	floaters := []FloaterEntry{
		{Player: floater, Direction: FloatUp, SourceScore: 1.0},
	}
	targets := []*swisslib.PlayerState{
		{ID: "t1", TPN: 1, Rating: 2400, Score: 2.0, ColorHistory: []swisslib.Color{W, B}}, // due W
		{ID: "t2", TPN: 2, Rating: 2300, Score: 2.0, ColorHistory: []swisslib.Color{B, W}}, // due B
		{ID: "t3", TPN: 3, Rating: 2200, Score: 2.0, ColorHistory: []swisslib.Color{W, B}}, // due W
	}

	pairs, remF, remT := PairFloaters(floaters, targets, false, false, nil)

	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	// Floater (due B) should pair with lowest TPN target due alternate colour (W).
	// t1 (TPN 1, due W) is lowest TPN due W.
	if pairs[0][1].ID != "t1" {
		t.Errorf("expected floater paired with t1 (lowest TPN, due W), got %s", pairs[0][1].ID)
	}
	if len(remF) != 0 {
		t.Errorf("expected 0 remaining floaters, got %d", len(remF))
	}
	if len(remT) != 2 {
		t.Errorf("expected 2 remaining targets, got %d", len(remT))
	}
}

func TestPairFloaters_OrderDFBeforeUF_UpperHalf(t *testing.T) {
	// Art. 3.6.3: in upper half, DF paired first, then UF.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack

	df := &swisslib.PlayerState{
		ID: "df1", TPN: 2, Rating: 2300, Score: 3.0,
		ColorHistory: []swisslib.Color{W}, // due B
	}
	uf := &swisslib.PlayerState{
		ID: "uf1", TPN: 5, Rating: 2100, Score: 1.0,
		ColorHistory: []swisslib.Color{B}, // due W
	}
	floaters := []FloaterEntry{
		{Player: uf, Direction: FloatUp, SourceScore: 1.0},
		{Player: df, Direction: FloatDown, SourceScore: 3.0},
	}
	targets := []*swisslib.PlayerState{
		{ID: "t1", TPN: 3, Rating: 2250, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "t2", TPN: 4, Rating: 2200, Score: 2.0, ColorHistory: []swisslib.Color{W}}, // due B
	}

	pairs, remF, remT := PairFloaters(floaters, targets, true, false, nil)

	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if len(remF) != 0 {
		t.Errorf("expected 0 remaining floaters, got %d", len(remF))
	}
	if len(remT) != 0 {
		t.Errorf("expected 0 remaining targets, got %d", len(remT))
	}

	// DF should be paired first (gets first pick of targets).
	// df (due B) should get target due alternate (W). t1 (TPN 3, due W) is the
	// only one due W, but Art. 3.8 says highest TPN for DF — t1 is the only
	// option due W. So df paired with t1.
	// uf (due W) then gets t2 (TPN 4, due B).
	if pairs[0][0].ID != "df1" {
		t.Errorf("expected DF paired first, got %s", pairs[0][0].ID)
	}
}

func TestPairFloaters_OrderUFBeforeDF_LowerHalf(t *testing.T) {
	// Art. 3.7.3: in lower half, UF paired first, then DF.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack

	df := &swisslib.PlayerState{
		ID: "df1", TPN: 2, Rating: 2300, Score: 2.0,
		ColorHistory: []swisslib.Color{W}, // due B
	}
	uf := &swisslib.PlayerState{
		ID: "uf1", TPN: 5, Rating: 2100, Score: 0.0,
		ColorHistory: []swisslib.Color{B}, // due W
	}
	floaters := []FloaterEntry{
		{Player: df, Direction: FloatDown, SourceScore: 2.0},
		{Player: uf, Direction: FloatUp, SourceScore: 0.0},
	}
	targets := []*swisslib.PlayerState{
		{ID: "t1", TPN: 3, Rating: 2250, Score: 1.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "t2", TPN: 4, Rating: 2200, Score: 1.0, ColorHistory: []swisslib.Color{W}}, // due B
	}

	pairs, remF, remT := PairFloaters(floaters, targets, false, false, nil)

	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if len(remF) != 0 {
		t.Errorf("expected 0 remaining floaters, got %d", len(remF))
	}
	if len(remT) != 0 {
		t.Errorf("expected 0 remaining targets, got %d", len(remT))
	}

	// UF should be paired first in lower half.
	if pairs[0][0].ID != "uf1" {
		t.Errorf("expected UF paired first in lower half, got %s", pairs[0][0].ID)
	}
}

func TestPairFloaters_MaxiRatingConstraint(t *testing.T) {
	// Art. 3.8 + 5.7: In maxi mode, candidates with >100 rating diff are skipped.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack

	floater := &swisslib.PlayerState{
		ID: "f1", TPN: 1, Rating: 2400, Score: 3.0,
		ColorHistory: []swisslib.Color{W}, // due B
	}
	floaters := []FloaterEntry{
		{Player: floater, Direction: FloatDown, SourceScore: 3.0},
	}
	// t1 is within 100 points of floater (2400-2350=50). t2 and t3 are too far.
	// In non-maxi: highest TPN due W = t3 (TPN 6). In maxi: only t1 is eligible.
	targets := []*swisslib.PlayerState{
		{ID: "t1", TPN: 4, Rating: 2350, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W, within 100
		{ID: "t2", TPN: 5, Rating: 2100, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W, >100 diff
		{ID: "t3", TPN: 6, Rating: 2000, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W, >100 diff
	}

	// Maxi: only t1 (rating 2350, |2400-2350|=50 <= 100) is eligible.
	pairsMaxi, _, _ := PairFloaters(floaters, targets, true, true, nil)
	if len(pairsMaxi) != 1 {
		t.Fatalf("maxi: expected 1 pair, got %d", len(pairsMaxi))
	}
	if pairsMaxi[0][1].ID != "t1" {
		t.Errorf("maxi: expected t1 (only eligible within 100-pt cap), got %s", pairsMaxi[0][1].ID)
	}

	// Non-maxi: highest TPN due W = t3 (TPN 6), no rating constraint.
	pairsNonMaxi, _, _ := PairFloaters(floaters, targets, true, false, nil)
	if len(pairsNonMaxi) != 1 {
		t.Fatalf("non-maxi: expected 1 pair, got %d", len(pairsNonMaxi))
	}
	if pairsNonMaxi[0][1].ID != "t3" {
		t.Errorf("non-maxi: expected t3 (highest TPN, no cap), got %s", pairsNonMaxi[0][1].ID)
	}
}

func TestSortFloaters_DownFloaters_HighestTPNFirst(t *testing.T) {
	// Art. 3.6.1: down-floaters paired by highest TPN first.
	floaters := []FloaterEntry{
		{Player: &swisslib.PlayerState{ID: "a", TPN: 1}, Direction: FloatDown, SourceScore: 3.0},
		{Player: &swisslib.PlayerState{ID: "b", TPN: 3}, Direction: FloatDown, SourceScore: 3.0},
		{Player: &swisslib.PlayerState{ID: "c", TPN: 2}, Direction: FloatDown, SourceScore: 3.0},
	}

	sorted := sortFloaters(floaters, true)
	if sorted[0].Player.ID != "b" || sorted[1].Player.ID != "c" || sorted[2].Player.ID != "a" {
		t.Errorf("expected TPN order b(3), c(2), a(1), got %s, %s, %s",
			sorted[0].Player.ID, sorted[1].Player.ID, sorted[2].Player.ID)
	}
}

func TestSortFloaters_UpFloaters_LowestTPNFirst(t *testing.T) {
	// Art. 3.7.1: up-floaters paired by lowest TPN first.
	floaters := []FloaterEntry{
		{Player: &swisslib.PlayerState{ID: "a", TPN: 3}, Direction: FloatUp, SourceScore: 1.0},
		{Player: &swisslib.PlayerState{ID: "b", TPN: 1}, Direction: FloatUp, SourceScore: 1.0},
		{Player: &swisslib.PlayerState{ID: "c", TPN: 2}, Direction: FloatUp, SourceScore: 1.0},
	}

	sorted := sortFloaters(floaters, false)
	if sorted[0].Player.ID != "b" || sorted[1].Player.ID != "c" || sorted[2].Player.ID != "a" {
		t.Errorf("expected TPN order b(1), c(2), a(3), got %s, %s, %s",
			sorted[0].Player.ID, sorted[1].Player.ID, sorted[2].Player.ID)
	}
}
