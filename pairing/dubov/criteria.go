// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// NumDubovViolations is the number of quality criteria (C4-C10).
const NumDubovViolations = 7

// Violation index constants for DubovCandidateScore.Violations.
const (
	IdxC4  = 0 // upfloater count
	IdxC5  = 1 // upfloater score sum (negated: higher = better -> stored as deficit)
	IdxC6  = 2 // colour preference violations
	IdxC7  = 3 // MaxT upfloater violations
	IdxC8  = 4 // consecutive-round upfloaters
	IdxC9  = 5 // MaxT upfloater-opponent violations
	IdxC10 = 6 // consecutive-round MaxT violations
)

// MaxT returns the maximum number of times a player may be upfloated.
// Per Dubov Art. 1.8: MaxT = 2 + floor(Rnds/5).
func MaxT(completedRounds int) int {
	return 2 + completedRounds/5
}

// UpfloatCount returns how many times a player has been upfloated
// across all completed rounds.
func UpfloatCount(p *swisslib.PlayerState) int {
	count := 0
	for _, f := range p.FloatHistory {
		if f == swisslib.FloatUp {
			count++
		}
	}
	return count
}

// --- Absolute criteria (C1-C3) ---

// C1NoRematches returns true if the two players have NOT already played each other.
func C1NoRematches(a, b *swisslib.PlayerState) bool {
	return !swisslib.HasPlayed(a, b)
}

// C3NoAbsoluteColorConflict returns true if the two players do NOT both
// have the same absolute colour preference. If either player has no
// absolute preference, there is no conflict.
func C3NoAbsoluteColorConflict(a, b *swisslib.PlayerState) bool {
	prefA := swisslib.ComputeColorPreference(a.ColorHistory)
	prefB := swisslib.ComputeColorPreference(b.ColorHistory)

	if !prefA.AbsolutePreference || !prefB.AbsolutePreference {
		return true // no conflict if either is non-absolute
	}

	// Both absolute - conflict only if same colour.
	if prefA.Color == nil || prefB.Color == nil {
		return true
	}
	return *prefA.Color != *prefB.Color
}

// SatisfiesAbsoluteCriteria checks C1, C3, and forbidden pairs for a proposed pair.
func SatisfiesAbsoluteCriteria(a, b *swisslib.PlayerState, forbidden map[[2]string]bool) bool {
	if !C1NoRematches(a, b) {
		return false
	}
	if !C3NoAbsoluteColorConflict(a, b) {
		return false
	}
	if len(forbidden) > 0 {
		key := swisslib.CanonicalPairKey(a.ID, b.ID)
		if forbidden[key] {
			return false
		}
	}
	return true
}

// --- Candidate scoring ---

// DubovCandidateScore holds the quality evaluation of a complete bracket pairing.
type DubovCandidateScore struct {
	UpfloaterCount    int                     // C4: number of upfloaters (lower = better)
	UpfloaterScoreSum float64                 // C5: sum of upfloater scores (higher = better)
	Violations        [NumDubovViolations]int // C4-C10 violation counts
	TranspositionIdx  int                     // Transposition sequence number (lower = better)
}

// Compare returns -1 if s is better than other, +1 if worse, 0 if equal.
// Comparison order per FIDE C.04.4.1:
//  1. C4: fewer upfloaters is better
//  2. C5: higher upfloater score sum is better
//  3. C6-C10: fewer violations is better (lexicographic)
//  4. Lower transposition index is better
func (s *DubovCandidateScore) Compare(other *DubovCandidateScore) int {
	// C4: fewer upfloaters.
	if s.UpfloaterCount != other.UpfloaterCount {
		if s.UpfloaterCount < other.UpfloaterCount {
			return -1
		}
		return 1
	}

	// C5: higher upfloater score sum (reversed: more is better).
	if s.UpfloaterScoreSum != other.UpfloaterScoreSum {
		if s.UpfloaterScoreSum > other.UpfloaterScoreSum {
			return -1
		}
		return 1
	}

	// C6-C10: lexicographic on violations.
	for i := IdxC6; i < NumDubovViolations; i++ {
		if s.Violations[i] != other.Violations[i] {
			if s.Violations[i] < other.Violations[i] {
				return -1
			}
			return 1
		}
	}

	// Transposition index tiebreak.
	if s.TranspositionIdx != other.TranspositionIdx {
		if s.TranspositionIdx < other.TranspositionIdx {
			return -1
		}
		return 1
	}

	return 0
}

