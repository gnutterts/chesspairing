// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"math"

	"github.com/gnutterts/chesspairing"
)

// roundHalfUp rounds x to the nearest whole number, rounding halves up
// (0.5 rounds up), as required by FIDE C.07 Article 10.1, 10.4 and 10.5.
func roundHalfUp(x float64) float64 {
	return math.Floor(x + 0.5)
}

func init() {
	Register("aro", func() chesspairing.TieBreaker { return &ARO{} })
}

// ARO computes the Average Rating of Opponents tiebreaker.
//
// The value is the arithmetic mean of the ratings of all opponents
// the player has played against, rounded to the nearest whole number
// (0.5 rounded up) per FIDE C.07 Article 10.1. Byes and absences are
// excluded (they have no opponent).
//
// This tiebreaker rewards playing against a stronger field and is
// categorized as FIDE Category D.
type ARO struct{}

func (a *ARO) ID() string   { return "aro" }
func (a *ARO) Name() string { return "Avg Rating of Opponents" }

func (a *ARO) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	data := buildOpponentData(state, scores)

	// Build rating lookup.
	ratings := make(map[string]int, len(state.Players))
	for _, p := range state.Players {
		ratings[p.ID] = p.Rating
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		games := data.playerGames[ps.PlayerID]
		if len(games) == 0 {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
			continue
		}

		var totalRating float64
		for _, g := range games {
			totalRating += float64(ratings[g.opponentID])
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    roundHalfUp(totalRating / float64(len(games))),
		}
	}
	return result, nil
}
