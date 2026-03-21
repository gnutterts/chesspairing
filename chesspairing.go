// Package chesspairing provides chess tournament pairing, scoring, and
// tiebreaking engines in pure Go. It implements FIDE-approved Swiss pairing
// systems (Dutch C.04.3, Burstein C.04.4.2, Dubov C.04.4.1, Lim C.04.4.3,
// and Double-Swiss C.04.5), Keizer pairing, and round-robin pairing, along
// with standard, Keizer, and football scoring systems and 25 tiebreaker
// algorithms.
//
// Engines operate on in-memory data structures (TournamentState, PlayerEntry,
// RoundData) and have no I/O, database, or network dependencies. They are
// safe for concurrent use when each goroutine supplies its own TournamentState.
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
