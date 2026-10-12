// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestMedianScoreForDecimalWin(t *testing.T) {
	// 0.1 * 3 / 2 is 0.15000000000000002 in floating point; scores on the
	// median are normalised to 0.15, so the median must be too.
	if got := medianScoreFor(0.1, 3); got != 0.15 || got != swisslib.ScoreGroupKey(0.05+0.1) {
		t.Fatalf("medianScoreFor(0.1, 3) = %v, want 0.15", got)
	}
	if got := medianScoreFor(1, 5); got != 2.5 {
		t.Fatalf("medianScoreFor(1, 5) = %v, want 2.5", got)
	}
}
