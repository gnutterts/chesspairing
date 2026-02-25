package swisslib

import "time"

// LookAheadFunc attempts to pair a bracket using only C1-C7 criteria.
// Returns true if a valid pairing exists (at least one pair), false otherwise.
// Used by C8 to check whether floaters allow the next bracket to be paired.
// The function must NOT call C8 recursively (no infinite recursion).
type LookAheadFunc func(bracket Bracket, ctx *CriteriaContext) bool

// CriteriaContext holds tournament-wide state needed to evaluate criteria.
type CriteriaContext struct {
	Players      map[string]*PlayerState
	TotalRounds  int
	CurrentRound int
	IsLastRound  bool
	TopScorers   map[string]bool // player IDs with >50% max score (final round only)

	// RemainingBrackets holds the brackets after the current one being paired.
	// Set by the orchestrator (dutch.go) before calling MatchBracketMulti.
	// Used by C8 to simulate whether floaters allow the next bracket to pair.
	RemainingBrackets []Bracket

	// LookAhead attempts to pair a bracket without C8 (to avoid infinite recursion).
	// Set by the orchestrator (dutch.go) as a closure wrapping MatchBracketMulti
	// with a criteria slice that has C8 set to nil.
	LookAhead LookAheadFunc

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

// SatisfiesAbsolute checks if ALL pairs in a candidate satisfy the absolute
// criteria C1 (no rematches) and C3 (no absolute color conflicts).
// Returns false if any pair violates an absolute criterion.
func SatisfiesAbsolute(cand *Candidate, ctx *CriteriaContext) bool {
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		if !C1NoRematches(pair, ctx) {
			return false
		}
		if !C3AbsoluteColorConflict(pair, ctx) {
			return false
		}
	}
	return true
}
