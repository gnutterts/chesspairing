// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// LimByeSelector selects the bye player per Lim system rules (Art. 1.1):
// the player with the lowest rank (highest TPN) in the lowest scoregroup,
// who has not already received a PAB (Basic Rules Art. 3).
type LimByeSelector struct{}

// SelectBye returns the player to receive the bye, or nil if all have
// already received one.
func (s LimByeSelector) SelectBye(players []*swisslib.PlayerState) *swisslib.PlayerState {
	// Filter to players who haven't received a bye.
	var eligible []*swisslib.PlayerState
	for _, p := range players {
		if !p.ByeReceived {
			eligible = append(eligible, p)
		}
	}
	if len(eligible) == 0 {
		return nil
	}

	// Sort by score ascending, then TPN descending (lowest rank = highest TPN).
	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].Score != eligible[j].Score {
			return eligible[i].Score < eligible[j].Score
		}
		return eligible[i].TPN > eligible[j].TPN
	})

	return eligible[0]
}

// Compile-time check.
var _ swisslib.ByeSelector = LimByeSelector{}
