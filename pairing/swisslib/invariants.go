// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

// AssertPairingInvariants checks universal structural properties of a pairing result.
// Call this from every integration test to catch bugs even when exact expected
// output is unknown.
//
// Properties checked:
//   - Every active player appears in exactly one pairing or one bye (completeness)
//   - No player appears more than once (uniqueness)
//   - No pairing is a rematch of a previous round (no-rematch, C1 equivalent)
//   - Board numbers are sequential starting from 1
//   - Bye type is ByePAB
//   - No inactive player appears in pairings or byes
func AssertPairingInvariants(t *testing.T, state *chesspairing.TournamentState, result *chesspairing.PairingResult) {
	t.Helper()

	activeIDs := make(map[string]bool)
	for _, p := range state.Players {
		if p.Active {
			activeIDs[p.ID] = true
		}
	}

	// Uniqueness: no player appears more than once.
	seen := make(map[string]int)
	for i, gp := range result.Pairings {
		seen[gp.WhiteID]++
		seen[gp.BlackID]++
		if gp.WhiteID == gp.BlackID {
			t.Errorf("pairing[%d]: player %s paired against themselves", i, gp.WhiteID)
		}
	}
	for _, bye := range result.Byes {
		seen[bye.PlayerID]++
	}
	for id, count := range seen {
		if count != 1 {
			t.Errorf("player %s appears %d times in pairings+byes (expected 1)", id, count)
		}
	}

	// Completeness: every active player is paired or has a bye.
	for id := range activeIDs {
		if seen[id] == 0 {
			t.Errorf("active player %s not found in pairings or byes", id)
		}
	}

	// No inactive player paired.
	for id := range seen {
		if !activeIDs[id] {
			t.Errorf("inactive player %s found in pairings or byes", id)
		}
	}

	// Board numbers sequential from 1.
	for i, gp := range result.Pairings {
		expected := i + 1
		if gp.Board != expected {
			t.Errorf("pairing[%d]: expected board %d, got %d", i, expected, gp.Board)
		}
	}

	// No rematches.
	prevPairs := make(map[[2]string]bool)
	for _, rd := range state.Rounds {
		for _, g := range rd.Games {
			if g.IsForfeit {
				continue // forfeits excluded from pairing history
			}
			key := CanonicalPairKey(g.WhiteID, g.BlackID)
			prevPairs[key] = true
		}
	}
	for _, gp := range result.Pairings {
		key := CanonicalPairKey(gp.WhiteID, gp.BlackID)
		if prevPairs[key] {
			t.Errorf("rematch detected: %s vs %s", gp.WhiteID, gp.BlackID)
		}
	}

	// Bye type check.
	for _, bye := range result.Byes {
		if bye.Type != chesspairing.ByePAB {
			t.Errorf("bye for %s has type %v, expected ByePAB", bye.PlayerID, bye.Type)
		}
	}
}
