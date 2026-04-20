// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import "github.com/gnutterts/chesspairing"

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
