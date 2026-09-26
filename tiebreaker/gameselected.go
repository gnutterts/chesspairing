// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("ge", func() chesspairing.TieBreaker { return &GamesElected{} })
}

// GamesElected computes FIDE C.07 Article 7 (GE): the number of rounds less
// requested byes and forfeit losses.
type GamesElected struct{}

func (g *GamesElected) ID() string   { return "ge" }
func (g *GamesElected) Name() string { return "Games Elected to Play" }

func (g *GamesElected) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)
	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		value := table.totalRounds
		for _, record := range table.records[score.PlayerID] {
			if record.IsVUR {
				value--
			}
		}
		result[i] = chesspairing.TieBreakValue{PlayerID: score.PlayerID, Value: float64(value)}
	}
	return result, nil
}
