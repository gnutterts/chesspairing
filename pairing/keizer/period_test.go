// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestMinRoundsBetweenRepeatsAlone verifies the historical minimum-gap rule
// without periods: after a round-4 encounter, round 5 is blocked and round 6
// is the earliest legal re-pairing when minRoundsBetweenRepeats=2.
func TestMinRoundsBetweenRepeatsAlone(t *testing.T) {
	allow := true
	minRounds := 2
	opts := Options{AllowRepeatPairings: &allow, MinRoundsBetweenRepeats: &minRounds}.WithDefaults()
	history := buildHistory([]chesspairing.RoundData{
		{Number: 4, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "X", Result: chesspairing.ResultDraw}}},
	}, *opts.ForfeitCountsAsMet)

	if canPair("A", "X", opts, history, 5) {
		t.Error("round 5 must be blocked by minRoundsBetweenRepeats=2")
	}
	if !canPair("A", "X", opts, history, 6) {
		t.Error("round 6 must be allowed by minRoundsBetweenRepeats=2")
	}
}

// TestPeriodRepeatRules verifies that noRepeatWithinPeriod extends the minimum
// gap: after a round-4 encounter with periodLength=6, rounds 5 and 6 are
// blocked (round 6 by the same period) and round 7 is the earliest legal
// re-pairing.
func TestPeriodRepeatRules(t *testing.T) {
	allow := true
	minRounds := 2
	periodLength := 6
	noRepeatWithinPeriod := true
	opts := Options{
		AllowRepeatPairings:     &allow,
		MinRoundsBetweenRepeats: &minRounds,
		PeriodLength:            &periodLength,
		NoRepeatWithinPeriod:    &noRepeatWithinPeriod,
	}.WithDefaults()
	history := buildHistory([]chesspairing.RoundData{
		{Number: 4, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "X", Result: chesspairing.ResultDraw}}},
	}, *opts.ForfeitCountsAsMet)

	if canPair("A", "X", opts, history, 5) {
		t.Error("round 5 must be blocked by minRoundsBetweenRepeats=2")
	}
	if canPair("A", "X", opts, history, 6) {
		t.Error("round 6 must be blocked by the same period (rounds 1-6)")
	}
	if !canPair("A", "X", opts, history, 7) {
		t.Error("round 7 must be allowed (next period and gap >= 2)")
	}
}
