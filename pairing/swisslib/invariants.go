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
//   - Algorithmically allocated byes have type ByePAB; pre-assigned byes
//     (declared via state.PreAssignedByes) keep their declared type
//   - No inactive player appears in pairings; inactive players may carry a
//     pre-assigned bye (e.g. a withdrawn player flagged absent for the round)
func AssertPairingInvariants(t *testing.T, state *chesspairing.TournamentState, result *chesspairing.PairingResult) {
	t.Helper()

	activeIDs := make(map[string]bool)
	for _, p := range state.Players {
		if state.IsActiveInRound(p.ID, state.CurrentRound) {
			activeIDs[p.ID] = true
		}
	}

	preAssigned := make(map[string]chesspairing.ByeType, len(state.PreAssignedByes))
	for _, b := range state.PreAssignedByes {
		preAssigned[b.PlayerID] = b.Type
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

	// No inactive player paired. Pre-assigned byes are exempt — a caller
	// may pre-declare an absence for a withdrawn player.
	for id := range seen {
		if !activeIDs[id] {
			if _, ok := preAssigned[id]; ok {
				continue
			}
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

	// No rematches. Walk only completed rounds (1..CurrentRound-1).
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	prevPairs := make(map[[2]string]bool)
	for ri := 0; ri < historyEnd; ri++ {
		for _, g := range state.Rounds[ri].Games {
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

	// Bye type check. Pre-assigned byes keep the declared type;
	// algorithmically allocated byes must be ByePAB.
	for _, bye := range result.Byes {
		if pre, ok := preAssigned[bye.PlayerID]; ok {
			if bye.Type != pre {
				t.Errorf("bye for %s has type %v, expected pre-assigned %v", bye.PlayerID, bye.Type, pre)
			}
			continue
		}
		if bye.Type != chesspairing.ByePAB {
			t.Errorf("bye for %s has type %v, expected ByePAB", bye.PlayerID, bye.Type)
		}
	}
}
