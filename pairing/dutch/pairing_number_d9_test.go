package dutch

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// C.04.2 art. 2.2.2 requires ranking equal-strength players by
// "FIDE-title (GM-IM-WGM-FM-WIM-CM-WFM-WCM-no title)" before art. 2.2.3
// "Alphabetically".  With equal ratings the required TPN order is
// Zara(GM), Yuri(IM), Anna(no title), Bob(no title), producing Zara-Anna
// and Bob-Yuri in Dutch round one.
func TestPairingNumber_D9_TitlesDetermineTPNAndDutchPairing(t *testing.T) {
	state := &chesspairing.TournamentState{Players: []chesspairing.PlayerEntry{
		{ID: "anna", DisplayName: "Anna", Rating: 2000},
		{ID: "zara", DisplayName: "Zara", Rating: 2000, Title: "GM"},
		{ID: "bob", DisplayName: "Bob", Rating: 2000},
		{ID: "yuri", DisplayName: "Yuri", Rating: 2000, Title: "IM"},
	}, CurrentRound: 1}
	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	opponent := map[string]string{}
	for _, game := range result.Pairings {
		opponent[game.WhiteID] = game.BlackID
		opponent[game.BlackID] = game.WhiteID
	}
	if opponent["anna"] != "zara" || opponent["bob"] != "yuri" {
		t.Fatalf("expected title-ranked pairs Zara-Anna and Bob-Yuri, got %+v", result.Pairings)
	}
}