// IsPerfect returns true if there are no upfloaters and no violations.
func (s *DubovCandidateScore) IsPerfect() bool {
	if s.UpfloaterCount != 0 {
		return false
	}
	for _, v := range s.Violations {
		if v != 0 {
			return false
		}
	}
	return true
}

// --- Quality criterion functions ---

// CriterionC4 returns the number of upfloaters in the candidate.
func CriterionC4(floaters []*swisslib.PlayerState) int {
	return len(floaters)
}

// CriterionC5 returns the sum of upfloater scores.
// Higher is better, so callers should compare reversed.
func CriterionC5(floaters []*swisslib.PlayerState) float64 {
	var sum float64
	for _, f := range floaters {
		sum += f.Score
	}
	return sum
}

// CriterionC6 returns the number of players not receiving their colour preference.
func CriterionC6(pairs []proposedPairing) int {
	count := 0
	for _, pair := range pairs {
		prefW := swisslib.ComputeColorPreference(pair.white.ColorHistory)
		prefB := swisslib.ComputeColorPreference(pair.black.ColorHistory)

		// White player would prefer black?
		if prefW.Color != nil && *prefW.Color == swisslib.ColorBlack {
			count++
		}
		// Black player would prefer white?
		if prefB.Color != nil && *prefB.Color == swisslib.ColorWhite {
			count++
		}
	}
	return count
}

// CriterionC7 returns the number of upfloaters who have reached or exceeded MaxT.
func CriterionC7(floaters []*swisslib.PlayerState, maxT int) int {
	count := 0
	for _, f := range floaters {
		if UpfloatCount(f) >= maxT {
			count++
		}
	}
	return count
}

// CriterionC8 returns the number of upfloaters who also upfloated in the previous round.
func CriterionC8(floaters []*swisslib.PlayerState) int {
	count := 0
	for _, f := range floaters {
		if len(f.FloatHistory) > 0 && swisslib.LastFloat(f.FloatHistory) == swisslib.FloatUp {
			count++
		}
	}
	return count
}

// CriterionC9 returns the number of upfloater opponents who are at/above MaxT.
func CriterionC9(pairs []proposedPairing, floaterIDs map[string]bool, maxT int) int {
	count := 0
	for _, pair := range pairs {
		// Check if white is a floater -> check black opponent.
		if floaterIDs[pair.white.ID] {
			if UpfloatCount(pair.black) >= maxT {
				count++
			}
		}
		// Check if black is a floater -> check white opponent.
		if floaterIDs[pair.black.ID] {
			if UpfloatCount(pair.white) >= maxT {
				count++
			}
		}
	}
	return count
}

// CriterionC10 returns the number of consecutive-round upfloaters at/above MaxT.
func CriterionC10(floaters []*swisslib.PlayerState, maxT int) int {
	count := 0
	for _, f := range floaters {
		if UpfloatCount(f) >= maxT && len(f.FloatHistory) > 0 && swisslib.LastFloat(f.FloatHistory) == swisslib.FloatUp {
			count++
		}
	}
	return count
}

// proposedPairing is a Dubov-internal pair representation before colour allocation.
type proposedPairing struct {
	white        *swisslib.PlayerState // tentatively assigned white
	black        *swisslib.PlayerState // tentatively assigned black
	bracketScore float64
}
