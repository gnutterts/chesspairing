// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"context"
	"errors"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestD1D_EvenFieldCannotReceiveTwoPairingAllocatedByes(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1", Rating: 2400}, {ID: "p2", Rating: 2300}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw}}},
		},
		CurrentRound: 2,
	}
	assertD1DImpossible(t, state, Options{})
}

func TestD1D_PlayerCannotReceiveSecondPairingAllocatedBye_ReturnsImpossible(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1", Rating: 2400}, {ID: "p2", Rating: 2300}, {ID: "p3", Rating: 2200}},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games:  []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins}},
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
		},
		CurrentRound: 2,
	}
	result, err := New(Options{ForbiddenPairs: [][]string{{"p1", "p3"}}}).Pair(context.Background(), state)
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	var pairingErr *chesspairing.PairingError
	// p3 already received the only PAB, p1 cannot play p2 again or p3, so
	// no complete pairing exists; PairingImpossible is the correct result.
	if !errors.As(err, &pairingErr) || pairingErr.Kind != chesspairing.PairingImpossible || pairingErr.Partial != nil {
		t.Fatalf("error = %#v, want impossible error without partial result", err)
	}
}

func assertD1DImpossible(t *testing.T, state *chesspairing.TournamentState, options Options) {
	t.Helper()
	result, err := New(options).Pair(context.Background(), state)
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	var pairingErr *chesspairing.PairingError
	if !errors.As(err, &pairingErr) || pairingErr.Kind != chesspairing.PairingImpossible || len(pairingErr.Missing) != 0 || pairingErr.Partial != nil {
		t.Fatalf("error = %#v, want impossible error without missing players or partial result", err)
	}
}
