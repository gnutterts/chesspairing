// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"context"
	"errors"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestD1B_UnavoidableRematch(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1", Rating: 2400}, {ID: "p2", Rating: 2300}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw}}},
		},
		CurrentRound: 2,
	}
	result, err := New(Options{}).Pair(context.Background(), state)
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	var pairingErr *chesspairing.PairingError
	if !errors.As(err, &pairingErr) || pairingErr.Kind != chesspairing.PairingIncomplete || len(pairingErr.Missing) != 2 || pairingErr.Partial == nil {
		t.Fatalf("error = %#v, want incomplete error with both players missing and a partial result", err)
	}
}
