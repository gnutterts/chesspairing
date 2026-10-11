package lim

import (
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestExchangeMatch_Art42FiveColumns(t *testing.T) {
	want := [][][2]string{
		{{"p1", "p4"}, {"p2", "p5"}, {"p3", "p6"}},
		{{"p1", "p5"}, {"p2", "p4"}, {"p3", "p6"}},
		{{"p1", "p6"}, {"p2", "p4"}, {"p3", "p5"}},
		{{"p1", "p3"}, {"p2", "p5"}, {"p4", "p6"}},
		{{"p1", "p2"}, {"p3", "p5"}, {"p4", "p6"}},
	}
	for column, pairsWant := range want {
		t.Run(fmt.Sprintf("column_%d", column+1), func(t *testing.T) {
			players := articlePlayers()
			for _, opponent := range []int{4, 5, 6, 3}[:column] {
				markPlayed(players, 1, opponent)
			}
			pairs, unpaired := ExchangeMatch(players, true, nil)
			if len(unpaired) != 0 {
				t.Fatalf("got %d floaters, want none", len(unpaired))
			}
			assertPairs(t, pairs, pairsWant)
			assertExchangeCoverage(t, players, pairs, unpaired)
		})
	}
}

func TestExchangeMatch_Art43FiveColumns(t *testing.T) {
	want := [][][2]string{
		{{"p1", "p6"}, {"p2", "p4"}, {"p3", "p5"}},
		{{"p1", "p6"}, {"p2", "p5"}, {"p3", "p4"}},
		{{"p1", "p6"}, {"p2", "p3"}, {"p4", "p5"}},
		{{"p1", "p3"}, {"p2", "p6"}, {"p4", "p5"}},
		{{"p1", "p2"}, {"p3", "p5"}, {"p4", "p6"}},
	}
	for column, pairsWant := range want {
		t.Run(fmt.Sprintf("column_%d", column+1), func(t *testing.T) {
			players := articlePlayers()
			markPlayed(players, 1, 4)
			markPlayed(players, 1, 5)
			for _, opponent := range []int{4, 5, 3, 6}[:column] {
				markPlayed(players, 2, opponent)
			}
			pairs, unpaired := ExchangeMatch(players, true, nil)
			if len(unpaired) != 0 {
				t.Fatalf("got %d floaters, want none", len(unpaired))
			}
			assertPairs(t, pairs, pairsWant)
			assertExchangeCoverage(t, players, pairs, unpaired)
		})
	}
}

func articlePlayers() []*swisslib.PlayerState {
	players := make([]*swisslib.PlayerState, 6)
	for i := range players {
		players[i] = makePlayer(fmt.Sprintf("p%d", i+1), i+1)
	}
	return players
}

func markPlayed(players []*swisslib.PlayerState, a, b int) {
	players[a-1].Opponents = append(players[a-1].Opponents, players[b-1].ID)
	players[b-1].Opponents = append(players[b-1].Opponents, players[a-1].ID)
}
