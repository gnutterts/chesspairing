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
//
// # Bye types and absences
//
// Six ByeType values cover the unplayed-round cases. They differ in
// scoring weight, in whether they count as a played round for
// tiebreakers, in pairing impact, and in TRF representation. The
// matrix below summarises the semantics; specifics for the standard
// scorer live on its Options fields, and Keizer scoring's bye and
// absent values are valuation-relative rather than fixed.
//
//	ByeType            Standard pts (default)   Counts as played   PAB-tracked   TRF code
//	-----------------  -----------------------  -----------------  ------------  --------
//	ByePAB             PointBye (1.0)           yes                yes           F
//	ByeHalf            PointDraw (0.5)          yes                no            H
//	ByeZero            PointLoss (0.0)          yes                no            Z
//	ByeAbsent          PointAbsent (0.0)        no                 no            U
//	ByeExcused         PointExcused (0.0)       no                 no            (directive)
//	ByeClubCommitment  PointClubCommitment (0)  no                 no            (directive)
//
// "Counts as played" affects rounds-played tiebreakers and Buchholz
// virtual-opponent calculations. "PAB-tracked" matters for the Swiss
// pairers' constraint that no player gets the pairing-allocated bye
// twice. "TRF code" is the round-column letter; ByeExcused and
// ByeClubCommitment have no TRF round-column representation and are
// carried in chesspairing directive comments instead (see the trf
// sub-package).
//
// Pre-assigned byes are configured via TournamentState.PreAssignedByes.
// The Swiss pairers honour them by partitioning the player pool before
// the matching step, so a pre-assigned bye of any type passes through
// to the PairingResult unchanged. The PAB-uniqueness constraint only
// applies to algorithmically allocated byes.
//
// Player withdrawals use PlayerEntry.WithdrawnAfterRound (a *int).
// state.IsActiveInRound(id, n) and state.ActivePlayerIDs(n) are the
// canonical accessors. Tiebreakers consult the active filter
// contemporaneously per historical round, so a player withdrawn after
// round 3 still contributes to opponents' Buchholz for rounds 1 and 2.
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
