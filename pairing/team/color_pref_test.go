// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

// --- Type A tests ---

func TestTypeA_NoHistory(t *testing.T) {
	p := &lexswiss.ParticipantState{ID: "t1"}
	pref := ComputeColorPreference(p, ColorPrefTypeA, false)
	if pref != ColorPrefNone {
		t.Errorf("no history: expected None, got %v", pref)
	}
}

func TestTypeA_CDNegative2(t *testing.T) {
	// CD = -2 → preference for White
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeA, false)
	if pref != ColorPrefWhite {
		t.Errorf("CD=-2: expected White, got %v", pref)
	}
}

func TestTypeA_CDPositive2(t *testing.T) {
	// CD = +2 → preference for Black
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorWhite},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeA, false)
	if pref != ColorPrefBlack {
		t.Errorf("CD=+2: expected Black, got %v", pref)
	}
}

func TestTypeA_CDZero_LastTwoBlack(t *testing.T) {
	// W, B, W, B, B → whites=2, blacks=3, CD = -1.
	// CD = -1, last two = B, B → preference for White.
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeA, false)
	if pref != ColorPrefWhite {
		t.Errorf("CD=-1, last two Black: expected White, got %v", pref)
	}
}

func TestTypeA_CDMinus1_LastTwoBlack(t *testing.T) {
	// CD = -1 and last two played = Black → preference for White
	// W, B, B → whites=1, blacks=2, CD=-1. Last two = B, B. Yes.
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeA, false)
	if pref != ColorPrefWhite {
		t.Errorf("CD=-1, last two Black: expected White, got %v", pref)
	}
}

func TestTypeA_CDPlus1_LastTwoWhite(t *testing.T) {
	// CD = +1 and last two played = White → preference for Black
	// B, W, W → whites=2, blacks=1, CD=+1. Last two = W, W. Yes.
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite, lexswiss.ColorWhite},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeA, false)
	if pref != ColorPrefBlack {
		t.Errorf("CD=+1, last two White: expected Black, got %v", pref)
	}
}

func TestTypeA_CDMinus1_MixedLast(t *testing.T) {
	// CD = -1, last two = W, B → no preference (last two not same)
	// B, W, B → whites=1, blacks=2, CD=-1. Last two = W, B. Mixed.
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite, lexswiss.ColorBlack},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeA, false)
	if pref != ColorPrefNone {
		t.Errorf("CD=-1, mixed last: expected None, got %v", pref)
	}
}

// --- Type B tests ---

func TestTypeB_StrongWhite(t *testing.T) {
	// Same conditions as Type A White → strong preference.
	// W, B, B → CD=-1, last two Black → strong White
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeB, false)
	if pref != ColorPrefStrongWhite {
		t.Errorf("Type B strong White: expected StrongWhite, got %v", pref)
	}
}

func TestTypeB_StrongBlack(t *testing.T) {
	// B, W, W → CD=+1, last two White → strong Black
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite, lexswiss.ColorWhite},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeB, false)
	if pref != ColorPrefStrongBlack {
		t.Errorf("Type B strong Black: expected StrongBlack, got %v", pref)
	}
}

func TestTypeB_MildWhite_CDMinus1(t *testing.T) {
	// CD = -1, last two NOT both Black → mild White (CD is -1)
	// B → CD=-1. Only 1 game, so last two can't be both Black.
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeB, false)
	if pref != ColorPrefMildWhite {
		t.Errorf("Type B mild White (CD=-1): expected MildWhite, got %v", pref)
	}
}

func TestTypeB_MildBlack_CDPlus1(t *testing.T) {
	// CD = +1, last two NOT both White → mild Black (CD is +1)
	// W → CD=+1. Only 1 game.
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeB, false)
	if pref != ColorPrefMildBlack {
		t.Errorf("Type B mild Black (CD=+1): expected MildBlack, got %v", pref)
	}
}

func TestTypeB_MildWhite_CDZero_NotLastRound(t *testing.T) {
	// CD = 0, not last round, last played = Black → mild White
	// W, B → CD=0. Last = Black. Not last round.
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeB, false)
	if pref != ColorPrefMildWhite {
		t.Errorf("Type B mild White (CD=0, last=B): expected MildWhite, got %v", pref)
	}
}

func TestTypeB_MildBlack_CDZero_NotLastRound(t *testing.T) {
	// CD = 0, not last round, last played = White → mild Black
	// B, W → CD=0. Last = White. Not last round.
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeB, false)
	if pref != ColorPrefMildBlack {
		t.Errorf("Type B mild Black (CD=0, last=W): expected MildBlack, got %v", pref)
	}
}

func TestTypeB_None_CDZero_LastRound(t *testing.T) {
	// CD = 0, IS last round → no preference (Art. 1.7.2.5)
	// B, W → CD=0. Last round = true.
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeB, true)
	if pref != ColorPrefNone {
		t.Errorf("Type B CD=0, last round: expected None, got %v", pref)
	}
}

func TestTypeB_NoHistory(t *testing.T) {
	// No games played → no preference (Art. 1.7.2.5)
	p := &lexswiss.ParticipantState{ID: "t1"}
	pref := ComputeColorPreference(p, ColorPrefTypeB, false)
	if pref != ColorPrefNone {
		t.Errorf("Type B no history: expected None, got %v", pref)
	}
}

// --- Type None tests ---

func TestTypeNone_AlwaysNone(t *testing.T) {
	p := &lexswiss.ParticipantState{
		ID:           "t1",
		ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack, lexswiss.ColorBlack},
	}
	pref := ComputeColorPreference(p, ColorPrefTypeNone, false)
	if pref != ColorPrefNone {
		t.Errorf("Type None: expected None, got %v", pref)
	}
}

// --- Helper tests ---

func TestColorDifference(t *testing.T) {
	tests := []struct {
		name    string
		history []lexswiss.Color
		want    int
	}{
		{"empty", nil, 0},
		{"single white", []lexswiss.Color{lexswiss.ColorWhite}, 1},
		{"single black", []lexswiss.Color{lexswiss.ColorBlack}, -1},
		{"equal", []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack}, 0},
		{"with none", []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorNone, lexswiss.ColorBlack}, 0},
		{"positive 2", []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorWhite}, 2},
		{"negative 2", []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack}, -2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorDifference(tt.history)
			if got != tt.want {
				t.Errorf("colorDifference(%v) = %d, want %d", tt.history, got, tt.want)
			}
		})
	}
}

func TestLastNPlayed(t *testing.T) {
	tests := []struct {
		name    string
		history []lexswiss.Color
		n       int
		want    []lexswiss.Color
	}{
		{"empty", nil, 2, nil},
		{"one from two", []lexswiss.Color{lexswiss.ColorWhite}, 2, []lexswiss.Color{lexswiss.ColorWhite}},
		{"two from three", []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorWhite}, 2, []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite}},
		{"skip none", []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorNone, lexswiss.ColorBlack}, 2, []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lastNPlayed(tt.history, tt.n)
			if len(got) != len(tt.want) {
				t.Errorf("lastNPlayed len=%d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("lastNPlayed[%d]=%v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
