// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPABLeavesLegalPairing(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "a", BlackID: "c", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "c", BlackID: "a", Result: chesspairing.ResultBlackWins},
				{WhiteID: "a", BlackID: "d", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "d", BlackID: "a", Result: chesspairing.ResultBlackWins},
				{WhiteID: "c", BlackID: "d", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "d", BlackID: "c", Result: chesspairing.ResultBlackWins},
			},
		}},
		CurrentRound: 2,
	}

	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "d" {
		t.Errorf("PAB = %v, want d: removing lower-ranked e leaves no C1 pairing (Art. 3.4.1)", result.Byes)
	}
}
