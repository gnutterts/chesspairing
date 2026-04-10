// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// ComputeARO returns the Average Rating of Opponents for a player.
// Per Dubov Art. 1.7: arithmetic mean of the ratings of all opponents
// the player has played. Returns 0 if the player has no opponents.
//
// The opponents list comes from PlayerState.Opponents, which already
// excludes forfeits (see swisslib.BuildPlayerStates).
func ComputeARO(player *swisslib.PlayerState, ratings map[string]int) float64 {
	if len(player.Opponents) == 0 {
		return 0
	}

	var total float64
	for _, oppID := range player.Opponents {
		total += float64(ratings[oppID])
	}

	return total / float64(len(player.Opponents))
}

// BuildRatingMap creates a player ID -> rating lookup from a PlayerState slice.
func BuildRatingMap(players []swisslib.PlayerState) map[string]int {
	m := make(map[string]int, len(players))
	for i := range players {
		m[players[i].ID] = players[i].Rating
	}
	return m
}

// SortByAROAscending sorts players by ascending ARO, breaking ties by
// ascending TPN (Art. 3.2.5). This is the G1/S1 sorting rule.
func SortByAROAscending(players []*swisslib.PlayerState, ratings map[string]int) {
	// Pre-compute AROs to avoid recomputation during sort.
	aros := make(map[string]float64, len(players))
	for _, p := range players {
		aros[p.ID] = ComputeARO(p, ratings)
	}

	sort.SliceStable(players, func(i, j int) bool {
		ai, aj := aros[players[i].ID], aros[players[j].ID]
		if ai != aj {
			return ai < aj
		}
		return players[i].TPN < players[j].TPN
	})
}
