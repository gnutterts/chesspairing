// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

func init() {
	Register("progressive", func() chesspairing.TieBreaker { return &Progressive{} })
	Register("progressive-cut1", func() chesspairing.TieBreaker { return &Progressive{cut1: true} })
}

// Progressive computes the progressive score tiebreaker.
//
// The progressive score (also called cumulative score) is the sum of
// cumulative round-by-round scores. A player who scores well in early
// rounds accumulates a higher progressive score than one who scores
// the same total but in later rounds.
//
// Example: a player scoring 1, 0, 1, 1 has cumulative scores
// [1, 1, 2, 3] and progressive = 1 + 1 + 2 + 3 = 7.
type Progressive struct {
	cut1 bool
}

func (p *Progressive) ID() string {
	if p.cut1 {
		return "progressive-cut1"
	}
	return "progressive"
}

func (p *Progressive) Name() string {
	if p.cut1 {
		return "Progressive Score Cut-1"
	}
	return "Progressive Score"
}

func (p *Progressive) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	// Build round-by-round scores for each player.
	roundScores := buildRoundScores(state)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var progressive float64
		var cumulative float64
		for roundIndex, roundScore := range roundScores[ps.PlayerID] {
			cumulative += roundScore
			if p.cut1 && roundIndex == 0 {
				continue
			}
			progressive += cumulative
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    progressive,
		}
	}
	return result, nil
}

// buildRoundScores returns a per-player, per-round score array.
// Each entry is the points awarded in that round under the tournament scoring options.
func buildRoundScores(state *chesspairing.TournamentState) map[string][]float64 {
	scores := make(map[string][]float64)
	opts := standard.ParseOptions(state.ScoringConfig.Options).WithDefaults()

	// Initialize a slot for every player; rounds outside the active window
	// will simply remain at zero.
	for _, p := range state.Players {
		scores[p.ID] = make([]float64, len(state.Rounds))
	}

	for roundIdx, round := range state.Rounds {
		for _, game := range round.Games {
			var white, black float64
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				white, black = *opts.PointWin, *opts.PointLoss
				if game.IsForfeit || game.Result.IsForfeit() {
					white, black = *opts.PointForfeitWin, *opts.PointForfeitLoss
				}
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				white, black = *opts.PointLoss, *opts.PointWin
				if game.IsForfeit || game.Result.IsForfeit() {
					white, black = *opts.PointForfeitLoss, *opts.PointForfeitWin
				}
			case chesspairing.ResultDraw:
				white, black = *opts.PointDraw, *opts.PointDraw
			case chesspairing.ResultDoubleForfeit:
			default:
				continue
			}
			if _, ok := scores[game.WhiteID]; ok {
				scores[game.WhiteID][roundIdx] += white
			}
			if _, ok := scores[game.BlackID]; ok {
				scores[game.BlackID][roundIdx] += black
			}
		}

		for _, bye := range round.Byes {
			if _, ok := scores[bye.PlayerID]; ok {
				switch bye.Type {
				case chesspairing.ByePAB, chesspairing.ByeFullPoint:
					scores[bye.PlayerID][roundIdx] = *opts.PointBye
				case chesspairing.ByeHalf:
					scores[bye.PlayerID][roundIdx] = *opts.PointDraw
				case chesspairing.ByeZero:
					scores[bye.PlayerID][roundIdx] = *opts.PointLoss
				case chesspairing.ByeAbsent:
					scores[bye.PlayerID][roundIdx] = *opts.PointAbsent
				case chesspairing.ByeExcused:
					scores[bye.PlayerID][roundIdx] = *opts.PointExcused
				case chesspairing.ByeClubCommitment:
					scores[bye.PlayerID][roundIdx] = *opts.PointClubCommitment
				}
			}
		}

		// Absent players get 0 for the round (already initialized to 0).
	}

	return scores
}
