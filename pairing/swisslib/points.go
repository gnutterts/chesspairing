// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// pairingPoints is the tournament's point scale for Swiss pairings.
type pairingPoints struct {
	win, draw, loss, forfeitWin, forfeitLoss float64
	pab, absent, excused, clubCommitment     float64
}

func pointsFor(cfg chesspairing.ScoringConfig) pairingPoints {
	opts := standard.Options{}
	switch cfg.System {
	// An empty system is standard scoring (chesspairing.ParseScoringSystem).
	case "", chesspairing.ScoringStandard, chesspairing.ScoringFootball:
		opts = standard.ParseOptions(cfg.Options)
		if cfg.System == chesspairing.ScoringFootball {
			// Mirror scoring/football.withFootballDefaults before applying standard
			// defaults to the remaining fields (excused and club commitment).
			if opts.PointWin == nil {
				opts.PointWin = chesspairing.Float64Ptr(3)
			}
			if opts.PointDraw == nil {
				opts.PointDraw = chesspairing.Float64Ptr(1)
			}
			if opts.PointForfeitWin == nil {
				opts.PointForfeitWin = chesspairing.Float64Ptr(3)
			}
		}
	}
	// C.04.1 Art. 4: without an explicit PAB value, award a win's points.
	if opts.PointBye == nil && opts.PointWin != nil {
		opts.PointBye = chesspairing.Float64Ptr(*opts.PointWin)
	}
	opts = opts.WithDefaults()
	return pairingPoints{
		win: *opts.PointWin, draw: *opts.PointDraw, loss: *opts.PointLoss,
		forfeitWin: *opts.PointForfeitWin, forfeitLoss: *opts.PointForfeitLoss,
		pab: *opts.PointBye, absent: *opts.PointAbsent,
		excused: *opts.PointExcused, clubCommitment: *opts.PointClubCommitment,
	}
}

// PairingPoints returns the tournament point policy used for Swiss pairings.
func PairingPoints(cfg chesspairing.ScoringConfig) pairingPoints { return pointsFor(cfg) }

// Win returns the points for a played win.
func (p pairingPoints) Win() float64 { return p.win }

// Draw returns the points for a draw.
func (p pairingPoints) Draw() float64 { return p.draw }

// Bye returns the points for a bye or absence.
func (p pairingPoints) Bye(bt chesspairing.ByeType) float64 { return p.bye(bt) }

// Game returns the pairing points for both sides of a game.
func (p pairingPoints) Game(game chesspairing.GameData) (float64, float64) {
	if game.Result == chesspairing.ResultDoubleForfeit {
		return 0, 0
	}
	if IsForfeit(game) {
		return p.unplayedGame(game.Result)
	}
	switch game.Result {
	case chesspairing.ResultWhiteWins:
		return p.win, p.loss
	case chesspairing.ResultBlackWins:
		return p.loss, p.win
	case chesspairing.ResultDraw, chesspairing.ResultPending:
		return p.draw, p.draw
	default:
		return 0, 0
	}
}

// unplayedGame returns the points awarded to each side of an unplayed game.
func (p pairingPoints) unplayedGame(result chesspairing.GameResult) (float64, float64) {
	switch result {
	case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
		return p.forfeitWin, p.forfeitLoss
	case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
		return p.forfeitLoss, p.forfeitWin
	case chesspairing.ResultDraw:
		return p.draw, p.draw
	default:
		return 0, 0
	}
}

// PairingWinPoints returns the configured points for a played win.
func PairingWinPoints(cfg chesspairing.ScoringConfig) float64 {
	return pointsFor(cfg).win
}

func (p pairingPoints) bye(bt chesspairing.ByeType) float64 {
	switch bt {
	case chesspairing.ByePAB, chesspairing.ByeFullPoint:
		return p.pab
	case chesspairing.ByeHalf:
		return p.draw
	case chesspairing.ByeZero:
		return p.loss
	case chesspairing.ByeAbsent:
		return p.absent
	case chesspairing.ByeExcused:
		return p.excused
	case chesspairing.ByeClubCommitment:
		return p.clubCommitment
	default:
		return 0
	}
}
