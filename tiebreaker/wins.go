// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("wins", func() chesspairing.TieBreaker { return &Wins{} })
}

// Wins computes the number of games won (FIDE Art. 7.2, WON).
//
// Byes and forfeit losses are excluded. A forfeit win is a regular game in
// tournaments with pre-determined pairings (FIDE Art. 15.2) and counts
// there; in Swiss tournaments only wins actually played over the board
// count.
type Wins struct{}

func (w *Wins) ID() string   { return "wins" }
func (w *Wins) Name() string { return "Games Won" }

func (w *Wins) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var wins float64
		for _, record := range table.records[ps.PlayerID] {
			if (record.Played || table.roundRobin && record.Category == ForfeitWin) && record.Points == 1 {
				wins++
			}
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    wins,
		}
	}
	return result, nil
}
