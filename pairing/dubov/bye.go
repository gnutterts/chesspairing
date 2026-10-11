// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// DubovByeSelector orders the eligible bye candidates per Articles 3.1.3-3.1.5:
// lowest score -> most games played -> highest TPN (lowest-ranked). Pair
// additionally applies Articles 3.1.1-3.1.2 by filtering ineligible players
// and trying candidates until the remaining field is completable.
type DubovByeSelector struct{}

// SelectBye returns the player to receive the PAB, or nil if all have
// already received one.
func (s DubovByeSelector) SelectBye(players []*swisslib.PlayerState) *swisslib.PlayerState {
	eligible := filterNoByeReceived(players)
	if len(eligible) == 0 {
		return nil
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		// 1. Lowest score first.
		if eligible[i].Score != eligible[j].Score {
			return eligible[i].Score < eligible[j].Score
		}
		// 2. Most games played first.
		gi := swisslib.GamesPlayed(eligible[i])
		gj := swisslib.GamesPlayed(eligible[j])
		if gi != gj {
			return gi > gj
		}
		// 3. Highest PairingNumber (lowest ranking) first.
		return swisslib.EffectivePairingNumber(eligible[i]) > swisslib.EffectivePairingNumber(eligible[j])
	})

	return eligible[0]
}

// filterNoByeReceived returns players who have not yet received a PAB.
func filterNoByeReceived(players []*swisslib.PlayerState) []*swisslib.PlayerState {
	var eligible []*swisslib.PlayerState
	for _, p := range players {
		if !p.PABIneligible.Any() {
			eligible = append(eligible, p)
		}
	}
	return eligible
}

// Compile-time check.
var _ swisslib.ByeSelector = DubovByeSelector{}
