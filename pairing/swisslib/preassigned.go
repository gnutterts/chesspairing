// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "github.com/gnutterts/chesspairing"

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
