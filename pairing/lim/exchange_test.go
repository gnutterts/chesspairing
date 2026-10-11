// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func makePlayer(id string, tpn int) *swisslib.PlayerState {
	return &swisslib.PlayerState{ID: id, TPN: tpn}
}

func TestExchangeMatch_BasicSixPlayers(t *testing.T) {
	// Art. 4.2 example: 6 players, proposed pairings 1v4, 2v5, 3v6.
	// All compatible → should produce 1v4, 2v5, 3v6.
	players := []*swisslib.PlayerState{
		makePlayer("p1", 1), makePlayer("p2", 2), makePlayer("p3", 3),
		makePlayer("p4", 4), makePlayer("p5", 5), makePlayer("p6", 6),
	}

	pairs, unpaired := ExchangeMatch(players, true, nil)
	if len(unpaired) != 0 {
		t.Errorf("expected no unpaired, got %d", len(unpaired))
	}
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}

	// Check pairings: 1v4, 2v5, 3v6 (when pairing downward, scrutiny
	// starts from the top — highest numbered in top half = player 3).
	expected := [][2]string{{"p1", "p4"}, {"p2", "p5"}, {"p3", "p6"}}
	for i, pair := range pairs {
		ids := [2]string{pair[0].ID, pair[1].ID}
		if ids != expected[i] {
			t.Errorf("pair %d: expected %v, got %v", i, expected[i], ids)
		}
	}
}

func TestExchangeMatch_ExchangeNeeded(t *testing.T) {
	// 6 players. Player 1 already played player 4 → exchange needed.
	players := []*swisslib.PlayerState{
		{ID: "p1", TPN: 1, Opponents: []string{"p4"}},
		makePlayer("p2", 2), makePlayer("p3", 3),
		{ID: "p4", TPN: 4, Opponents: []string{"p1"}},
		makePlayer("p5", 5), makePlayer("p6", 6),
	}

	pairs, unpaired := ExchangeMatch(players, true, nil)
	if len(unpaired) != 0 {
		t.Errorf("expected no unpaired, got %d", len(unpaired))
	}
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}

	// Player 1 should be paired with someone other than player 4.
	for _, pair := range pairs {
		if (pair[0].ID == "p1" && pair[1].ID == "p4") || (pair[0].ID == "p4" && pair[1].ID == "p1") {
			t.Error("player 1 should NOT be paired with player 4")
		}
	}
}

func TestExchangeMatch_UnpairedFloater(t *testing.T) {
	// 4 players, but p1 has played everyone else → can't pair p1.
	players := []*swisslib.PlayerState{
		{ID: "p1", TPN: 1, Opponents: []string{"p2", "p3", "p4"}},
		{ID: "p2", TPN: 2, Opponents: []string{"p1"}},
		{ID: "p3", TPN: 3, Opponents: []string{"p1"}},
		{ID: "p4", TPN: 4, Opponents: []string{"p1"}},
	}

	pairs, unpaired := ExchangeMatch(players, true, nil)
	// p1 must float (unpaired), and one other must also float to keep even.
	if len(unpaired) < 1 {
		t.Error("expected at least 1 unpaired player")
	}
	// Check that p1 is among unpaired.
	found := false
	for _, u := range unpaired {
		if u.ID == "p1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected p1 to be unpaired")
	}
	// Remaining should form valid pairs.
	for _, pair := range pairs {
		if swisslib.HasPlayed(pair[0], pair[1]) {
			t.Errorf("invalid pair: %s vs %s already played", pair[0].ID, pair[1].ID)
		}
	}
}

func TestExchangeMatch_TwoPlayers(t *testing.T) {
	players := []*swisslib.PlayerState{
		makePlayer("p1", 1), makePlayer("p2", 2),
	}

	pairs, unpaired := ExchangeMatch(players, true, nil)
	if len(pairs) != 1 || len(unpaired) != 0 {
		t.Errorf("expected 1 pair, 0 unpaired; got %d pairs, %d unpaired", len(pairs), len(unpaired))
	}
}

func TestExchangeMatch_OddPlayers(t *testing.T) {
	// Odd number of players — one should be returned as unpaired.
	players := []*swisslib.PlayerState{
		makePlayer("p1", 1), makePlayer("p2", 2), makePlayer("p3", 3),
	}

	// The caller should have already removed the bye player before calling
	// ExchangeMatch. But if odd, we return the last player as unpaired.
	pairs, unpaired := ExchangeMatch(players, true, nil)
	if len(pairs) != 1 {
		t.Errorf("expected 1 pair, got %d", len(pairs))
	}
	if len(unpaired) != 1 {
		t.Errorf("expected 1 unpaired, got %d", len(unpaired))
	}
}

func TestExchangeMatch_CrossHalfExchange(t *testing.T) {
	// Art. 4.2 permits an S1-S1 exchange after trying S2.
	players := []*swisslib.PlayerState{
		{ID: "p1", TPN: 1},
		{ID: "p2", TPN: 2},
		{ID: "p3", TPN: 3, Opponents: []string{"p4", "p5", "p6"}},
		{ID: "p4", TPN: 4, Opponents: []string{"p3"}},
		{ID: "p5", TPN: 5, Opponents: []string{"p3"}},
		{ID: "p6", TPN: 6, Opponents: []string{"p3"}},
	}

	pairs, unpaired := ExchangeMatch(players, true, nil)
	if len(unpaired) != 0 {
		t.Errorf("expected no unpaired players, got %d", len(unpaired))
	}
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}
	assertPairs(t, pairs, [][2]string{{"p1", "p4"}, {"p2", "p3"}, {"p5", "p6"}})
	assertExchangeCoverage(t, players, pairs, unpaired)
}

func TestExchangeMatch_CrossHalfExchangeS1S1(t *testing.T) {
	// Art. 4.2 column 5 permits the S1-S1 pairing 1 v 2.
	players := []*swisslib.PlayerState{
		{ID: "p1", TPN: 1, Opponents: []string{"p3", "p4"}},
		{ID: "p2", TPN: 2, Opponents: []string{"p3"}},
		{ID: "p3", TPN: 3, Opponents: []string{"p1", "p2"}},
		{ID: "p4", TPN: 4, Opponents: []string{"p1"}},
	}

	pairs, unpaired := ExchangeMatch(players, true, nil)
	if len(unpaired) != 0 {
		t.Errorf("expected no unpaired players, got %d", len(unpaired))
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	assertPairs(t, pairs, [][2]string{{"p1", "p2"}, {"p3", "p4"}})
	assertExchangeCoverage(t, players, pairs, unpaired)
}

func assertPairs(t *testing.T, pairs [][2]*swisslib.PlayerState, want [][2]string) {
	t.Helper()
	if len(pairs) != len(want) {
		t.Fatalf("got %d pairs, want %d", len(pairs), len(want))
	}
	for i, pair := range pairs {
		got := [2]string{pair[0].ID, pair[1].ID}
		if got != want[i] {
			t.Errorf("pair %d: got %v, want %v", i, got, want[i])
		}
	}
}
