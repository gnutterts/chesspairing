// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestFIDEVoorbeeld_baku_ex1 verifies FIDE C.04.7 art. 1.4.4 ex. 1: a
// nine-round individual tournament with standard scoring (win = 1) assigns GA
// players one virtual point in the first three rounds and half a virtual point
// in the next two rounds.
func TestFIDEVoorbeeld_baku_ex1(t *testing.T) {
	acceleration := "baku"
	totalRounds := 9
	p := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds})

	var got []string
	for round := 1; round <= totalRounds; round++ {
		state := &chesspairing.TournamentState{
			Players: []chesspairing.PlayerEntry{
				{ID: "p1", DisplayName: "P1", Rating: 2600},
				{ID: "p2", DisplayName: "P2", Rating: 2500},
				{ID: "p3", DisplayName: "P3", Rating: 2400},
				{ID: "p4", DisplayName: "P4", Rating: 2300},
			},
			CurrentRound: round,
			ScoringConfig: chesspairing.ScoringConfig{
				System:  chesspairing.ScoringStandard,
				Options: map[string]any{"pointWin": 1.0},
			},
		}
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair() error: %v", round, err)
		}
		got = append(got, bakuVP(t, result.Notes))
	}

	want := "1.0 1.0 1.0 0.5 0.5 0.0 0.0 0.0 0.0"
	if strings.Join(got, " ") != want {
		t.Errorf("virtual points = %s, want %s", strings.Join(got, " "), want)
	}
}

// TestFIDEVoorbeeld_baku_ex2 verifies FIDE C.04.7 art. 1.4.4 ex. 2: an
// 11-round team competition with matchpoints (win = 2) assigns GA teams two
// virtual points in the first three rounds and one virtual point in the next
// three rounds.
func TestFIDEVoorbeeld_baku_ex2(t *testing.T) {
	acceleration := "baku"
	totalRounds := 11
	p := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds})

	var got []string
	for round := 1; round <= totalRounds; round++ {
		state := &chesspairing.TournamentState{
			Players: []chesspairing.PlayerEntry{
				{ID: "p1", DisplayName: "P1", Rating: 2600},
				{ID: "p2", DisplayName: "P2", Rating: 2500},
				{ID: "p3", DisplayName: "P3", Rating: 2400},
				{ID: "p4", DisplayName: "P4", Rating: 2300},
			},
			CurrentRound: round,
			ScoringConfig: chesspairing.ScoringConfig{
				System:  chesspairing.ScoringStandard,
				Options: map[string]any{"pointWin": 2.0},
			},
		}
		result, err := p.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair() error: %v", round, err)
		}
		got = append(got, bakuVP(t, result.Notes))
	}

	want := "2.0 2.0 2.0 1.0 1.0 1.0 0.0 0.0 0.0 0.0 0.0"
	if strings.Join(got, " ") != want {
		t.Errorf("virtual points = %s, want %s", strings.Join(got, " "), want)
	}
}

// bakuVP extracts the virtual-point value from a Baku acceleration note.
func bakuVP(t *testing.T, notes []string) string {
	t.Helper()
	for _, note := range notes {
		if !strings.HasPrefix(note, "Baku acceleration:") {
			continue
		}
		const prefix = "VP="
		idx := strings.LastIndex(note, prefix)
		if idx < 0 {
			t.Fatalf("Baku acceleration note missing VP=: %q", note)
		}
		return note[idx+len(prefix):]
	}
	t.Fatalf("no Baku acceleration note in %v", notes)
	return ""
}
