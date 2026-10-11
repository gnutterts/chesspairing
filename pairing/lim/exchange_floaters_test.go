package lim

import (
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestExchangeMatch_InitialFloaterUsesArticle310Selection(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "incompatible", TPN: 1},
		{ID: "a", TPN: 2},
		{ID: "replacement", TPN: 3},
		{ID: "previous", TPN: 4},
	}
	_, unpaired := exchangeMatchWithFloaterSelector(players, true, func(a, b *swisslib.PlayerState) bool {
		return a.ID != "incompatible" && b.ID != "incompatible"
	}, func(pairable []*swisslib.PlayerState) *swisslib.PlayerState {
		return selectFloater(pairable, []*swisslib.PlayerState{{ID: "target"}}, nil, nil, false, true, map[string]bool{"previous": true}, true, nil)
	})
	foundReplacement, foundPrevious := false, false
	for _, player := range unpaired {
		foundReplacement = foundReplacement || player.ID == "replacement"
		foundPrevious = foundPrevious || player.ID == "previous"
	}
	if !foundReplacement || foundPrevious {
		t.Fatalf("Art. 2.3.4 should protect previous with type-D replacement, unpaired = %#v", unpaired)
	}
}

func TestExchangeMatch_Art442FloatsLowestNumberedCompanion(t *testing.T) {
	players := articlePlayers()
	markPlayed(players, 1, 4)
	markPlayed(players, 1, 5)
	markPlayed(players, 1, 3)
	for _, opponent := range []int{1, 3, 4, 5} {
		markPlayed(players, 2, opponent)
	}

	pairs, floaters := ExchangeMatch(players, true, nil)
	assertPairs(t, pairs, [][2]string{{"p1", "p6"}, {"p3", "p4"}})
	assertFloaterIDs(t, floaters, "p2", "p5")
	assertExchangeCoverage(t, players, pairs, floaters)
}

func TestExchangeMatch_Art441UsesInitialFloaterOnce(t *testing.T) {
	players := articlePlayers()
	for _, opponent := range []int{1, 2, 3, 4, 5} {
		markPlayed(players, 6, opponent)
	}
	for _, opponent := range []int{2, 3, 4} {
		markPlayed(players, 1, opponent)
	}

	pairs, floaters := ExchangeMatch(players, true, nil)
	assertPairs(t, pairs, [][2]string{{"p2", "p4"}, {"p3", "p5"}})
	assertFloaterIDs(t, floaters, "p6", "p1")
	assertExchangeCoverage(t, players, pairs, floaters)
}

func TestExchangeMatch_Art441ExchangesInitialFloaterWithPlayerTwo(t *testing.T) {
	players := make([]*swisslib.PlayerState, 7)
	for i := range players {
		players[i] = &swisslib.PlayerState{ID: fmt.Sprintf("p%d", i+1), TPN: i + 1}
	}
	for opponent := 2; opponent <= 5; opponent++ {
		markPlayed(players, 1, opponent)
	}
	for opponent := 3; opponent <= 6; opponent++ {
		markPlayed(players, 2, opponent)
	}

	pairs, floaters := ExchangeMatch(players, true, nil)
	if len(pairs) != 3 {
		t.Fatalf("Art. 4.4.1 pairs = %#v, want three pairs", pairs)
	}
	assertFloaterIDs(t, floaters, "p2")
	assertExchangeCoverage(t, players, pairs, floaters)
}

func TestExchangeMatch_Art442MayFloatProposedPartner(t *testing.T) {
	players := articlePlayers()
	compatible := func(a, b *swisslib.PlayerState) bool {
		pair := [2]string{a.ID, b.ID}
		if pair[0] > pair[1] {
			pair[0], pair[1] = pair[1], pair[0]
		}
		return map[[2]string]bool{
			{"p1", "p4"}: true,
			{"p2", "p5"}: true,
			{"p3", "p4"}: true,
			{"p4", "p6"}: true,
		}[pair]
	}

	pairs, floaters := ExchangeMatchWithCompatibility(players, true, nil, compatible)
	assertPairs(t, pairs, [][2]string{{"p1", "p4"}, {"p2", "p5"}})
	assertFloaterIDs(t, floaters, "p3", "p6")
	assertExchangeCoverage(t, players, pairs, floaters)
}

func assertExchangeCoverage(t *testing.T, players []*swisslib.PlayerState, pairs [][2]*swisslib.PlayerState, floaters []*swisslib.PlayerState) {
	t.Helper()
	seen := make(map[string]bool, len(players))
	for _, pair := range pairs {
		for _, player := range pair {
			if seen[player.ID] {
				t.Errorf("player %s appears more than once", player.ID)
			}
			seen[player.ID] = true
		}
	}
	for _, player := range floaters {
		if seen[player.ID] {
			t.Errorf("player %s appears more than once", player.ID)
		}
		seen[player.ID] = true
	}
	for _, player := range players {
		if !seen[player.ID] {
			t.Errorf("player %s is neither paired nor floated", player.ID)
		}
	}
}

func assertFloaterIDs(t *testing.T, floaters []*swisslib.PlayerState, want ...string) {
	t.Helper()
	if len(floaters) != len(want) {
		t.Fatalf("got %d floaters, want %d", len(floaters), len(want))
	}
	for i, player := range floaters {
		if player.ID != want[i] {
			t.Errorf("floater %d: got %s, want %s", i, player.ID, want[i])
		}
	}
}
