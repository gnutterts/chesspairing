// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "testing"

// AddVirtualPoints makes Score and PairingScore the pairing score and orders
// the players by it, so that group A players rank above group B players with
// the same real score.
func TestAddVirtualPointsOrdersByPairingScore(t *testing.T) {
	players := []PlayerState{
		{ID: "b1", InitialRank: 1, Score: 0.5, PairingScore: 0.5},
		{ID: "a2", InitialRank: 2, Score: 0, PairingScore: 0},
		{ID: "a3", InitialRank: 3, Score: 0, PairingScore: 0},
	}
	AddVirtualPoints(players, func(id string) float64 {
		if id == "b1" {
			return 0
		}
		return 1
	})
	want := []string{"a2", "a3", "b1"}
	for i, id := range want {
		if players[i].ID != id || players[i].TPN != i+1 {
			t.Fatalf("position %d = %s (TPN %d), want %s (TPN %d)", i, players[i].ID, players[i].TPN, id, i+1)
		}
	}
	if players[0].Score != 1 || players[0].PairingScore != 1 {
		t.Errorf("score = %v / %v, want 1 / 1", players[0].Score, players[0].PairingScore)
	}
	if players[2].Score != 0.5 {
		t.Errorf("group B score = %v, want unchanged 0.5", players[2].Score)
	}
}
