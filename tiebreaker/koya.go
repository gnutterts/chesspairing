// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

func init() {
	Register("koya", func() chesspairing.TieBreaker { return &Koya{} })
}

// Koya computes the Koya system tiebreaker.
//
// The Koya system counts the number of points scored against opponents
// who have 50% or more of the maximum possible score. In a round-robin,
// this is a useful tiebreaker because it rewards consistency against
// the stronger half of the field.
//
// Implementation:
//   - Determine the "qualifying" threshold: half the maximum possible tournament score
//   - Find all opponents whose score >= threshold
//   - Sum the player's results against those opponents using the configured points
type Koya struct{}

func (k *Koya) ID() string   { return "koya" }
func (k *Koya) Name() string { return "Koya System" }

func (k *Koya) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	// Qualifying threshold: 50% of the maximum possible tournament score.
	opts := standard.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	threshold := float64(len(state.Rounds)) * *opts.PointWin / 2

	// Build set of qualifying opponents (score >= threshold).
	qualifying := make(map[string]bool, len(scores))
	for _, ps := range scores {
		if ps.Score >= threshold {
			qualifying[ps.PlayerID] = true
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var koyaScore float64
		for _, record := range table.records[ps.PlayerID] {
			regular := record.Played || table.roundRobin && (record.Category == ForfeitWin || record.Category == ForfeitLoss)
			if !regular || !qualifying[record.OpponentID] {
				continue
			}
			koyaScore += record.Points
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    koyaScore,
		}
	}
	return result, nil
}
