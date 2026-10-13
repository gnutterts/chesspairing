// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"fmt"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("player-rating", func() chesspairing.TieBreaker { return &PlayerRating{} })
}

// PlayerRating computes the player's own rating tiebreaker
// (FIDE Art. 10.6, RTNG).
//
// The value is the player's rating. Higher rating ranks higher.
//
// FIDE Category B tiebreaker.
type PlayerRating struct{}

func (pr *PlayerRating) ID() string   { return "player-rating" }
func (pr *PlayerRating) Name() string { return "Player Rating" }

func (pr *PlayerRating) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	override, hasOverride := unratedOverride(state)
	// Build rating lookup.
	ratings := make(map[string]int, len(state.Players))
	for _, p := range state.Players {
		ratings[p.ID] = p.Rating
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		rating := float64(ratings[ps.PlayerID])
		if rating == 0 {
			if !hasOverride {
				return nil, fmt.Errorf("unrated player %s (rating 0): C.07 Art. 10 drops this tie-break unless the tournament regulations define the handling of unrated players (scoring option \"unratedRating\")", playerName(state, ps.PlayerID))
			}
			rating = override
		}
		result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: rating}
	}
	return result, nil
}
