// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// TestFIDEExample_dubov_1 verifies the worked example in FIDE C.04.4.1
// (Dubov System, effective from 1 February 2026), Article 4.4.2 Note.
//
// Source text: "If, for instance, players A, B, C (listed according to their
// ascending TPN) are in G2, the different Transpositions are {A, B, C}
// {A, C, B} {B, A, C} {B, C, A} {C, A, B} and {C, B, A}, in that exact order."
//
// The example is rebuilt with the package-internal helper
// generateDubovTranspositions, which enumerates the G2 transpositions in the
// order prescribed by Article 4.4.
func TestFIDEExample_dubov_1(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "A", TPN: 1},
		{ID: "B", TPN: 2},
		{ID: "C", TPN: 3},
	}

	// Transcription guard: the FIDE note lists A, B, C in ascending TPN order.
	if len(players) != 3 {
		t.Fatalf("input must contain exactly 3 players, got %d", len(players))
	}
	for i, p := range players {
		if p.TPN != i+1 {
			t.Fatalf("input players must be in ascending TPN order, got TPN %d at position %d", p.TPN, i)
		}
	}

	perms := generateDubovTranspositions(players, 120)

	var gotParts []string
	for _, perm := range perms {
		ids := make([]string, len(perm))
		for i, p := range perm {
			ids[i] = p.ID
		}
		gotParts = append(gotParts, "{"+strings.Join(ids, ", ")+"}")
	}
	got := strings.Join(gotParts, " ")
	fide := "{A, B, C} {A, C, B} {B, A, C} {B, C, A} {C, A, B} {C, B, A}"
	if got != fide {
		t.Errorf("dubov-4.4.2: transpositions = %s, want FIDE %s", got, fide)
	}
}
