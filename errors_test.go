// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestPairingError(t *testing.T) {
	cases := []struct {
		err  *chesspairing.PairingError
		want string
	}{
		{err: &chesspairing.PairingError{Kind: chesspairing.PairingIncomplete, System: "dutch", Missing: []string{"p2", "p1"}}, want: "dutch: pairing is incomplete: missing p1, p2"},
		{err: &chesspairing.PairingError{Kind: chesspairing.PairingImpossible, System: "dubov"}, want: "dubov: no pairing satisfies the absolute criteria"},
		{err: &chesspairing.PairingError{Kind: chesspairing.PairingInvalidInput, System: "lim"}, want: "lim: invalid pairing input"},
	}
	for _, tc := range cases {
		if got := tc.err.Error(); got != tc.want {
			t.Errorf("Error() = %q, want %q", got, tc.want)
		}
	}

	sentinel := &chesspairing.PairingError{
		Kind:   chesspairing.PairingNoPABCandidate,
		System: "dutch",
		Err:    swisslib.ErrNoPABCandidate,
	}
	if got, want := sentinel.Error(), "dutch: no player is eligible for the pairing-allocated bye: no player eligible for the pairing-allocated bye (C2)"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(sentinel, swisslib.ErrNoPABCandidate) {
		t.Error("PairingError does not unwrap the PAB sentinel")
	}
	wrapped := fmt.Errorf("pairing failed: %w", sentinel)
	var got *chesspairing.PairingError
	if !errors.As(wrapped, &got) || got != sentinel {
		t.Errorf("errors.As() = %v, want original pairing error", got)
	}
}

func TestValidatePairing(t *testing.T) {
	state := pairingState()
	cases := []struct {
		name    string
		result  *chesspairing.PairingResult
		kind    chesspairing.PairingErrorKind
		missing []string
	}{
		{name: "complete", result: &chesspairing.PairingResult{Pairings: []chesspairing.GamePairing{{WhiteID: "p1", BlackID: "p2"}, {WhiteID: "p3", BlackID: "p4"}}}},
		{name: "one missing", result: &chesspairing.PairingResult{Pairings: []chesspairing.GamePairing{{WhiteID: "p1", BlackID: "p2"}}, Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}}}, kind: chesspairing.PairingIncomplete, missing: []string{"p4"}},
		{name: "several missing", result: &chesspairing.PairingResult{}, kind: chesspairing.PairingIncomplete, missing: []string{"p1", "p2", "p3", "p4"}},
		{name: "duplicate", result: &chesspairing.PairingResult{Pairings: []chesspairing.GamePairing{{WhiteID: "p1", BlackID: "p2"}, {WhiteID: "p1", BlackID: "p3"}}}, kind: chesspairing.PairingInvalidInput},
		{name: "two PABs", result: &chesspairing.PairingResult{Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByePAB}, {PlayerID: "p2", Type: chesspairing.ByePAB}}}, kind: chesspairing.PairingInvalidInput},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := chesspairing.ValidatePairing(state, tc.result)
			if tc.kind == 0 {
				if err != nil {
					t.Fatalf("ValidatePairing() error = %v", err)
				}
				return
			}
			var pairingErr *chesspairing.PairingError
			if !errors.As(err, &pairingErr) || pairingErr.Kind != tc.kind || !reflect.DeepEqual(pairingErr.Missing, tc.missing) {
				t.Fatalf("ValidatePairing() = %#v, want kind %v, missing %v", err, tc.kind, tc.missing)
			}
		})
	}

	repeat := pairingState()
	repeat.Rounds = []chesspairing.RoundData{{Number: 1, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2"}}}}
	if err := chesspairing.ValidatePairing(repeat, &chesspairing.PairingResult{Pairings: []chesspairing.GamePairing{{WhiteID: "p1", BlackID: "p2"}, {WhiteID: "p3", BlackID: "p4"}}}); err == nil {
		t.Fatal("repeated Dutch pairing passed validation")
	}

	keizer := repeat
	keizer.PairingConfig = chesspairing.PairingConfig{System: chesspairing.PairingKeizer, Options: map[string]any{"allowRepeatPairings": true}}
	if err := chesspairing.ValidatePairing(keizer, &chesspairing.PairingResult{Pairings: []chesspairing.GamePairing{{WhiteID: "p1", BlackID: "p2"}, {WhiteID: "p3", BlackID: "p4"}}}); err != nil {
		t.Fatalf("Keizer repeated pairing error = %v", err)
	}
}

func pairingState() *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players:       []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}, {ID: "p3"}, {ID: "p4"}},
		CurrentRound:  2,
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDutch},
	}
}
