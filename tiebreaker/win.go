// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

func init() {
	Register("win", func() chesspairing.TieBreaker { return &Win{} })
}

// Win computes the number of rounds where the participant obtained as many
// points as awarded for a win (FIDE Art. 7.1, WIN).
//
// This includes OTB wins, forfeit wins, and pairing-allocated byes (PAB).
// Awards are compared with the tournament's configured win points.
//
// FIDE Category B tiebreaker.
type Win struct{}

func (w *Win) ID() string   { return "win" }
func (w *Win) Name() string { return "Rounds Won" }

func (w *Win) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	opts := standard.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	winPoints := *opts.PointWin
	winCount := make(map[string]float64, len(scores))

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			var winner string
			points := *opts.PointWin
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				winner = game.WhiteID
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				winner = game.BlackID
			default:
				continue
			}
			if game.IsForfeit || game.Result.IsForfeit() {
				points = *opts.PointForfeitWin
			}
			if points == winPoints {
				winCount[winner]++
			}
		}

		for _, bye := range round.Byes {
			var points float64
			switch bye.Type {
			case chesspairing.ByePAB, chesspairing.ByeFullPoint:
				points = *opts.PointBye
			case chesspairing.ByeHalf:
				points = *opts.PointDraw
			case chesspairing.ByeZero:
				points = *opts.PointLoss
			case chesspairing.ByeAbsent:
				points = *opts.PointAbsent
			case chesspairing.ByeExcused:
				points = *opts.PointExcused
			case chesspairing.ByeClubCommitment:
				points = *opts.PointClubCommitment
			}
			if points == winPoints {
				winCount[bye.PlayerID]++
			}
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
