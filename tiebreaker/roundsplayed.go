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
// (FIDE Art. 7.6, REP). It uses the same per-round record model as Games
// Elected: every round that is not a voluntary unplayed round counts once.
//
// Unplayed rounds are subtracted from the total round count:
//   - Half-point bye (ByeHalf)
//   - Zero-point bye (ByeZero)
//   - Absent (ByeAbsent or not appearing in round at all)
//   - Excused absence (ByeExcused)
//   - Club commitment (ByeClubCommitment)
//   - Forfeit loss
//   - Rounds after a withdrawal (a zero-point bye under FIDE Art. 16.1.1)
//   - Rounds before a late join
//
// PAB (pairing-allocated bye) and forfeit wins count as played. A pending
// game is not yet an unplayed round, so its round counts as played.
//
// FIDE Category B tiebreaker.
type RoundsPlayed struct{}

func (rp *RoundsPlayed) ID() string   { return "rounds-played" }
func (rp *RoundsPlayed) Name() string { return "Rounds Played" }

func (rp *RoundsPlayed) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
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
