// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"fmt"

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
//   - Determine the qualifying threshold: koyaLimit percent of the maximum score
//   - Find all opponents whose score >= threshold
//   - Sum the player's results against those opponents using the configured points
type Koya struct{}

func (k *Koya) ID() string   { return "koya" }
func (k *Koya) Name() string { return "Koya System" }

func (k *Koya) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	limit := 50.0
	if _, ok := state.ScoringConfig.Options["koyaLimit"]; ok {
		value, valid := chesspairing.GetFloat64(state.ScoringConfig.Options, "koyaLimit")
		if !valid {
			return nil, fmt.Errorf("koyaLimit must be a percentage (float64)")
		}
		limit = value
	}
	if limit < 0 || limit > 100 {
		return nil, fmt.Errorf("koya limit must be between 0 and 100 percent")
	}
	opts := standard.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	win := *opts.PointWin
	if state.PairingConfig.System == chesspairing.PairingDoubleSwiss {
		win *= 2
	}
	threshold := float64(len(state.Rounds)) * win * limit / 100

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
