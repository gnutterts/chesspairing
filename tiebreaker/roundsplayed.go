// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("rounds-played", func() chesspairing.TieBreaker { return &RoundsPlayed{} })
}

// RoundsPlayed computes the number of rounds effectively played
// (FIDE Art. 7.6, REP).
//
// Unplayed rounds are subtracted from the total round count:
//   - Half-point bye (ByeHalf)
//   - Zero-point bye (ByeZero)
//   - Absent (ByeAbsent or not appearing in round at all)
//   - Forfeit loss
//
// PAB (pairing-allocated bye) and forfeit wins count as played.
//
// FIDE Category B tiebreaker.
type RoundsPlayed struct{}

func (rp *RoundsPlayed) ID() string   { return "rounds-played" }
func (rp *RoundsPlayed) Name() string { return "Rounds Played" }

func (rp *RoundsPlayed) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	totalRounds := len(state.Rounds)

	// Build set of active player IDs.
	activeSet := make(map[string]bool, len(state.Players))
	for _, p := range state.Players {
		if p.Active {
			activeSet[p.ID] = true
		}
	}

	// Count unplayed rounds per player.
	unplayed := make(map[string]int, len(scores))

	for _, round := range state.Rounds {
		played := make(map[string]bool)

		for _, game := range round.Games {
			played[game.WhiteID] = true
			played[game.BlackID] = true

			// Forfeit loss counts as unplayed for the loser.
			switch game.Result {
			case chesspairing.ResultForfeitWhiteWins:
				unplayed[game.BlackID]++
			case chesspairing.ResultForfeitBlackWins:
				unplayed[game.WhiteID]++
			case chesspairing.ResultDoubleForfeit:
				unplayed[game.WhiteID]++
				unplayed[game.BlackID]++
			}
		}

		for _, bye := range round.Byes {
			played[bye.PlayerID] = true

			// Half-bye, zero-bye, and absent-bye are unplayed.
			// PAB is played (player receives full point).
			switch bye.Type {
			case chesspairing.ByeHalf, chesspairing.ByeZero, chesspairing.ByeAbsent:
				unplayed[bye.PlayerID]++
			}
		}

		// Active players not in games or byes are absent (unplayed).
		for id := range activeSet {
			if !played[id] {
				unplayed[id]++
			}
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		rep := float64(totalRounds - unplayed[ps.PlayerID])
		if rep < 0 {
			rep = 0
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    rep,
		}
	}
	return result, nil
}
