// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestPABScoresWinAndDraw(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "a", BlackID: "d", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "d", BlackID: "a", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "b", BlackID: "c", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "c", BlackID: "b", Result: chesspairing.ResultWhiteWins},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "e", Type: chesspairing.ByePAB}},
		}},
		CurrentRound: 2,
	}
	participants, err := lexswiss.BuildParticipantStatesWithPAB(state, 1.5)
	if err != nil {
		t.Fatal(err)
	}
	for _, participant := range participants {
		if participant.ID == "e" && participant.Score != 1.5 {
			t.Fatalf("PAB score = %v, want 1.5", participant.Score)
		}
	}
	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	// C.04.5 3.5.3-3.5.5: a would leave b and c, who have met, so b floats up.
	if !hasPair(result.Pairings, "e", "b") || !hasPair(result.Pairings, "a", "c") {
		t.Errorf("want FIDE pairing E-B and A-C (C.04.5 3.5.3-3.5.5), got %v", result.Pairings)
	}
}

func hasPair(pairings []chesspairing.GamePairing, first, second string) bool {
	for _, pairing := range pairings {
		if (pairing.WhiteID == first && pairing.BlackID == second) || (pairing.WhiteID == second && pairing.BlackID == first) {
			return true
		}
	}
	return false
}
