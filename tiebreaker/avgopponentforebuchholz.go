// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("avg-opponent-fore-buchholz", func() chesspairing.TieBreaker { return &AvgOpponentForeBuchholz{} })
}

// AvgOpponentForeBuchholz computes the Average of Opponents' Fore Buchholz
// tiebreaker (FIDE Art. 8.2 with Art. 8.3, AOB(FB)).
//
// Article 8.2 allows averaging the Fore Buchholz score of the opponents
// instead of their Buchholz. Reading A is applied: only opponents played
// over the board are counted, so a pending final-round opponent is excluded.
type AvgOpponentForeBuchholz struct{}

func (a *AvgOpponentForeBuchholz) ID() string   { return "avg-opponent-fore-buchholz" }
func (a *AvgOpponentForeBuchholz) Name() string { return "Avg Opponent Fore Buchholz" }

func (a *AvgOpponentForeBuchholz) Compute(ctx context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	fb := &ForeBuchholz{}
	fbValues, err := fb.Compute(ctx, state, scoresForAllPlayers(table))
	if err != nil {
		return nil, err
	}
	fbMap := make(map[string]float64, len(fbValues))
	for _, v := range fbValues {
		fbMap[v.PlayerID] = v.Value
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		games := playedRecords(table.records[ps.PlayerID])
		if len(games) == 0 {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
			continue
		}

		var totalOppFB float64
		for _, game := range games {
			totalOppFB += fbMap[game.OpponentID]
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    roundToTwoDecimals(totalOppFB / float64(len(games))),
		}
	}
	return result, nil
}
