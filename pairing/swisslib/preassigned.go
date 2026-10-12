// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"fmt"

	"github.com/gnutterts/chesspairing"
)

// FilterPreAssignedByes returns a shallow-copied TournamentState with the
// players named in state.PreAssignedByes removed from the Players slice,
// together with the slice of pre-assigned ByeEntry values. Pairers call
// this at the top of Pair() so BuildPlayerStates and the matching engine
// never see those players, then prepend the returned bye entries to
// PairingResult.Byes.
//
// If state has no pre-assigned byes, the returned state is the original
// pointer and the bye slice is nil — zero allocation in the common path.
//
// PreAssignedByes is assumed to have been validated upstream (see
// chesspairing.TournamentState.Validate).
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

// ValidatePreAssignedPAB checks that no player with a pre-assigned ByePAB is
// ineligible under FIDE C.04.3 2.1.2 [C2] (Basic Rules Art. 4). A participant
// who has already received a pairing-allocated bye, or has already scored in one
// single round, without playing, as many points as rewarded for a win, shall
// not receive the pairing-allocated bye.
//
// Other bye types (half-point, zero-point, full-point requested byes) are
// unchanged and ignored here.
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
		if b.Type == chesspairing.ByePAB {
			if inel := ineligibility[b.PlayerID]; inel.Any() {
				return fmt.Errorf("player %s is ineligible for pairing-allocated bye under C2 (C.04.3 2.1.2)", b.PlayerID)
			}
		}
	}
	return nil
}
