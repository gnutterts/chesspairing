// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package chesspairing provides chess tournament pairing, scoring, and
// tiebreaking engines in pure Go. It implements FIDE-approved Swiss pairing
// systems (Dutch C.04.3, Burstein C.04.4.2, Dubov C.04.4.1, Lim C.04.4.3,
// Double-Swiss C.04.5, and Team Swiss C.04.6), Keizer pairing, and
// round-robin pairing, along with standard, Keizer, and football scoring
// systems and 25 tiebreaker algorithms.
//
// Engines operate on in-memory data structures (TournamentState, PlayerEntry,
// RoundData) and have no I/O, database, or network dependencies. They are
// safe for concurrent use when each goroutine supplies its own TournamentState.
//
// Context: all engine interface methods accept context.Context as their
// first parameter for API compatibility with service layers. However,
// since all computation is CPU-bound and in-memory (no I/O, no network),
// the context is not currently checked for cancellation. Callers should
// still pass a context for forward compatibility.
//
// # Forfeit handling across subsystems
//
// A single FIDE-aligned semantics for forfeits doesn't exist: the rule
// depends on the question being asked. Subsystems in this module make
// different choices, all consistent with the FIDE handbook:
//
//	Subsystem            Single forfeit (1-0f / 0-1f)     Double forfeit (0-0f)
//	-----------------    ------------------------------   --------------------------
//	Scorer               Awards PointForfeitWin/Loss      Awards 0 to both
//	TieBreaker           Excluded from opponent data      Excluded from opponent data
//	PlayedPairs          Excluded by default              Always excluded
//	standings.Build      Counts as +1 win or +1 loss      0 across the board
//
// The PlayedPairs default (excluding single forfeits) matches FIDE's
// position that a forfeit didn't really happen as a chess game and
// therefore the players may meet again. Setting HistoryOptions.IncludeForfeits
// to true crosses into house-rule territory.
package chesspairing

import "context"

// Pairer generates pairings for a round given tournament state.
type Pairer interface {
	Pair(ctx context.Context, state *TournamentState) (*PairingResult, error)
}

// Scorer calculates standings from game results.
type Scorer interface {
	Score(ctx context.Context, state *TournamentState) ([]PlayerScore, error)
	PointsForResult(result GameResult, rctx ResultContext) float64
}

// TieBreaker computes a single tiebreak value for each player.
type TieBreaker interface {
	ID() string
	Name() string
	Compute(ctx context.Context, state *TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
