// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("direct-encounter", func() chesspairing.TieBreaker { return &DirectEncounter{} })
}

// DirectEncounter computes the direct encounter (head-to-head) tiebreaker.
//
// For each group of tied players (same primary score), the direct encounter
// value is the score from games played ONLY against other members of the
// tied group. Games against non-tied players are ignored.
//
// If a player is not tied with anyone, their direct encounter value is 0
// (it doesn't matter since there's nothing to break).
type DirectEncounter struct{}

func (de *DirectEncounter) ID() string   { return "direct-encounter" }
func (de *DirectEncounter) Name() string { return "Direct Encounter" }

func (de *DirectEncounter) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	// Group players by their primary score to identify tied groups.
	groups := make(map[float64]map[string]bool)
	for _, ps := range scores {
		if groups[ps.Score] == nil {
			groups[ps.Score] = make(map[string]bool)
		}
		groups[ps.Score][ps.PlayerID] = true
	}

	// For each player, compute their score from games against tied opponents.
	deScores := make(map[string]float64, len(scores))
	for _, ps := range scores {
		tiedGroup := groups[ps.Score]
		if len(tiedGroup) <= 1 {
			continue // not tied with anyone
		}

		var deScore float64
		for _, record := range table.records[ps.PlayerID] {
			if !tiedGroup[record.OpponentID] {
				continue // skip games against non-tied players
			}
			// C.07:2026 Article 15.2: a forfeit only counts as a regular game
			// with pre-determined pairings (round robin). In Swiss tournaments
			// forfeits are not head-to-head results.
			forfeitAsGame := table.roundRobin && (record.Category == ForfeitWin || record.Category == ForfeitLoss)
			if !record.Played && !forfeitAsGame {
				continue
			}
			deScore += record.Points
		}
		deScores[ps.PlayerID] = deScore
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    deScores[ps.PlayerID],
		}
	}
	return result, nil
}
