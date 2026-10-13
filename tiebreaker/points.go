// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// tournamentPoints resolves the tournament's scoring scale before tie-break calculations.
func tournamentPoints(state *chesspairing.TournamentState) standard.Options {
	var opts standard.Options
	switch state.ScoringConfig.System {
	case "", chesspairing.ScoringStandard, chesspairing.ScoringFootball:
		opts = standard.ParseOptions(state.ScoringConfig.Options)
		if state.ScoringConfig.System == chesspairing.ScoringFootball {
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
	// An unset bye follows the win value, as in the standings scorer; a
	// Double Swiss pairing-allocated bye is worth a win plus a draw (C.04.5 1.4).
	if opts.PointBye == nil && opts.PointWin != nil {
		bye := *opts.PointWin
		if state.PairingConfig.System == chesspairing.PairingDoubleSwiss {
			draw := 0.5
			if opts.PointDraw != nil {
				draw = *opts.PointDraw
			}
			bye += draw
		}
		opts.PointBye = chesspairing.Float64Ptr(bye)
	}
	return opts.WithDefaults()
}
