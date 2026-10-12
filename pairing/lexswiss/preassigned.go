// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"fmt"

	"github.com/gnutterts/chesspairing"
)

// FilterPreAssignedByes mirrors swisslib.FilterPreAssignedByes for
// lexicographic Swiss systems (team, doubleswiss). See the swisslib
// version for details.
func FilterPreAssignedByes(state *chesspairing.TournamentState) (*chesspairing.TournamentState, []chesspairing.ByeEntry) {
	if len(state.PreAssignedByes) == 0 {
		return state, nil
	}

	skip := make(map[string]bool, len(state.PreAssignedByes))
	for _, b := range state.PreAssignedByes {
		skip[b.PlayerID] = true
	}

	filtered := make([]chesspairing.PlayerEntry, 0, len(state.Players))
	for _, p := range state.Players {
		if !skip[p.ID] {
			filtered = append(filtered, p)
		}
	}

	byes := make([]chesspairing.ByeEntry, len(state.PreAssignedByes))
	copy(byes, state.PreAssignedByes)

	clone := *state
	clone.Players = filtered
	clone.PreAssignedByes = nil
	return &clone, byes
}

// ValidatePreAssignedPAB checks that no participant with a pre-assigned ByePAB
// is ineligible under C2 (C.04.5 Article 2.1.2 / Basic Rules Art. 4). Team
// states are validated by the Team Swiss pairer itself (pairing/team), which
// applies its own option-derived eligibility rules.
func ValidatePreAssignedPAB(state *chesspairing.TournamentState) error {
	if len(state.PreAssignedByes) == 0 {
		return nil
	}
	var hasPAB bool
	for _, b := range state.PreAssignedByes {
		if b.Type == chesspairing.ByePAB {
			hasPAB = true
			break
		}
	}
	if !hasPAB {
		return nil
	}

	ineligibility := ComputePABIneligibility(state)
	for _, b := range state.PreAssignedByes {
		if b.Type != chesspairing.ByePAB {
			continue
		}
		if inel := ineligibility[b.PlayerID]; inel.Any() {
			return fmt.Errorf("participant %s is ineligible for pairing-allocated bye under C2 (C.04.5 2.1.2)", b.PlayerID)
		}
	}
	return nil
}
