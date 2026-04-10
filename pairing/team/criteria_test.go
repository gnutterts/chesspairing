// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestBuildCriteriaFunc_NoColorPrefs(t *testing.T) {
	// When colour preferences are disabled, criteria always pass.
	fn := BuildCriteriaFunc(ColorPrefTypeNone, false, false)
	a := &lexswiss.ParticipantState{ID: "t1"}
	b := &lexswiss.ParticipantState{ID: "t2"}
	if fn != nil {
		if !fn(a, b) {
			t.Error("no color prefs: criteria should pass")
		}
	}
}

func TestBuildCriteriaFunc_TypeA_OppositePrefs(t *testing.T) {
	// Type A: one wants White, other wants Black → should pass.
	fn := BuildCriteriaFunc(ColorPrefTypeA, false, false)
	// t1: W, B, B → CD=-1, last two=B,B → pref White
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	// t2: B, W, W → CD=+1, last two=W,W → pref Black
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite, lexswiss.ColorWhite},
	}
	if fn == nil {
		t.Fatal("expected non-nil criteria func")
	}
	if !fn(a, b) {
		t.Error("opposite prefs: should pass C8")
	}
}

func TestBuildCriteriaFunc_TypeA_SamePrefs(t *testing.T) {
	// Type A: both want White → C8 violation (one can't get preference).
	fn := BuildCriteriaFunc(ColorPrefTypeA, false, false)
	// Both: W, B, B → CD=-1, last two=B,B → pref White
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	if fn == nil {
		t.Fatal("expected non-nil criteria func")
	}
	if fn(a, b) {
		t.Error("same prefs (both White): should fail C8")
	}
}

func TestBuildCriteriaFunc_TypeA_OnePrefOneNone(t *testing.T) {
	// One has preference, other has none → should pass (C8 satisfied for both).
	fn := BuildCriteriaFunc(ColorPrefTypeA, false, false)
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack},
	}
	if fn == nil {
		t.Fatal("expected non-nil criteria func")
	}
	if !fn(a, b) {
		t.Error("one pref + one none: should pass C8")
	}
}

func TestBuildCriteriaFunc_TypeB_SameStrongPrefs(t *testing.T) {
	// Type B: both have strong White → C9 violation.
	fn := BuildCriteriaFunc(ColorPrefTypeB, false, false)
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	if fn == nil {
		t.Fatal("expected non-nil criteria func")
	}
	if fn(a, b) {
		t.Error("both strong White: should fail C9")
	}
}

func TestBuildCriteriaFunc_TypeB_StrongVsMild_SameColor(t *testing.T) {
	// One strong White + one mild White → C8 violation (one can't get White).
	// But C9 is satisfied (only one strong, it can be granted).
	fn := BuildCriteriaFunc(ColorPrefTypeB, false, false)
	// t1: strong White (CD=-1, last two Black)
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	// t2: mild White (CD=-1, last two NOT both Black)
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack},
	}
	if fn == nil {
		t.Fatal("expected non-nil criteria func")
	}
	// Both want White (strong + mild) → C8 violation.
	if fn(a, b) {
		t.Error("strong+mild same colour: should fail C8")
	}
}

func TestBuildCriteriaFunc_TypeB_StrongVsMild_OppositeColor(t *testing.T) {
	// Strong White + mild Black → both satisfied.
	fn := BuildCriteriaFunc(ColorPrefTypeB, false, false)
	// t1: strong White
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	// t2: mild Black (CD=+1)
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite},
	}
	if fn == nil {
		t.Fatal("expected non-nil criteria func")
	}
	if !fn(a, b) {
		t.Error("strong White + mild Black: should pass")
	}
}

func TestBuildCriteriaFunc_LastTwoRounds_Relaxed(t *testing.T) {
	// In the last two rounds, C10 is relaxed. C8/C9 are NOT relaxed by round count.
	// But C7 and C10 are relaxed. Since C8/C9 are still checked, this test
	// verifies that C8/C9 are still enforced in last two rounds.
	fn := BuildCriteriaFunc(ColorPrefTypeA, true, false)
	// Both want White → C8 violation even in last 2 rounds.
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	if fn == nil {
		t.Fatal("expected non-nil criteria func")
	}
	if fn(a, b) {
		t.Error("same prefs in last 2 rounds: C8 should still fail")
	}
}

func TestBuildCriteriaFunc_NoBothNone(t *testing.T) {
	// Both have no preference → always passes.
	fn := BuildCriteriaFunc(ColorPrefTypeA, false, false)
	a := &lexswiss.ParticipantState{ID: "t1"}
	b := &lexswiss.ParticipantState{ID: "t2"}
	if fn == nil {
		t.Fatal("expected non-nil criteria func")
	}
	if !fn(a, b) {
		t.Error("both no pref: should pass")
	}
}
