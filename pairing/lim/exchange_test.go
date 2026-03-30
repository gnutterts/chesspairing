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

func TestExchangeMatch_CrossHalfExchange(t *testing.T) {
	// 6 players. S1={p1,p2,p3}, S2={p4,p5,p6}.
	// p3 has played all S2 players → no S2 partner available.
	// Without cross-half exchange, tryExchangePairing fails at p3 and falls
	// through to greedyPair. With cross-half exchange, p3 takes p2 as a
	// cross-half partner (S1-S1), p1 pairs with p4 (S1-S2), and the
	// displaced S2 leftover p5-p6 pair among themselves.
	players := []*swisslib.PlayerState{
		{ID: "p1", TPN: 1, Active: true},
		{ID: "p2", TPN: 2, Active: true},
		{ID: "p3", TPN: 3, Opponents: []string{"p4", "p5", "p6"}, Active: true},
		{ID: "p4", TPN: 4, Opponents: []string{"p3"}, Active: true},
		{ID: "p5", TPN: 5, Opponents: []string{"p3"}, Active: true},
		{ID: "p6", TPN: 6, Opponents: []string{"p3"}, Active: true},
	}

	pairs, unpaired := ExchangeMatch(players, true, nil)
	if len(unpaired) != 0 {
		t.Errorf("expected no unpaired, got %d", len(unpaired))
	}
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}

	// Verify all pairs are compatible (no prior opponents).
	for _, pair := range pairs {
		if swisslib.HasPlayed(pair[0], pair[1]) {
			t.Errorf("invalid pair: %s vs %s already played", pair[0].ID, pair[1].ID)
		}
	}

	// Verify the cross-half pair: p3 paired with p2 (both S1).
	foundCrossHalf := false
	for _, pair := range pairs {
		ids := [2]string{pair[0].ID, pair[1].ID}
		if (ids[0] == "p3" && ids[1] == "p2") || (ids[0] == "p2" && ids[1] == "p3") {
			foundCrossHalf = true
		}
	}
	if !foundCrossHalf {
		t.Error("expected cross-half pair p3-p2")
	}

	// Verify the S2-S2 leftover pair: p5 paired with p6.
	foundLeftover := false
	for _, pair := range pairs {
		ids := [2]string{pair[0].ID, pair[1].ID}
		if (ids[0] == "p5" && ids[1] == "p6") || (ids[0] == "p6" && ids[1] == "p5") {
			foundLeftover = true
		}
	}
	if !foundLeftover {
		t.Error("expected S2-S2 leftover pair p5-p6")
	}
}

func TestExchangeMatch_CrossHalfExchangeS1S1(t *testing.T) {
	// 4 players. S1={p1,p2}, S2={p3,p4}.
	// p1 has played p3 and p4 (can't pair with any S2).
	// p2 has played p3 (can pair with p4 only).
	// Cross-half: p1 pairs with p2 (S1-S1), leaving p3 and p4 (S2-S2).
	// But p3 and p4 must be compatible for a complete pairing.
	players := []*swisslib.PlayerState{
		{ID: "p1", TPN: 1, Opponents: []string{"p3", "p4"}, Active: true},
		{ID: "p2", TPN: 2, Opponents: []string{"p3"}, Active: true},
		{ID: "p3", TPN: 3, Opponents: []string{"p1", "p2"}, Active: true},
		{ID: "p4", TPN: 4, Opponents: []string{"p1"}, Active: true},
	}

	pairs, unpaired := ExchangeMatch(players, true, nil)
	if len(unpaired) != 0 {
		t.Errorf("expected no unpaired, got %d", len(unpaired))
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// Verify all pairs are compatible.
	for _, pair := range pairs {
		if swisslib.HasPlayed(pair[0], pair[1]) {
			t.Errorf("invalid pair: %s vs %s already played", pair[0].ID, pair[1].ID)
		}
	}
}
