// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPreAssignedPAB_C2Validation(t *testing.T) {
	pairer := New(Options{})

	// 1. Second PAB rejected with an error
	stateSecondPAB := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "B", BlackID: "C", Result: chesspairing.ResultDraw},
				{WhiteID: "C", BlackID: "B", Result: chesspairing.ResultDraw},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "A", Type: chesspairing.ByePAB}},
		}},
		CurrentRound: 2,
		PreAssignedByes: []chesspairing.ByeEntry{
			{PlayerID: "A", Type: chesspairing.ByePAB},
		},
	}
	_, err := pairer.Pair(context.Background(), stateSecondPAB)
	if err == nil {
		t.Fatal("expected error for second PAB on player A in doubleswiss, got nil")
	}
	if !strings.Contains(err.Error(), "A") || !strings.Contains(err.Error(), "C2") {
		t.Errorf("error should name player A and C2, got %v", err)
	}

	// 2. Player with forfeit win rejected
	stateForfeitWin := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultForfeitWhiteWins},
				{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultForfeitBlackWins},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "C", Type: chesspairing.ByeHalf}},
		}},
		CurrentRound: 2,
		PreAssignedByes: []chesspairing.ByeEntry{
			{PlayerID: "A", Type: chesspairing.ByePAB},
		},
	}
	_, err = pairer.Pair(context.Background(), stateForfeitWin)
	if err == nil {
		t.Fatal("expected error for forfeit winner A receiving pre-assigned PAB in doubleswiss, got nil")
	}
	if !strings.Contains(err.Error(), "A") || !strings.Contains(err.Error(), "C2") {
		t.Errorf("error should name player A and C2, got %v", err)
	}
}
