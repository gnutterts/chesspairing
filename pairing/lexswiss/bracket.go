// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import "sort"

// CriteriaFunc is a function that checks whether a proposed pair satisfies
// system-specific criteria. Returns true if the pair is acceptable.
//
// The function is called for each candidate pair during the lexicographic
// enumeration. If it returns false, the pair is skipped and the next
// candidate is tried.
//
// Double-Swiss uses this for C8 (colour preferences).
// Team Swiss uses this for C8-C10 (colour preferences).
type CriteriaFunc func(a, b *ParticipantState) bool

// PairBracket pairs all participants in a bracket using the lexicographic
// algorithm described in Art. 3.6 (shared by Double-Swiss and Team Swiss).
//
// The algorithm enumerates all legal pairings in lexicographic order and
// selects the first one satisfying all criteria. Lexicographic order means:
// the participant with the lowest TPN is paired with the lowest-TPN
// available partner first. If that leads to a dead end (remaining participants
// can't all be paired), the algorithm backtracks and tries the next partner.
//
// Absolute criteria enforced by PairBracket:
//   - C1: No two participants play each other more than once
//   - Forbidden pairs are not paired
//
// Additional criteria are checked via the CriteriaFunc parameter.
// If criteriaFn is nil, only C1 and forbidden pairs are checked.
//
// Parameters:
//   - participants: sorted by TPN ascending
//   - forbidden: forbidden pairs map (nil if none)
//   - criteriaFn: additional criteria function (nil = no extra criteria)
//
// Returns the list of pairs (each pair is [lower-TPN, higher-TPN]).
// If no complete pairing is possible, returns the best partial pairing
// (as many pairs as possible in lexicographic order).
func PairBracket(participants []*ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc) [][2]*ParticipantState {
	n := len(participants)
	if n < 2 {
		return nil
	}

	// Sort by TPN ascending.
	sorted := make([]*ParticipantState, n)
	copy(sorted, participants)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].TPN < sorted[j].TPN
	})

	// Try to find a complete pairing using DFS with backtracking.
	used := make([]bool, n)
	pairs := make([][2]*ParticipantState, 0, n/2)

	if pairRecursive(sorted, used, &pairs, forbidden, criteriaFn) {
		return pairs
	}

	// No complete pairing found. Return best partial pairing.
	// Reset and do a greedy partial match.
	pairs = pairs[:0]
	for i := range used {
		used[i] = false
	}
	greedyPartialPair(sorted, used, &pairs, forbidden, criteriaFn)
	return pairs
}

// pairRecursive attempts to find a complete pairing using DFS.
// Returns true if a complete pairing is found.
func pairRecursive(participants []*ParticipantState, used []bool, pairs *[][2]*ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc) bool {
	n := len(participants)

	// Find the first unused participant.
	firstUnused := -1
	for i := 0; i < n; i++ {
		if !used[i] {
			firstUnused = i
			break
		}
	}

	// If no unused participant, we're done (or only 1 left for odd count).
	if firstUnused == -1 {
		return true
	}

	// Count remaining unused participants.
	remaining := 0
	for i := firstUnused; i < n; i++ {
		if !used[i] {
			remaining++
		}
	}

	// If only 1 unused participant remains (odd count), consider it complete.
	if remaining == 1 {
		return true
	}

	// Try pairing firstUnused with each subsequent unused participant
	// in lexicographic order (ascending TPN).
	used[firstUnused] = true
	for j := firstUnused + 1; j < n; j++ {
		if used[j] {
			continue
		}

		a, b := participants[firstUnused], participants[j]

		// C1: No repeat pairings.
		if HasPlayed(a, b) {
			continue
		}

		// Forbidden pairs.
		if isForbidden(a.ID, b.ID, forbidden) {
			continue
		}

		// System-specific criteria.
		if criteriaFn != nil && !criteriaFn(a, b) {
			continue
		}

		// Try this pairing.
		used[j] = true
		*pairs = append(*pairs, [2]*ParticipantState{a, b})

		if pairRecursive(participants, used, pairs, forbidden, criteriaFn) {
			return true
		}

		// Backtrack.
		used[j] = false
		*pairs = (*pairs)[:len(*pairs)-1]
	}

	used[firstUnused] = false
	return false
}

// greedyPartialPair pairs as many participants as possible using a greedy
// approach when no complete pairing exists.
func greedyPartialPair(participants []*ParticipantState, used []bool, pairs *[][2]*ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc) {
	n := len(participants)
	for i := 0; i < n; i++ {
		if used[i] {
			continue
		}
		for j := i + 1; j < n; j++ {
			if used[j] {
				continue
			}

			a, b := participants[i], participants[j]
			if HasPlayed(a, b) {
				continue
			}
			if isForbidden(a.ID, b.ID, forbidden) {
				continue
			}
			if criteriaFn != nil && !criteriaFn(a, b) {
				continue
			}

			used[i] = true
			used[j] = true
			*pairs = append(*pairs, [2]*ParticipantState{a, b})
			break
		}
	}
}
