// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("standard-points", func() chesspairing.TieBreaker { return &StandardPoints{} })
}

// StandardPoints computes the standard points tiebreaker (FIDE Art. 7.7, STD).
//
// For each round, the player gets:
//   - 1 if they scored more points than their opponent
//   - 0.5 if they scored the same
//   - 0 if they scored fewer
//
// For unplayed rounds (byes/absences), the awarded points are compared to 0.5
// (the draw value): PAB(1.0)→1, half-bye(0.5)→0.5, zero-bye/absent(0)→0.
//
// FIDE Category B tiebreaker.
type StandardPoints struct{}

func (sp *StandardPoints) ID() string   { return "standard-points" }
func (sp *StandardPoints) Name() string { return "Standard Points" }

// roundResult holds the points awarded to a player in one round, and those
// of their scheduled opponent when the round was played.
type roundResult struct {
	points    float64
	hasOpp    bool
	oppPoints float64
}

// addStandardRoundPoints awards one game's points in the player's round
// slot. A Double Swiss round holds two games against the same scheduled
// opponent, so their points are added before the comparison of Article 7.7.
func addStandardRoundPoints(playerRounds map[string][]roundResult, roundIdx int, playerID string, points, oppPoints float64) {
	slot, ok := playerRounds[playerID]
	if !ok {
		return
	}
	rr := slot[roundIdx]
	if rr.hasOpp {
		rr.points += points
		rr.oppPoints += oppPoints
	} else {
		rr = roundResult{points: points, hasOpp: true, oppPoints: oppPoints}
	}
	slot[roundIdx] = rr
}

func (sp *StandardPoints) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	opts := tournamentPoints(state)
	drawValue := *opts.PointDraw

	// Build per-round, per-player points awarded.
	playerRounds := make(map[string][]roundResult)

	// Initialize a slot for every player. Inactive-in-this-round players
	// will simply have empty roundResult entries for those rounds.
	for _, p := range state.Players {
		playerRounds[p.ID] = make([]roundResult, len(state.Rounds))
	}

	for roundIdx, round := range state.Rounds {
		played := make(map[string]bool)

		for _, game := range round.Games {
			var whitePoints, blackPoints float64

			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				whitePoints, blackPoints = *opts.PointWin, *opts.PointLoss
				if game.IsForfeit || game.Result.IsForfeit() {
					whitePoints, blackPoints = *opts.PointForfeitWin, *opts.PointForfeitLoss
				}
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				whitePoints, blackPoints = *opts.PointLoss, *opts.PointWin
				if game.IsForfeit || game.Result.IsForfeit() {
					whitePoints, blackPoints = *opts.PointForfeitLoss, *opts.PointForfeitWin
				}
			case chesspairing.ResultDraw:
				whitePoints, blackPoints = *opts.PointDraw, *opts.PointDraw
			case chesspairing.ResultDoubleForfeit:
			default:
				continue
			}

			addStandardRoundPoints(playerRounds, roundIdx, game.WhiteID, whitePoints, blackPoints)
			addStandardRoundPoints(playerRounds, roundIdx, game.BlackID, blackPoints, whitePoints)
			played[game.WhiteID] = true
			played[game.BlackID] = true
		}

		for _, bye := range round.Byes {
			if _, ok := playerRounds[bye.PlayerID]; !ok {
				continue
			}
			var pts float64
			switch bye.Type {
			case chesspairing.ByePAB, chesspairing.ByeFullPoint:
				pts = *opts.PointBye
			case chesspairing.ByeHalf:
				pts = *opts.PointDraw
			case chesspairing.ByeZero:
				pts = *opts.PointLoss
			case chesspairing.ByeAbsent:
				pts = *opts.PointAbsent
			case chesspairing.ByeExcused:
				pts = *opts.PointExcused
			case chesspairing.ByeClubCommitment:
				pts = *opts.PointClubCommitment
			}
			playerRounds[bye.PlayerID][roundIdx] = roundResult{
				points: pts, hasOpp: false,
			}
			played[bye.PlayerID] = true
		}

		// Absent players: 0 points, no opponent (already zero-initialized).
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var std float64
		for _, rr := range playerRounds[ps.PlayerID] {
			if rr.hasOpp {
				// Compare with opponent.
				if rr.points > rr.oppPoints {
					std += 1.0
				} else if rr.points == rr.oppPoints {
					std += 0.5
				}
			} else {
				// Unplayed round: compare awarded points to draw value.
				if rr.points > drawValue {
					std += 1.0
				} else if rr.points == drawValue {
					std += 0.5
				}
			}
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    std,
		}
	}
	return result, nil
}
