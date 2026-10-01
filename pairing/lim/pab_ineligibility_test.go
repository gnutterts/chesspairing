// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"context"
	"errors"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestNoPABCandidate(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:      []chesspairing.PlayerEntry{{ID: "p1"}},
		Rounds:       []chesspairing.RoundData{{Number: 1, Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByePAB}}}},
		CurrentRound: 2,
	}
	_, err := New(Options{}).Pair(context.Background(), state)
	if !errors.Is(err, swisslib.ErrNoPABCandidate) {
		t.Fatalf("Pair() error = %v, want ErrNoPABCandidate", err)
	}
}

func TestPABIneligibilityForfeitWinner(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}, {ID: "p3"}, {ID: "p4"}, {ID: "p5"}},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
					{WhiteID: "p5", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
					{WhiteID: "p3", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeZero}},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
					{WhiteID: "p1", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByeZero}},
			},
		},
		CurrentRound: 4,
	}
	_, err := New(Options{}).Pair(context.Background(), state)
	if !errors.Is(err, swisslib.ErrNoPABCandidate) {
		t.Fatalf("Pair() error = %v, want ErrNoPABCandidate", err)
	}
}
