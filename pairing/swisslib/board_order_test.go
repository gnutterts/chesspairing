// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "testing"

func TestBoardBefore_ScoreSumAndHigherRankedTPN(t *testing.T) {
	p := func(score float64, tpn int) BoardParticipant { return BoardParticipant{Score: score, TPN: tpn} }
	// A lower-ranked opponent's small TPN must not override the score sum.
	if !BoardBefore(p(2, 8), p(1.5, 9), p(2, 1), p(0, 2)) {
		t.Fatal("higher score sum must precede lower score sum")
	}
	// The higher scorer's TPN, not the minimum TPN of the pair, decides.
	if !BoardBefore(p(2, 3), p(1, 9), p(1, 1), p(2, 4)) {
		t.Fatal("higher-ranked participant's TPN must decide equal sums")
	}
	if !BoardBefore(p(3, 9), p(0, 10), p(2, 1), p(1, 2)) {
		t.Fatal("higher pairing score must decide first")
	}
}
