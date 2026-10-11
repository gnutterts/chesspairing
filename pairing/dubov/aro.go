// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// ComputeARO returns the Article 1.7.1 Average Rating of Opponents. It is the
// arithmetic mean of played opponents' ratings, rounded to the nearest integer
// (with halves rounded up), or zero when no game has been played.
func ComputeARO(player *swisslib.PlayerState, ratings map[string]int) int {
	if len(player.Opponents) == 0 {
		return 0
	}

	total := 0
	for _, opponentID := range player.Opponents {
		total += ratings[opponentID]
	}
	return (2*total + len(player.Opponents)) / (2 * len(player.Opponents))
}

// BuildRatingMap creates a player ID -> rating lookup from player states.
func BuildRatingMap(players []swisslib.PlayerState) map[string]int {
	m := make(map[string]int, len(players))
	for i := range players {
		m[players[i].ID] = players[i].Rating
	}
	return m
}

// BuildRatingMapFromState includes withdrawn opponents, which still count in
// the ARO of a game played over the board.
func BuildRatingMapFromState(state *chesspairing.TournamentState) map[string]int {
	m := make(map[string]int, len(state.Players))
	for i := range state.Players {
		m[state.Players[i].ID] = state.Players[i].Rating
	}
	return m
}

// SortByAROAscending applies Article 3.2.5.
func SortByAROAscending(players []*swisslib.PlayerState, ratings map[string]int) {
	aros := make(map[string]int, len(players))
	for _, player := range players {
		aros[player.ID] = ComputeARO(player, ratings)
	}
	sort.SliceStable(players, func(i, j int) bool {
		if aros[players[i].ID] != aros[players[j].ID] {
			return aros[players[i].ID] < aros[players[j].ID]
		}
		return swisslib.EffectivePairingNumber(players[i]) < swisslib.EffectivePairingNumber(players[j])
	})
}
