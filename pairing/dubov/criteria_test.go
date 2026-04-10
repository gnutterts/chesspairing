// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestMaxT(t *testing.T) {
	tests := []struct {
		completedRounds int
		want            int
	}{
		{0, 2},  // 2 + 0/5
		{1, 2},  // 2 + 1/5
		{4, 2},  // 2 + 4/5
		{5, 3},  // 2 + 5/5
		{10, 4}, // 2 + 10/5
		{15, 5}, // 2 + 15/5
	}
	for _, tt := range tests {
		got := MaxT(tt.completedRounds)
		if got != tt.want {
			t.Errorf("MaxT(%d) = %d, want %d", tt.completedRounds, got, tt.want)
		}
	}
}

func TestUpfloatCount(t *testing.T) {
	p := &swisslib.PlayerState{
		ID:           "a",
		FloatHistory: []swisslib.Float{swisslib.FloatUp, swisslib.FloatNone, swisslib.FloatUp},
	}
	got := UpfloatCount(p)
	if got != 2 {
		t.Errorf("UpfloatCount = %d, want 2", got)
	}
}

func TestCriterionC1_NoRematches(t *testing.T) {
	a := &swisslib.PlayerState{ID: "a", Opponents: []string{"b"}}
	b := &swisslib.PlayerState{ID: "b", Opponents: []string{"a"}}
	c := &swisslib.PlayerState{ID: "c", Opponents: []string{}}

	if C1NoRematches(a, b) {
		t.Error("a and b already played - C1 should return false")
	}
	if !C1NoRematches(a, c) {
		t.Error("a and c never played - C1 should return true")
	}
}

func TestCriterionC3_AbsoluteColorConflict(t *testing.T) {
	// Both absolute white preference -> conflict
	a := &swisslib.PlayerState{
		ID:           "a",
		ColorHistory: []swisslib.Color{swisslib.ColorBlack, swisslib.ColorBlack, swisslib.ColorBlack},
	}
	c := &swisslib.PlayerState{
		ID:           "c",
		ColorHistory: []swisslib.Color{swisslib.ColorBlack, swisslib.ColorBlack, swisslib.ColorBlack},
	}

	if C3NoAbsoluteColorConflict(a, c) {
		t.Error("both have absolute white preference - C3 should return false")
	}

	// One absolute white, one absolute black -> OK
	d := &swisslib.PlayerState{
		ID:           "d",
		ColorHistory: []swisslib.Color{swisslib.ColorWhite, swisslib.ColorWhite, swisslib.ColorWhite},
	}
	if !C3NoAbsoluteColorConflict(a, d) {
		t.Error("different absolute preferences - C3 should return true")
	}

	// One no preference -> OK
	e := &swisslib.PlayerState{
		ID:           "e",
		ColorHistory: nil,
	}
	if !C3NoAbsoluteColorConflict(a, e) {
		t.Error("e has no preference - C3 should return true")
	}
}

func TestDubovCandidateScore_Compare(t *testing.T) {
	better := DubovCandidateScore{
		UpfloaterCount: 0,
		Violations:     [NumDubovViolations]int{0, 0, 0, 0, 0, 0, 0},
	}
	worse := DubovCandidateScore{
		UpfloaterCount: 1,
		Violations:     [NumDubovViolations]int{0, 0, 0, 0, 0, 0, 0},
	}

	if better.Compare(&worse) != -1 {
		t.Error("fewer upfloaters should be better")
	}
	if worse.Compare(&better) != 1 {
		t.Error("more upfloaters should be worse")
	}
	if better.Compare(&better) != 0 {
		t.Error("identical should be equal")
	}
}

func TestDubovCandidateScore_C5HigherIsBetter(t *testing.T) {
	a := DubovCandidateScore{
		UpfloaterCount:    1,
		UpfloaterScoreSum: 3.0,
		Violations:        [NumDubovViolations]int{1, 0, 0, 0, 0, 0, 0},
	}
	b := DubovCandidateScore{
		UpfloaterCount:    1,
		UpfloaterScoreSum: 2.0,
		Violations:        [NumDubovViolations]int{1, 0, 0, 0, 0, 0, 0},
	}

	if a.Compare(&b) != -1 {
		t.Error("higher upfloater score sum should be better")
	}
}

func TestDubovCandidateScore_IsPerfect(t *testing.T) {
	perfect := DubovCandidateScore{}
	if !perfect.IsPerfect() {
		t.Error("zero values should be perfect")
	}

	imperfect := DubovCandidateScore{UpfloaterCount: 1}
	if imperfect.IsPerfect() {
		t.Error("non-zero upfloater count should not be perfect")
	}

	imperfect2 := DubovCandidateScore{Violations: [NumDubovViolations]int{0, 0, 1, 0, 0, 0, 0}}
	if imperfect2.IsPerfect() {
		t.Error("non-zero violation should not be perfect")
	}
}

func TestCriterionC6_ColorPrefViolations(t *testing.T) {
	// Both players get their preference -> 0 violations
	pairs := []proposedPairing{
		{
			white: &swisslib.PlayerState{
				ID:           "a",
				ColorHistory: []swisslib.Color{swisslib.ColorBlack}, // prefers white
			},
			black: &swisslib.PlayerState{
				ID:           "b",
				ColorHistory: []swisslib.Color{swisslib.ColorWhite}, // prefers black
			},
		},
	}

	got := CriterionC6(pairs)
	if got != 0 {
		t.Errorf("expected 0 violations, got %d", got)
	}

	// White player actually prefers black -> 1 violation
	pairs2 := []proposedPairing{
		{
			white: &swisslib.PlayerState{
				ID:           "a",
				ColorHistory: []swisslib.Color{swisslib.ColorWhite}, // prefers black, but gets white
			},
			black: &swisslib.PlayerState{
				ID:           "b",
				ColorHistory: []swisslib.Color{swisslib.ColorWhite}, // prefers black
			},
		},
	}

	got2 := CriterionC6(pairs2)
	if got2 != 1 {
		t.Errorf("expected 1 violation, got %d", got2)
	}
}

func TestCriterionC7_MaxTViolations(t *testing.T) {
	floaters := []*swisslib.PlayerState{
		{ID: "a", FloatHistory: []swisslib.Float{swisslib.FloatUp, swisslib.FloatUp}}, // 2 upfloats
		{ID: "b", FloatHistory: []swisslib.Float{swisslib.FloatUp}},                   // 1 upfloat
	}

	// MaxT = 2: player a is at limit
	got := CriterionC7(floaters, 2)
	if got != 1 {
		t.Errorf("expected 1 MaxT violation (player a), got %d", got)
	}

	// MaxT = 3: neither at limit
	got2 := CriterionC7(floaters, 3)
	if got2 != 0 {
		t.Errorf("expected 0 MaxT violations, got %d", got2)
	}
}

func TestCriterionC8_ConsecutiveUpfloats(t *testing.T) {
	floaters := []*swisslib.PlayerState{
		{ID: "a", FloatHistory: []swisslib.Float{swisslib.FloatUp}},                     // upfloated last round
		{ID: "b", FloatHistory: []swisslib.Float{swisslib.FloatNone}},                   // did not upfloat
		{ID: "c", FloatHistory: []swisslib.Float{swisslib.FloatDown, swisslib.FloatUp}}, // upfloated last round
	}

	got := CriterionC8(floaters)
	if got != 2 {
		t.Errorf("expected 2 consecutive upfloaters (a, c), got %d", got)
	}
}
