// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("black-wins", func() chesspairing.TieBreaker { return &BlackWins{} })
}

// BlackWins computes the number of games won with the Black pieces
// (FIDE Art. 7.4, BWG).
//
// Only over-the-board wins count in Swiss tournaments. In tournaments with
// pre-determined pairings, FIDE Art. 15.2 treats a forfeit win as a regular
// game, so a forfeit win with Black counts there.
//
// FIDE Category B tiebreaker.
type BlackWins struct{}

func (bw *BlackWins) ID() string   { return "black-wins" }
func (bw *BlackWins) Name() string { return "Black Wins" }

func (bw *BlackWins) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	blackWinCount := make(map[string]float64, len(scores))

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			if game.Result == chesspairing.ResultBlackWins ||
				state.PairingConfig.System == chesspairing.PairingRoundRobin && game.Result == chesspairing.ResultForfeitBlackWins {
				blackWinCount[game.BlackID]++
			}
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    blackWinCount[ps.PlayerID],
		}
	}
	return result, nil
}
