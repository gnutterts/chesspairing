// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"fmt"
	"strings"
	"testing"
)

// TestFIDEExample_baku_1 verifies the first worked example in FIDE C.04.7,
// Article 1.4.4.
//
// Source text: "In a nine-round individual tournament that uses the standard
// scoring point system, the accelerated rounds are five. The players in GA
// are assigned one virtual point in the first three rounds, and half virtual
// point in the next two rounds."
func TestFIDEExample_baku_1(t *testing.T) {
	const totalRounds = 9

	// Transcription guard: the source uses a nine-round tournament.
	if totalRounds != 9 {
		t.Fatalf("input must be a nine-round tournament, got %d", totalRounds)
	}

	accelerated, fullVP, halfVP := BakuAccelerationRounds(totalRounds)
	fideRounds := "5 (3 full, 2 half)"
	onsRounds := fmt.Sprintf("%d (%d full, %d half)", accelerated, fullVP, halfVP)
	if accelerated != 5 || fullVP != 3 || halfVP != 2 {
		t.Errorf("baku-1.4.4-ex1-rounds: got %s, want FIDE %s", onsRounds, fideRounds)
	}

	// Virtual points per round for a GA participant, rounds 1..9.
	var onsParts []string
	for r := 1; r <= totalRounds; r++ {
		onsParts = append(onsParts, fmt.Sprintf("%.1f", BakuVirtualPoints(1.0, totalRounds, r, true)))
	}
	ons := strings.Join(onsParts, " ")
	fide := "1.0 1.0 1.0 0.5 0.5 0.0 0.0 0.0 0.0"

	if ons != fide {
		t.Errorf("baku-1.4.4-ex1-vp: got %s, want FIDE %s", ons, fide)
	}
}

// Cases pending: baku-1.4.4-ex2-vp — see FIDE C.04.7 Article 1.4.4,
// virtual points 2.0 2.0 2.0 1.0 1.0 1.0 0.0 0.0 0.0 0.0 0.0; activated by B06.
//
// TestFIDEExample_baku_2 verifies the second worked example in FIDE C.04.7,
// Article 1.4.4.
//
// Source text: "In an 11-round team competition with matchpoints as the
// primary score (2 MP for win, 1 MP for draw), the teams in GA are assigned
// two virtual matchpoints in the first three rounds and one virtual
// matchpoint in the next three rounds."
//
// The round structure (six accelerated rounds, three full-VP and three
// half-VP rounds) is rebuilt with BakuAccelerationRounds. BakuVirtualPoints
// is called with the matchpoint win value (2.0), so its magnitude matches the
// matchpoint values in the source and is hard-asserted against them.
func TestFIDEExample_baku_2(t *testing.T) {
	const totalRounds = 11

	// Transcription guard: the source uses an 11-round team competition.
	if totalRounds != 11 {
		t.Fatalf("input must be an 11-round competition, got %d", totalRounds)
	}

	accelerated, fullVP, halfVP := BakuAccelerationRounds(totalRounds)
	fideRounds := "6 (3 full, 3 half)"
	onsRounds := fmt.Sprintf("%d (%d full, %d half)", accelerated, fullVP, halfVP)
	if accelerated != 6 || fullVP != 3 || halfVP != 3 {
		t.Errorf("baku-1.4.4-ex2-rounds: got %s, want FIDE %s", onsRounds, fideRounds)
	}

	// Virtual points per round for a GA team, rounds 1..11.
	// FIDE (matchpoints): 2 in the first three rounds, 1 in the next three.
	var onsParts []string
	for r := 1; r <= totalRounds; r++ {
		onsParts = append(onsParts, fmt.Sprintf("%.1f", BakuVirtualPoints(2.0, totalRounds, r, true)))
	}
	ons := strings.Join(onsParts, " ")
	fide := "2.0 2.0 2.0 1.0 1.0 1.0 0.0 0.0 0.0 0.0 0.0"

	if ons != fide {
		t.Fatalf("baku-1.4.4-ex2-vp: virtual points = %s, want FIDE %s", ons, fide)
	}
}
