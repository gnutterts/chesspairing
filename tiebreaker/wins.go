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
	// Article 7.2 counts games, so in a Double Swiss both games of a match
	// count separately; the per-round opponent records are not used.
	winCount := make(map[string]float64, len(scores))
	roundRobin := state.PairingConfig.System == chesspairing.PairingRoundRobin

	// Tie-breaks read WithdrawnAfterRound directly and drop the rounds
	// after it, like the per-round record model does.
	withdrawnAfter := make(map[string]int, len(state.Players))
	for _, player := range state.Players {
		if player.WithdrawnAfterRound != nil {
			withdrawnAfter[player.ID] = *player.WithdrawnAfterRound
		}
	}

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			winner := ""
			switch game.Result {
			case chesspairing.ResultWhiteWins:
				winner = game.WhiteID
			case chesspairing.ResultBlackWins:
				winner = game.BlackID
			case chesspairing.ResultForfeitWhiteWins:
				// Article 15.2: with pre-determined pairings a forfeit win
				// is a regular game; in Swiss tournaments it stays unplayed.
				if roundRobin {
					winner = game.WhiteID
				}
			case chesspairing.ResultForfeitBlackWins:
				if roundRobin {
					winner = game.BlackID
				}
			}
			if winner == "" {
				continue
			}
			if after, withdrawn := withdrawnAfter[winner]; withdrawn && round.Number > after {
				// The round is behind the player's withdrawal.
				continue
			}
			winCount[winner]++
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
