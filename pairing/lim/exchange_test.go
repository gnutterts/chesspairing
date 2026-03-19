package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func makePlayer(id string, tpn int) *swisslib.PlayerState {
	return &swisslib.PlayerState{ID: id, TPN: tpn, Active: true}
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
		{ID: "p1", TPN: 1, Opponents: []string{"p4"}, Active: true},
		makePlayer("p2", 2), makePlayer("p3", 3),
		{ID: "p4", TPN: 4, Opponents: []string{"p1"}, Active: true},
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
		{ID: "p1", TPN: 1, Opponents: []string{"p2", "p3", "p4"}, Active: true},
		{ID: "p2", TPN: 2, Opponents: []string{"p1"}, Active: true},
		{ID: "p3", TPN: 3, Opponents: []string{"p1"}, Active: true},
		{ID: "p4", TPN: 4, Opponents: []string{"p1"}, Active: true},
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
