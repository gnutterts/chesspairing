// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("avg-opponent-tpr", func() chesspairing.TieBreaker { return &AvgOpponentTPR{} })
}

// AvgOpponentTPR computes the average of opponents' Tournament Performance
// Ratings (FIDE Art. 10.4, APRO).
//
// For each player, this first computes TPR for every opponent, then averages
// those values. The result is rounded to the nearest whole number (0.5
// rounded up).
//
// FIDE Category D tiebreaker.
type AvgOpponentTPR struct{}

func (a *AvgOpponentTPR) ID() string   { return "avg-opponent-tpr" }
func (a *AvgOpponentTPR) Name() string { return "Avg Opponent TPR" }

func (a *AvgOpponentTPR) Compute(ctx context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	// Compute TPR for every player first.
	tpr := &PerformanceRating{}
	tprValues, err := tpr.Compute(ctx, state, scoresForAllPlayers(table))
	if err != nil {
		return nil, err
	}
	tprMap := make(map[string]float64, len(tprValues))
	for _, v := range tprValues {
		tprMap[v.PlayerID] = v.Value
	}

	// Average opponents' TPR.
	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		games := playedRecords(table.records[ps.PlayerID])
		if len(games) == 0 {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
			continue
		}

		var totalOppTPR float64
		for _, game := range games {
			totalOppTPR += tprMap[game.OpponentID]
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    roundHalfUp(totalOppTPR / float64(len(games))),
		}
	}
	return result, nil
}
