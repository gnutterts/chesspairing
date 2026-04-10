// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("win", func() chesspairing.TieBreaker { return &Win{} })
}

// Win computes the number of rounds where the participant obtained as many
// points as awarded for a win (FIDE Art. 7.1, WIN).
//
// This includes OTB wins, forfeit wins, and full-point byes (PAB).
// Half-point byes and zero-point byes do not count.
//
// FIDE Category B tiebreaker.
type Win struct{}

func (w *Win) ID() string   { return "win" }
func (w *Win) Name() string { return "Rounds Won" }

func (w *Win) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	// WIN counts rounds where the player's awarded points equal a win (1.0).
	winCount := make(map[string]float64, len(scores))

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				winCount[game.WhiteID]++
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				winCount[game.BlackID]++
			}
			// Draws, pending, double forfeits: neither player gets win-points.
		}

		for _, bye := range round.Byes {
			if bye.Type == chesspairing.ByePAB {
				winCount[bye.PlayerID]++
			}
			// Half-point and zero-point byes do not award win-points.
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    winCount[ps.PlayerID],
		}
	}
	return result, nil
}
