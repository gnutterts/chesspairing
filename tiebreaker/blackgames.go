// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("black-games", func() chesspairing.TieBreaker { return &BlackGames{} })
}

// BlackGames computes the number of games played over the board with the
// Black pieces (FIDE Art. 7.3, BPG).
//
// Pending games are never counted. Forfeit games are unplayed rounds in Swiss
// tournaments, but FIDE Art. 15.2 treats a forfeit win as a regular game in
// tournaments with pre-determined pairings, so a forfeit win with Black
// counts there; forfeit losses remain excluded.
//
// FIDE Category B tiebreaker.
type BlackGames struct{}

func (bg *BlackGames) ID() string   { return "black-games" }
func (bg *BlackGames) Name() string { return "Games with Black" }

func (bg *BlackGames) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	// Count games with Black per FIDE Art. 7.3: completed OTB games always
	// count, and in round robins a forfeit win with Black also counts.
	blackCount := make(map[string]float64, len(scores))

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins, chesspairing.ResultDraw:
				blackCount[game.BlackID]++
			case chesspairing.ResultForfeitBlackWins:
				if state.PairingConfig.System == chesspairing.PairingRoundRobin {
					blackCount[game.BlackID]++
				}
			}
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    blackCount[ps.PlayerID],
		}
	}
	return result, nil
}
