// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("performance-rating", func() chesspairing.TieBreaker { return &PerformanceRating{} })
}

// PerformanceRating computes the tournament performance rating
// (FIDE Art. 10.2, TPR).
//
// TPR = roundHalfUp(ARO + dp(p)), where:
//   - ARO is the average rating of opponents, rounded to the nearest whole
//     number (0.5 rounded up)
//   - p = board points / games (fractional score over the board)
//   - dp(p) is the rating difference dp from the FIDE B.02 conversion table
//
// Board points are taken from the games actually played over the board;
// byes and forfeits contribute no board points. For players with no games,
// TPR = 0. ARO is first rounded to the nearest whole number (0.5 rounded
// up), the rating difference dp from the conversion table is then added to
// it, and the final result is rounded again to the nearest whole number
// (0.5 rounded up), exactly as roundHalfUp(aro + dp) does.
//
// FIDE Category D tiebreaker.
type PerformanceRating struct{}

func (tpr *PerformanceRating) ID() string   { return "performance-rating" }
func (tpr *PerformanceRating) Name() string { return "Performance Rating" }

func (tpr *PerformanceRating) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		games := playedRecords(table.records[ps.PlayerID])
		if len(games) == 0 {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
			continue
		}

		// ARO: average rating of opponents, rounded to the nearest whole
		// number (0.5 rounded up) per FIDE C.07 Article 10.1.
		var totalRating float64
		for _, game := range games {
			totalRating += float64(game.OppRating)
		}
		aro := roundHalfUp(totalRating / float64(len(games)))

		// Fractional score: points scored in games played over the board
		// divided by the number of games. playerGames already excludes
		// forfeits, and byes never produce game entries, so this counts
		// only board points from actual games.
		var boardPoints float64
		for _, game := range games {
			boardPoints += game.Points
		}
		p := boardPoints / float64(len(games))
		if p > 1.0 {
			p = 1.0
		}
		if p < 0.0 {
			p = 0.0
		}

		dp := dpFromP(p)
		tprValue := roundHalfUp(aro + dp)

		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    tprValue,
		}
	}
	return result, nil
}
