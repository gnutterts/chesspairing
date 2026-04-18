// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "time"

// CriteriaContext holds tournament-wide state needed to evaluate criteria.
type CriteriaContext struct {
	Players      map[string]*PlayerState
	TotalRounds  int
	CurrentRound int
	IsLastRound  bool
	TopScorers   map[string]bool // player IDs with >50% max score (final round only)

	// ForbiddenPairs contains canonicalized player ID pairs that must not be
	// paired together (e.g., players from the same club or family members).
	// Keys are [2]string with IDs in lexicographic order.
	// Enforced as an absolute criterion alongside C1 and C3.
	ForbiddenPairs map[[2]string]bool

	// Deadline is the time by which the pairing algorithm must complete.
	// When set, the matching algorithm returns the best result found so far
	// when the deadline is exceeded. Zero value means no deadline.
	Deadline time.Time
}

// DeadlineExceeded returns true if a deadline is set and has been exceeded.
func (ctx *CriteriaContext) DeadlineExceeded() bool {
	if ctx.Deadline.IsZero() {
		return false
	}
	return time.Now().After(ctx.Deadline)
}

// ProposedPairing is a candidate pairing being evaluated.
// BracketScore records the originating bracket's score for board ordering.
type ProposedPairing struct {
	White        *PlayerState
	Black        *PlayerState
	BracketScore float64 // native bracket score (for board ordering: higher bracket pairs first)
}

// BracketPairing is the complete pairing result for a bracket.
type BracketPairing struct {
	Pairs    []ProposedPairing
	Floaters []*PlayerState // players that couldn't be paired in this bracket
}

// --- Absolute criteria (C1, C3 --- never relaxed, filter functions) ---

// C1NoRematches returns true if the two players have NOT already played
// each other (forfeits excluded from opponent history).
func C1NoRematches(pair *ProposedPairing, ctx *CriteriaContext) bool {
	return !HasPlayed(pair.White, pair.Black)
}

// C2NoSecondPAB returns true if the player has not already received a bye.
// Used to validate bye candidates, not pair evaluation.
func C2NoSecondPAB(player *PlayerState, ctx *CriteriaContext) bool {
	return !player.ByeReceived
}

// C3AbsoluteColorConflict returns true if the pairing does NOT create an
// absolute color conflict.
//
// FIDE C.04.3 (Feb 2026) Article 2.1.3 [C3]:
// "Non-topscorers with the same absolute colour preference shall not meet."
//
// This means C3 only applies when BOTH players are non-topscorers.
// If EITHER player is a topscorer (Article 1.8: score > 50% of max possible
// in the final round), C3 does not apply.
func C3AbsoluteColorConflict(pair *ProposedPairing, ctx *CriteriaContext) bool {
	prefW := ComputeColorPreference(pair.White.ColorHistory)
	prefB := ComputeColorPreference(pair.Black.ColorHistory)

	// If neither has absolute preference, C3 passes.
	if !prefW.AbsolutePreference || !prefB.AbsolutePreference {
		return true
	}

	// Both have absolute --- check if same color.
	if prefW.Color != nil && prefB.Color != nil && *prefW.Color != *prefB.Color {
		return true // opposite absolutes, fine
	}

	// Same absolute color --- C3 does not apply if either is a topscorer.
	if ctx.IsLastRound && (ctx.TopScorers[pair.White.ID] || ctx.TopScorers[pair.Black.ID]) {
		return true // topscorer exception per FIDE C3
	}

	return false // conflict: both are non-topscorers with same absolute preference
}

// C4CompleteBracket checks that all players in a bracket are accounted for
// as either paired or floaters. Structural validation.
func C4CompleteBracket(bp *BracketPairing, playerCount int) bool {
	return len(bp.Pairs)*2+len(bp.Floaters) == playerCount
}

// CanonicalPairKey returns a canonicalized key for a pair of player IDs.
// The IDs are sorted lexicographically so that (a,b) and (b,a) produce the same key.
func CanonicalPairKey(a, b string) [2]string {
	if a <= b {
		return [2]string{a, b}
	}
	return [2]string{b, a}
}

// IsForbiddenPair returns true if the two players are in the forbidden pairs list.
// This is an absolute criterion: forbidden pairs must never be matched.
func IsForbiddenPair(pair *ProposedPairing, ctx *CriteriaContext) bool {
	if len(ctx.ForbiddenPairs) == 0 {
		return false
	}
	key := CanonicalPairKey(pair.White.ID, pair.Black.ID)
	return ctx.ForbiddenPairs[key]
}

// IsPairForbiddenByID checks if two player IDs are in the forbidden pairs list.
// Convenience function for edge-generation code that works with player IDs directly.
func IsPairForbiddenByID(aID, bID string, ctx *CriteriaContext) bool {
	if len(ctx.ForbiddenPairs) == 0 {
		return false
	}
	return ctx.ForbiddenPairs[CanonicalPairKey(aID, bID)]
}
