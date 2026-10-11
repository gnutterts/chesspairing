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
	// t1 (TPN 4, due B) is lowest TPN due B.
	if pairs[0][1].ID != "t1" {
		t.Errorf("expected floater paired with t1 (lowest TPN, due B), got %s", pairs[0][1].ID)
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
	// t3 (TPN 3, due W) is highest TPN due W.
	if pairs[0][1].ID != "t3" {
		t.Errorf("expected floater paired with t3 (highest TPN, due W), got %s", pairs[0][1].ID)
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
	// Art. 3.8 + 5.7: Maxi limits exchanges between proposed opponents.
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack

	floater := &swisslib.PlayerState{
		ID: "f1", TPN: 1, Rating: 2400, Score: 3.0,
		ColorHistory: []swisslib.Color{W}, // due B
	}
	floaters := []FloaterEntry{
		{Player: floater, Direction: FloatDown, SourceScore: 3.0},
	}
	// t1 is the first compatible opponent in Art. 3.8 order; its rating, and
	// the floater's rating, do not affect that first choice.
	targets := []*swisslib.PlayerState{
		{ID: "t1", TPN: 4, Rating: 2350, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W, within 100
		{ID: "t2", TPN: 5, Rating: 2100, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W, >100 diff
		{ID: "t3", TPN: 6, Rating: 2000, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W, >100 diff
	}

	// Maxi: t1 remains the first compatible opponent.
	pairsMaxi, _, _ := PairFloaters(floaters, targets, true, true, nil)
	if len(pairsMaxi) != 1 {
		t.Fatalf("maxi: expected 1 pair, got %d", len(pairsMaxi))
	}
	if pairsMaxi[0][1].ID != "t1" {
		t.Errorf("maxi: expected t1 (first compatible opponent), got %s", pairsMaxi[0][1].ID)
	}

	// Non-maxi: t1 is likewise first in Art. 3.8 order.
	pairsNonMaxi, _, _ := PairFloaters(floaters, targets, true, false, nil)
	if len(pairsNonMaxi) != 1 {
		t.Fatalf("non-maxi: expected 1 pair, got %d", len(pairsNonMaxi))
	}
	if pairsNonMaxi[0][1].ID != "t1" {
		t.Errorf("non-maxi: expected t1 (lowest TPN, no cap), got %s", pairsNonMaxi[0][1].ID)
	}
}

func TestPairFloaters_MaxiExchangeComparesOpponents(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	floater := &swisslib.PlayerState{ID: "f", TPN: 1, Rating: 2600, ColorHistory: []swisslib.Color{W}}
	// proposed is first in Art. 3.8 order, despite its rating being far from f.
	proposed := &swisslib.PlayerState{ID: "p", TPN: 2, Rating: 2000, ColorHistory: []swisslib.Color{W}}
	alternate := &swisslib.PlayerState{ID: "a", TPN: 3, Rating: 2090, ColorHistory: []swisslib.Color{B}}
	if got := selectFloaterOpponent(floater, FloatDown, []*swisslib.PlayerState{proposed, alternate}, true, nil); got != alternate {
		t.Fatalf("Maxi exchange within 100 points should select alternate, got %v", got)
	}
	alternate.Rating = 2200
	if got := selectFloaterOpponent(floater, FloatDown, []*swisslib.PlayerState{proposed, alternate}, true, nil); got != proposed {
		t.Fatalf("Maxi exchange over 100 points should retain proposed opponent, got %v", got)
	}
}

func TestPairFloaters_LimEvenRound2Order(t *testing.T) {
	// Art. 3.6.1 pairs the higher pairing number first. As Art. 4.2 calls
	// player #1 the highest numbered player, P04 gets the first target P07,
	// before P10 gets P15.
	floaters := []FloaterEntry{
		{Player: &swisslib.PlayerState{ID: "P10", TPN: 10}, Direction: FloatDown, SourceScore: 1},
		{Player: &swisslib.PlayerState{ID: "P04", TPN: 4}, Direction: FloatDown, SourceScore: 1},
	}
	targets := []*swisslib.PlayerState{
		{ID: "P07", TPN: 7},
		{ID: "P15", TPN: 15},
	}

	pairs, remainingFloaters, remainingTargets := PairFloaters(floaters, targets, true, false, nil)
	if len(remainingFloaters) != 0 || len(remainingTargets) != 0 {
		t.Fatalf("remaining floaters = %d, remaining targets = %d", len(remainingFloaters), len(remainingTargets))
	}
	if len(pairs) != 2 || pairs[0][0].ID != "P04" || pairs[0][1].ID != "P07" || pairs[1][0].ID != "P10" || pairs[1][1].ID != "P15" {
		t.Fatalf("pairs = %#v, want P04-P07 and P10-P15", pairs)
	}
}

func TestSortFloaters_DownFloaters_HighestPairingNumberFirst(t *testing.T) {
	// Art. 4.2 calls #1 the highest numbered player, so Art. 3.6.1 pairs
	// down-floaters by lowest TPN first.
	floaters := []FloaterEntry{
		{Player: &swisslib.PlayerState{ID: "a", TPN: 1}, Direction: FloatDown, SourceScore: 3.0},
		{Player: &swisslib.PlayerState{ID: "b", TPN: 3}, Direction: FloatDown, SourceScore: 3.0},
		{Player: &swisslib.PlayerState{ID: "c", TPN: 2}, Direction: FloatDown, SourceScore: 3.0},
	}

	sorted := sortFloaters(floaters, true)
	if sorted[0].Player.ID != "a" || sorted[1].Player.ID != "c" || sorted[2].Player.ID != "b" {
		t.Errorf("expected TPN order a(1), c(2), b(3), got %s, %s, %s",
			sorted[0].Player.ID, sorted[1].Player.ID, sorted[2].Player.ID)
	}
}

func TestSortFloaters_UpFloaters_LowestPairingNumberFirst(t *testing.T) {
	// Art. 4.2 calls #1 the highest numbered player, so Art. 3.7.1 pairs
	// up-floaters by highest TPN first.
	floaters := []FloaterEntry{
		{Player: &swisslib.PlayerState{ID: "a", TPN: 3}, Direction: FloatUp, SourceScore: 1.0},
		{Player: &swisslib.PlayerState{ID: "b", TPN: 1}, Direction: FloatUp, SourceScore: 1.0},
		{Player: &swisslib.PlayerState{ID: "c", TPN: 2}, Direction: FloatUp, SourceScore: 1.0},
	}

	sorted := sortFloaters(floaters, false)
	if sorted[0].Player.ID != "a" || sorted[1].Player.ID != "c" || sorted[2].Player.ID != "b" {
		t.Errorf("expected TPN order a(3), c(2), b(1), got %s, %s, %s",
			sorted[0].Player.ID, sorted[1].Player.ID, sorted[2].Player.ID)
	}
}

// TestUpFloaterRequiresAlternateDueColour: Art. 3.8 pairs the up-floater with
// the lowest numbered player "who is due the alternate colour"; P14 has no
// played game and so no due colour, P07 is due Black.
func TestUpFloaterRequiresAlternateDueColour(t *testing.T) {
	p03 := &swisslib.PlayerState{ID: "P03", TPN: 8, ColorHistory: []swisslib.Color{swisslib.ColorBlack}}
	p07 := &swisslib.PlayerState{ID: "P07", TPN: 5, ColorHistory: []swisslib.Color{swisslib.ColorWhite}}
	p14 := &swisslib.PlayerState{ID: "P14", TPN: 7}
	pairs, _, _ := PairFloaters(
		[]FloaterEntry{{Player: p03, Direction: FloatUp, SourceScore: 0.5}},
		[]*swisslib.PlayerState{p07, p14},
		false, false, nil,
	)
	if len(pairs) != 1 || (pairs[0][0].ID != "P07" && pairs[0][1].ID != "P07") {
		t.Fatalf("Art. 3.8: expected P03-P07, got %v", pairs)
	}
}

// TestFloaterKeepsFirstOpponentWhenNobodyIsDueAlternateColour: without any
// player due the alternate colour, the first available opponent is kept.
func TestFloaterKeepsFirstOpponentWhenNobodyIsDueAlternateColour(t *testing.T) {
	floater := &swisslib.PlayerState{ID: "F", TPN: 1, ColorHistory: []swisslib.Color{swisslib.ColorBlack}}
	first := &swisslib.PlayerState{ID: "A", TPN: 2}
	second := &swisslib.PlayerState{ID: "B", TPN: 3, ColorHistory: []swisslib.Color{swisslib.ColorBlack}}
	pairs, _, _ := PairFloaters(
		[]FloaterEntry{{Player: floater, Direction: FloatDown, SourceScore: 2}},
		[]*swisslib.PlayerState{first, second},
		true, false, nil,
	)
	if len(pairs) != 1 || (pairs[0][0].ID != "A" && pairs[0][1].ID != "A") {
		t.Fatalf("expected F-A, got %v", pairs)
	}
}
