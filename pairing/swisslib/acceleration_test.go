// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "testing"

func TestBakuAccelerationRounds(t *testing.T) {
	tests := []struct {
		totalRounds            int
		wantAccelerated        int
		wantFullVP, wantHalfVP int
	}{
		{5, 3, 2, 1},
		{7, 4, 2, 2},
		{9, 5, 3, 2},
		{11, 6, 3, 3},
		{13, 7, 4, 3},
	}

	for _, tt := range tests {
		accelerated, fullVP, halfVP := BakuAccelerationRounds(tt.totalRounds)
		if accelerated != tt.wantAccelerated {
			t.Errorf("BakuAccelerationRounds(%d): accelerated = %d, want %d",
				tt.totalRounds, accelerated, tt.wantAccelerated)
		}
		if fullVP != tt.wantFullVP {
			t.Errorf("BakuAccelerationRounds(%d): fullVP = %d, want %d",
				tt.totalRounds, fullVP, tt.wantFullVP)
		}
		if halfVP != tt.wantHalfVP {
			t.Errorf("BakuAccelerationRounds(%d): halfVP = %d, want %d",
				tt.totalRounds, halfVP, tt.wantHalfVP)
		}
	}
}

func TestBakuGASize(t *testing.T) {
	tests := []struct {
		totalPlayers int
		wantGA       int
	}{
		{4, 2},
		{8, 4},
		{9, 6},
		{10, 6},
		{12, 6},
		{16, 8},
		{20, 10},
		{161, 82},
	}

	for _, tt := range tests {
		ga := BakuGASize(tt.totalPlayers)
		if ga != tt.wantGA {
			t.Errorf("BakuGASize(%d) = %d, want %d",
				tt.totalPlayers, ga, tt.wantGA)
		}
	}
}

func TestBakuVirtualPoints_9Round(t *testing.T) {
	tests := []struct {
		round int
		isGA  bool
		want  float64
	}{
		{1, true, 1.0},
		{3, true, 1.0},
		{4, true, 0.5},
		{5, true, 0.5},
		{6, true, 0.0},
		{1, false, 0.0},
		{3, false, 0.0},
	}

	for _, tt := range tests {
		vp := BakuVirtualPoints(9, tt.round, tt.isGA)
		if vp != tt.want {
			t.Errorf("BakuVirtualPoints(9, %d, %v) = %.1f, want %.1f",
				tt.round, tt.isGA, vp, tt.want)
		}
	}
}

func TestBakuVirtualPoints_5Round(t *testing.T) {
	tests := []struct {
		round int
		isGA  bool
		want  float64
	}{
		{1, true, 1.0},
		{2, true, 1.0},
		{3, true, 0.5},
		{4, true, 0.0},
	}

	for _, tt := range tests {
		vp := BakuVirtualPoints(5, tt.round, tt.isGA)
		if vp != tt.want {
			t.Errorf("BakuVirtualPoints(5, %d, %v) = %.1f, want %.1f",
				tt.round, tt.isGA, vp, tt.want)
		}
	}
}

func TestApplyBakuAcceleration(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", InitialRank: 1, Score: 0.0, PairingScore: 0.0},
		{ID: "p2", InitialRank: 2, Score: 0.0, PairingScore: 0.0},
		{ID: "p3", InitialRank: 3, Score: 0.0, PairingScore: 0.0},
		{ID: "p4", InitialRank: 4, Score: 0.0, PairingScore: 0.0},
	}

	// 9-round tournament, round 1, GA size 2 (top 2 players).
	ApplyBakuAcceleration(players, 1, 9, 2)

	// p1 and p2 are in GA → PairingScore = 0.0 + 1.0 = 1.0
	if players[0].PairingScore != 1.0 {
		t.Errorf("p1 PairingScore = %.1f, want 1.0", players[0].PairingScore)
	}
	if players[1].PairingScore != 1.0 {
		t.Errorf("p2 PairingScore = %.1f, want 1.0", players[1].PairingScore)
	}

	// p3 and p4 are NOT in GA → PairingScore unchanged.
	if players[2].PairingScore != 0.0 {
		t.Errorf("p3 PairingScore = %.1f, want 0.0", players[2].PairingScore)
	}
	if players[3].PairingScore != 0.0 {
		t.Errorf("p4 PairingScore = %.1f, want 0.0", players[3].PairingScore)
	}
}
