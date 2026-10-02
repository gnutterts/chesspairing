// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

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
				game("p1", "p4", chesspairing.ResultWhiteWins),
				game("p5", "p2", chesspairing.ResultBlackWins),
				game("p3", "p6", chesspairing.ResultWhiteWins),
			),
			round(2, "p6",
				game("p7", "p1", chesspairing.ResultDraw),
				game("p2", "p3", chesspairing.ResultBlackWins),
				game("p4", "p5", chesspairing.ResultBlackWins),
			),
			round(3, "p4",
				game("p3", "p7", chesspairing.ResultWhiteWins),
				game("p1", "p2", chesspairing.ResultDraw),
				game("p6", "p5", chesspairing.ResultBlackWins),
			),
			round(4, "p2",
				game("p1", "p3", chesspairing.ResultWhiteWins),
				game("p5", "p7", chesspairing.ResultBlackWins),
				game("p6", "p4", chesspairing.ResultBlackWins),
			),
		},
		CurrentRound: 5,
	}

	// p2, p4, p6 and p7 already had the bye. A complete pairing exists with the
	// bye for p1: p6-p7, p3-p5, p2-p4 (p6 and p7 can only meet with p7 white).
	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	if len(result.Pairings) != 3 || len(result.Byes) != 1 {
		t.Fatalf("got %d pairings and %d byes, want 3 and 1", len(result.Pairings), len(result.Byes))
	}
	switch bye := result.Byes[0].PlayerID; bye {
	case "p1", "p3", "p5":
	default:
		t.Fatalf("bye = %s, want a player who has not had one (p1, p3 or p5)", bye)
	}
}
