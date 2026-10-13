// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package football implements football-style scoring (3-1-0) for chess tournaments.
//
// Football scoring uses the same mechanism as standard scoring but with
// different default point values: 3 for a win, 1 for a draw, 0 for a loss.
// This is a popular alternative in informal club tournaments that rewards
// decisive results more heavily.
//
// All point values are configurable via Options, following the same
// pointer-nil pattern as standard scoring.
package football

import (
	"context"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// Scorer implements the chesspairing.Scorer interface for football scoring.
// It uses the standard scorer with football-specific defaults.
type Scorer struct {
	opts standard.Options
}

// New creates a new football scorer with the given options.
// Any nil option fields receive football defaults (3-1-0) rather than
// standard defaults (1-½-0).
func New(opts standard.Options) *Scorer {
	return &Scorer{opts: opts}
}

// NewFromMap creates a new football scorer from a map[string]any config.
func NewFromMap(m map[string]any) *Scorer {
	return New(standard.ParseOptions(m))
}

// Score calculates football-style scores for all active players.
func (s *Scorer) Score(ctx context.Context, state *chesspairing.TournamentState) ([]chesspairing.PlayerScore, error) {
	// Keep an unset bye value unset until Score can see the pairing system.
	resolved := withFootballDefaults(s.opts)
	resolved.PointBye = s.opts.PointBye
	return standard.New(resolved).Score(ctx, state)
}

// PointsForResult returns the points awarded for a specific game result
// using football scoring defaults. Without a tournament state, an unset PAB
// uses the win value (Double Swiss adds the draw only in Score).
func (s *Scorer) PointsForResult(result chesspairing.GameResult, rctx chesspairing.ResultContext) float64 {
	resolved := withFootballDefaults(s.opts)
	resolved.PointBye = s.opts.PointBye
	return standard.New(resolved).PointsForResult(result, rctx)
}

// withFootballDefaults fills in nil fields with football defaults.
func withFootballDefaults(opts standard.Options) standard.Options {
	if opts.PointWin == nil {
		opts.PointWin = float64Ptr(3.0)
	}
	if opts.PointDraw == nil {
		opts.PointDraw = float64Ptr(1.0)
	}
	if opts.PointLoss == nil {
		opts.PointLoss = float64Ptr(0.0)
	}
	if opts.PointBye == nil {
		opts.PointBye = float64Ptr(3.0)
	}
	if opts.PointForfeitWin == nil {
		opts.PointForfeitWin = float64Ptr(3.0)
	}
	if opts.PointForfeitLoss == nil {
		opts.PointForfeitLoss = float64Ptr(0.0)
	}
	if opts.PointAbsent == nil {
		opts.PointAbsent = float64Ptr(0.0)
	}
	return opts
}

func float64Ptr(v float64) *float64 { return &v }

// Ensure Scorer implements chesspairing.Scorer.
var _ chesspairing.Scorer = (*Scorer)(nil)
