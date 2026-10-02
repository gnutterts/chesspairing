// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"context"
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPair_ByeGoesToEligiblePlayerWhenColorsRestrictTheField(t *testing.T) {
	game := func(w, b string, r chesspairing.GameResult) chesspairing.GameData {
		return chesspairing.GameData{WhiteID: w, BlackID: b, Result: r}
	}
	round := func(number int, bye string, games ...chesspairing.GameData) chesspairing.RoundData {
		return chesspairing.RoundData{
			Number: number,
			Games:  games,
			Byes:   []chesspairing.ByeEntry{{PlayerID: bye, Type: chesspairing.ByePAB}},
		}
	}
	var players []chesspairing.PlayerEntry
	for i := 1; i <= 7; i++ {
		players = append(players, chesspairing.PlayerEntry{ID: fmt.Sprintf("p%d", i), Rating: 2600 - i*10})
	}
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			round(1, "p7",
				game("p1", "p4", chesspairing.ResultBlackWins),
				game("p5", "p2", chesspairing.ResultDraw),
				game("p3", "p6", chesspairing.ResultBlackWins),
			),
			round(2, "p3",
				game("p4", "p7", chesspairing.ResultBlackWins),
				game("p6", "p5", chesspairing.ResultBlackWins),
				game("p2", "p1", chesspairing.ResultWhiteWins),
			),
			round(3, "p1",
				game("p7", "p2", chesspairing.ResultDraw),
				game("p5", "p3", chesspairing.ResultDraw),
				game("p4", "p6", chesspairing.ResultBlackWins),
			),
			round(4, "p4",
				game("p7", "p5", chesspairing.ResultWhiteWins),
				game("p2", "p6", chesspairing.ResultDraw),
				game("p3", "p1", chesspairing.ResultDraw),
			),
		},
		CurrentRound: 5,
	}

	// p1, p3, p4 and p7 already had the bye. A complete pairing exists with the
	// bye for p2: p6-p7, p3-p4, p1-p5.
	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 3 || len(result.Byes) != 1 {
		t.Fatalf("got %d pairings and %d byes, want 3 and 1", len(result.Pairings), len(result.Byes))
	}
	switch bye := result.Byes[0].PlayerID; bye {
	case "p2", "p5", "p6":
	default:
		t.Fatalf("bye = %s, want a player who has not had one (p2, p5 or p6)", bye)
	}
}
