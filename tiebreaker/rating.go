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

// roundToTwoDecimals rounds x to two decimal places, rounding halves up.
// AOB (Article 8.2) values are published with two decimals.
func roundToTwoDecimals(x float64) float64 {
	return math.Round(x*100) / 100
}

func init() {
	Register("aro", func() chesspairing.TieBreaker { return &ARO{} })
	Register("aro-cut1", func() chesspairing.TieBreaker { return &ARO{cut1: true} })
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
type ARO struct {
	cut1 bool
}

func (a *ARO) ID() string {
	if a.cut1 {
		return "aro-cut1"
	}
	return "aro"
}

func (a *ARO) Name() string {
	if a.cut1 {
		return "Avg Rating of Opponents Cut-1"
	}
	return "Avg Rating of Opponents"
}

func (a *ARO) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		games := playedRecords(table.records[ps.PlayerID])
		if len(games) == 0 || (a.cut1 && len(games) == 1) {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
			continue
		}

		var totalRating float64
		lowestRating := games[0].OppRating
		for _, game := range games {
			totalRating += float64(game.OppRating)
			if game.OppRating < lowestRating {
				lowestRating = game.OppRating
			}
		}
		divisor := len(games)
		if a.cut1 {
			totalRating -= float64(lowestRating)
			divisor--
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    roundHalfUp(totalRating / float64(divisor)),
		}
	}
	return result, nil
}
