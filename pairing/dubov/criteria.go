// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// MaxT returns Article 1.8.2's maximum number of upfloats for a tournament
// with totalRounds planned rounds.
func MaxT(totalRounds int) int { return 2 + totalRounds/5 }

// UpfloatCount returns the number of prior upfloats.
func UpfloatCount(player *swisslib.PlayerState) int {
	count := 0
	for _, float := range player.FloatHistory {
		if float == swisslib.FloatUp {
			count++
		}
	}
	return count
}

// C1NoRematches implements C1.
func C1NoRematches(a, b *swisslib.PlayerState) bool { return !swisslib.HasPlayed(a, b) }

// C3NoAbsoluteColorConflict implements C3.
func C3NoAbsoluteColorConflict(a, b *swisslib.PlayerState) bool {
	preferenceA := preference(a)
	preferenceB := preference(b)
	if !preferenceA.AbsolutePreference || !preferenceB.AbsolutePreference || preferenceA.Color == nil || preferenceB.Color == nil {
		return true
	}
	return *preferenceA.Color != *preferenceB.Color
}

// SatisfiesAbsoluteCriteria implements C1 and C3, plus explicitly forbidden
// pairs supplied by the tournament.
func SatisfiesAbsoluteCriteria(a, b *swisslib.PlayerState, forbidden map[[2]string]bool) bool {
	if !C1NoRematches(a, b) || !C3NoAbsoluteColorConflict(a, b) {
		return false
	}
	return !forbidden[swisslib.CanonicalPairKey(a.ID, b.ID)]
}

// preference implements Article 1.6.4 locally: an unplayed player mildly
// prefers Black. Other pairing systems continue to use swisslib unchanged.
func preference(player *swisslib.PlayerState) swisslib.ColorPreference {
	result := swisslib.ComputeColorPreference(player.ColorHistory)
	if len(result.PlayedColors) == 0 {
		black := swisslib.ColorBlack
		result.Color = &black
	}
	return result
}

// NumDubovViolations and DubovCandidateScore are retained for compatibility
// with callers of the former criterion inspector. Pairing uses candidateQuality.
const NumDubovViolations = 7

type DubovCandidateScore struct {
	UpfloaterCount    int
	UpfloaterScoreSum float64
	Violations        [NumDubovViolations]int
}

func (s *DubovCandidateScore) Compare(other *DubovCandidateScore) int {
	if s.UpfloaterCount < other.UpfloaterCount {
		return -1
	}
	if s.UpfloaterCount > other.UpfloaterCount {
		return 1
	}
	if s.UpfloaterScoreSum > other.UpfloaterScoreSum {
		return -1
	}
	if s.UpfloaterScoreSum < other.UpfloaterScoreSum {
		return 1
	}
	for i := range s.Violations {
		if s.Violations[i] < other.Violations[i] {
			return -1
		}
		if s.Violations[i] > other.Violations[i] {
			return 1
		}
	}
	return 0
}

func (s *DubovCandidateScore) IsPerfect() bool {
	if s.UpfloaterCount != 0 {
		return false
	}
	for _, violation := range s.Violations {
		if violation != 0 {
			return false
		}
	}
	return true
}

// CriterionC7 counts players who do not get their colour preference
// (Article 2.3.3).
func CriterionC7(pairs []proposedPairing) int {
	return qualityFor(pairs, nil, true, 0).colorViolations
}

// CriterionC8 counts maximum upfloaters (Article 2.3.4).
func CriterionC8(players []*swisslib.PlayerState, maxT int) int {
	count := 0
	for _, player := range players {
		if UpfloatCount(player) >= maxT {
			count++
		}
	}
	return count
}

// CriterionC10 counts players who upfloated in the previous round
// (Article 2.3.6).
func CriterionC10(players []*swisslib.PlayerState) int {
	count := 0
	for _, player := range players {
		if len(player.FloatHistory) > 0 && swisslib.LastFloat(player.FloatHistory) == swisslib.FloatUp {
			count++
		}
	}
	return count
}

// proposedPairing is a Dubov-internal pair representation before colour allocation.
type proposedPairing struct {
	white        *swisslib.PlayerState
	black        *swisslib.PlayerState
	bracketScore float64
}

// candidateQuality is C5-C10 for a fully paired bracket. C6 is represented by
// its required descending vector of score differences.
type candidateQuality struct {
	upfloaters         int
	differences        []float64
	colorViolations    int
	maximumUpfloaters  int
	maximumTimes       int
	previousUpfloaters int
}

func qualityFor(pairs []proposedPairing, upfloaterIDs map[string]bool, lastRound bool, maxT int) candidateQuality {
	quality := candidateQuality{upfloaters: len(upfloaterIDs)}
	for _, pair := range pairs {
		if upfloaterIDs[pair.white.ID] || upfloaterIDs[pair.black.ID] {
			difference := pair.white.Score - pair.black.Score
			if difference < 0 {
				difference = -difference
			}
			quality.differences = append(quality.differences, difference)
		}
		whitePreference := preference(pair.white)
		blackPreference := preference(pair.black)
		if whitePreference.Color != nil && *whitePreference.Color == swisslib.ColorBlack {
			quality.colorViolations++
		}
		if blackPreference.Color != nil && *blackPreference.Color == swisslib.ColorWhite {
			quality.colorViolations++
		}
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(quality.differences)))
	if lastRound {
		return quality
	}
	for _, pair := range pairs {
		for _, player := range []*swisslib.PlayerState{pair.white, pair.black} {
			if !upfloaterIDs[player.ID] {
				continue
			}
			count := UpfloatCount(player)
			if count >= maxT {
				quality.maximumUpfloaters++
				quality.maximumTimes += count
			}
			if len(player.FloatHistory) > 0 && swisslib.LastFloat(player.FloatHistory) == swisslib.FloatUp {
				quality.previousUpfloaters++
			}
		}
	}
	return quality
}

// betterQuality reports whether left is better under the lexicographic C5-C10
// order. Equal qualities deliberately retain the earlier Article 4.1 set.
func betterQuality(left, right candidateQuality) bool {
	if left.upfloaters != right.upfloaters {
		return left.upfloaters < right.upfloaters
	}
	// C6 compares the descending difference vectors. A shorter vector has
	// fewer pairs involving upfloaters (fewer upfloater-upfloater pairs), which
	// is preferred once all shared differences are equal.
	n := len(left.differences)
	if len(right.differences) < n {
		n = len(right.differences)
	}
	for i := 0; i < n; i++ {
		if left.differences[i] != right.differences[i] {
			return left.differences[i] < right.differences[i]
		}
	}
	if len(left.differences) != len(right.differences) {
		return len(left.differences) < len(right.differences)
	}
	if left.colorViolations != right.colorViolations {
		return left.colorViolations < right.colorViolations
	}
	if left.maximumUpfloaters != right.maximumUpfloaters {
		return left.maximumUpfloaters < right.maximumUpfloaters
	}
	if left.maximumTimes != right.maximumTimes {
		return left.maximumTimes < right.maximumTimes
	}
	return left.previousUpfloaters < right.previousUpfloaters
}
