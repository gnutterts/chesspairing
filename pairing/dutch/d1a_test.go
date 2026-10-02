// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"errors"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/burstein"
)

func TestD1A_AllOpponentsAlreadyMet(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1", Rating: 2400}, {ID: "p2", Rating: 2300}, {ID: "p3", Rating: 2200}, {ID: "p4", Rating: 2100}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2"}, {WhiteID: "p3", BlackID: "p4"}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p3"}, {WhiteID: "p2", BlackID: "p4"}}},
			{Number: 3, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p4"}, {WhiteID: "p2", BlackID: "p3"}}},
		},
		CurrentRound: 4,
	}
	pairers := []struct {
		name   string
		pairer chesspairing.Pairer
	}{
		{name: "dutch", pairer: New(Options{})},
		{name: "burstein", pairer: burstein.New(burstein.Options{})},
	}
	for _, tc := range pairers {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.pairer.Pair(context.Background(), state)
			if result != nil {
				t.Fatalf("result = %#v, want nil", result)
			}
			var pairingErr *chesspairing.PairingError
			if !errors.As(err, &pairingErr) || pairingErr.Kind != chesspairing.PairingImpossible || len(pairingErr.Missing) != 0 || pairingErr.Partial != nil {
				t.Fatalf("error = %#v, want impossible error without missing players or partial result", err)
			}
		})
	}
}
