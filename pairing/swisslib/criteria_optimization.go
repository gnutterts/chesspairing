// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

// floatAtRound returns the float direction a player had at a specific round
// (0-indexed into FloatHistory). Returns FloatNone if history is too short.
func floatAtRound(p *PlayerState, roundIdx int) Float {
	if roundIdx < 0 || roundIdx >= len(p.FloatHistory) {
		return FloatNone
	}
	return p.FloatHistory[roundIdx]
}
