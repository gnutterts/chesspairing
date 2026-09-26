// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("avg-opponent-ptp", func() chesspairing.TieBreaker { return &AvgOpponentPTP{} })
}

// AvgOpponentPTP computes the average of opponents' Performance with
// Tournament Points (FIDE Art. 10.5, APPO).
//
// For each player, this first computes PTP for every opponent, then averages
// those values. The result is rounded to the nearest whole number (0.5
// rounded up).
//
// FIDE Category D tiebreaker.
type AvgOpponentPTP struct{}

func (a *AvgOpponentPTP) ID() string   { return "avg-opponent-ptp" }
func (a *AvgOpponentPTP) Name() string { return "Avg Opponent PTP" }

func (a *AvgOpponentPTP) Compute(ctx context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	// Compute PTP for every player first.
	ptp := &PerformancePoints{}
	ptpValues, err := ptp.Compute(ctx, state, scoresForAllPlayers(table))
	if err != nil {
		return nil, err
	}
	ptpMap := make(map[string]float64, len(ptpValues))
	for _, v := range ptpValues {
		ptpMap[v.PlayerID] = v.Value
	}

	// Average opponents' PTP.
	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		games := playedRecords(table.records[ps.PlayerID])
		if len(games) == 0 {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
			continue
		}

		var totalOppPTP float64
		for _, game := range games {
			totalOppPTP += ptpMap[game.OpponentID]
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    roundHalfUp(totalOppPTP / float64(len(games))),
		}
	}
	return result, nil
}
