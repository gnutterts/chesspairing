// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func FuzzPair(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(1001))
	f.Fuzz(func(t *testing.T, seed uint64) {
		state := fuzzState(seed)
		result, err := New(Options{}).Pair(context.Background(), state)
		if err != nil {
			var pairingErr *chesspairing.PairingError
			if !errors.As(err, &pairingErr) {
				t.Fatalf("Pair() error = %v, want *PairingError", err)
			}
			return
		}
		if err := chesspairing.ValidatePairing(state, result); err != nil {
			t.Fatalf("ValidatePairing() error = %v", err)
		}
	})
}

func fuzzState(seed uint64) *chesspairing.TournamentState {
	players := make([]chesspairing.PlayerEntry, int(seed%8)+4)
	for i := range players {
		seed = seed*6364136223846793005 + 1
		players[i] = chesspairing.PlayerEntry{ID: fmt.Sprintf("p%d", i+1), Rating: 1200 + int(seed%1600), PairingNumber: i + 1}
	}
	return &chesspairing.TournamentState{Players: players, CurrentRound: 1, PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingKeizer}}
}
