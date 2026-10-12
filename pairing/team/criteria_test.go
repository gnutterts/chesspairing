// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestC8Cost_NoColorPrefs(t *testing.T) {
	// When colour preferences are disabled, C8 cost is 0.
	a := &lexswiss.ParticipantState{ID: "t1"}
	b := &lexswiss.ParticipantState{ID: "t2"}
	if got := C8Cost(a, b, ColorPrefTypeNone, false); got != 0 {
		t.Errorf("C8Cost = %d, want 0", got)
	}
}

func TestC8Cost_TypeA_OppositePrefs(t *testing.T) {
	// Type A: one wants White, other wants Black → 0 unfulfilled.
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
	if got := C8Cost(a, b, ColorPrefTypeA, false); got != 0 {
		t.Errorf("opposite prefs: C8Cost = %d, want 0", got)
	}
}

func TestC8Cost_TypeA_SamePrefs(t *testing.T) {
	// Type A: both want White → C8 unfulfilled = 1.
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	if got := C8Cost(a, b, ColorPrefTypeA, false); got != 1 {
		t.Errorf("same prefs (both White): C8Cost = %d, want 1", got)
	}
}

func TestC8Cost_TypeA_OnePrefOneNone(t *testing.T) {
	// One has preference, other has none → 0 unfulfilled.
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack},
	}
	if got := C8Cost(a, b, ColorPrefTypeA, false); got != 0 {
		t.Errorf("one pref + one none: C8Cost = %d, want 0", got)
	}
}

func TestC8C9Cost_TypeB_SameStrongPrefs(t *testing.T) {
	// Type B: both have strong White → C8 = 1, C9 = 1.
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	if got := C8Cost(a, b, ColorPrefTypeB, false); got != 1 {
		t.Errorf("both strong White: C8Cost = %d, want 1", got)
	}
	if got := C9Cost(a, b, ColorPrefTypeB, false); got != 1 {
		t.Errorf("both strong White: C9Cost = %d, want 1", got)
	}
}

func TestC8C9Cost_TypeB_StrongVsMild_SameColor(t *testing.T) {
	// One strong White + one mild White → C8 = 1, C9 = 0.
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack},
	}
	if got := C8Cost(a, b, ColorPrefTypeB, false); got != 1 {
		t.Errorf("strong+mild same colour: C8Cost = %d, want 1", got)
	}
	if got := C9Cost(a, b, ColorPrefTypeB, false); got != 0 {
		t.Errorf("strong+mild same colour: C9Cost = %d, want 0", got)
	}
}

func TestC8C9Cost_TypeB_StrongVsMild_OppositeColor(t *testing.T) {
	// Strong White + mild Black → both satisfied.
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite},
	}
	if got := C8Cost(a, b, ColorPrefTypeB, false); got != 0 {
		t.Errorf("strong White + mild Black: C8Cost = %d, want 0", got)
	}
	if got := C9Cost(a, b, ColorPrefTypeB, false); got != 0 {
		t.Errorf("strong White + mild Black: C9Cost = %d, want 0", got)
	}
}

func TestC8Cost_LastTwoRounds_Relaxed(t *testing.T) {
	// In the last two rounds, C10 is off, but C8 is still charged per pair.
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	if got := C8Cost(a, b, ColorPrefTypeA, false); got != 1 {
		t.Errorf("same prefs in last 2 rounds: C8Cost = %d, want 1", got)
	}
}

func TestC8Cost_NoBothNone(t *testing.T) {
	// Both have no preference → C8 = 0.
	a := &lexswiss.ParticipantState{ID: "t1"}
	b := &lexswiss.ParticipantState{ID: "t2"}
	if got := C8Cost(a, b, ColorPrefTypeA, false); got != 0 {
		t.Errorf("both no pref: C8Cost = %d, want 0", got)
	}
}

func TestBracketPairCost_LastTwoRounds_C8StillCharged(t *testing.T) {
	// In the last two rounds, C10 is off but the combined bracket cost still
	// charges C8: a pair of teams with the same Type A preference costs W^2
	// with W = bracketSize + 1 (here 3), i.e. the maximum C10 weight sum
	// cannot outweigh it.
	a := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	b := &lexswiss.ParticipantState{
		ID:           "t2",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	cost, err := BracketPairCost(2, nil, ColorPrefTypeA, true /*isLastTwoRounds*/, false)
	if err != nil {
		t.Fatalf("BracketPairCost() error: %v", err)
	}
	if got := cost(a, b); got != 9 {
		t.Errorf("combined cost with isLastTwoRounds = %d, want 9 (W^2)", got)
	}
}
