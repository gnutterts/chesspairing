// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"context"
	"errors"
	"sort"
)

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

// ErrNoCompletePairing is returned when a bracket cannot be paired completely.
var ErrNoCompletePairing = errors.New("no complete pairing exists for bracket")

// PairBracket pairs all participants in a bracket using the lexicographic
// algorithm described in Art. 3.6 (shared by Double-Swiss and Team Swiss).
//
// The algorithm enumerates all legal pairings in lexicographic order and
// selects the first one satisfying all criteria. Pairings are ordered by their
// Article 3.6 identifier: sorted top-member TPNs followed by their corresponding
// bottom-member TPNs. Odd-sized brackets retain their existing sequence-order
// handling.
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
// If no complete pairing is possible, returns ErrNoCompletePairing.
func PairBracket(ctx context.Context, participants []*ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc) ([][2]*ParticipantState, error) {
	n := len(participants)
	if n < 2 {
		return nil, nil
	}

	sorted := make([]*ParticipantState, n)
	copy(sorted, participants)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].TPN < sorted[j].TPN
	})

	if n%2 != 0 {
		used := make([]bool, n)
		pairs := make([][2]*ParticipantState, 0, n/2)
		if ok, err := pairRecursive(ctx, sorted, used, &pairs, forbidden, criteriaFn); err != nil {
			return nil, err
		} else if ok {
			return pairs, nil
		}
		return nil, ErrNoCompletePairing
	}

	var result [][2]*ParticipantState
	found, err := enumerateBracketPairings(ctx, sorted, forbidden, criteriaFn, func(pairs [][2]*ParticipantState) bool {
		result = pairs
		return false
	})
	if err != nil {
		return nil, err
	}
	if found {
		return result, nil
	}
	return nil, ErrNoCompletePairing
}

// enumerateBracketPairings visits complete legal pairings in Article 3.6.3
// identifier order. Returning false from visit stops enumeration.
func enumerateBracketPairings(ctx context.Context, participants []*ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc, visit func([][2]*ParticipantState) bool) (bool, error) {
	tops := make([]int, 0, len(participants)/2)
	return selectTopMembers(ctx, participants, forbidden, criteriaFn, tops, 0, visit)
}

func selectTopMembers(ctx context.Context, participants []*ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc, tops []int, start int, visit func([][2]*ParticipantState) bool) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if len(tops) == len(participants)/2 {
		available := make([]bool, len(participants))
		for _, top := range tops {
			available[top] = true
		}
		return assignBottomMembers(ctx, participants, forbidden, criteriaFn, tops, available, 0, nil, visit)
	}

	needed := len(participants)/2 - len(tops)
	for i := start; i <= len(participants)-needed; i++ {
		nextTops := append(tops, i)
		if !canMatchTops(participants, forbidden, criteriaFn, nextTops) {
			continue
		}
		found, err := selectTopMembers(ctx, participants, forbidden, criteriaFn, nextTops, i+1, visit)
		if err != nil || found {
			return found, err
		}
	}
	return false, nil
}

// canMatchTops reports whether every selected top has a distinct legal bottom
// among participants not selected as a top. It prunes top-member prefixes that
// cannot occur in a complete Article 3.6 pairing without changing their order.
func canMatchTops(participants []*ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc, tops []int) bool {
	selected := make([]bool, len(participants))
	for _, top := range tops {
		selected[top] = true
	}
	bottomFor := make([]int, len(participants))
	for i := range bottomFor {
		bottomFor[i] = -1
	}
	var assign func(int, []bool) bool
	assign = func(top int, seen []bool) bool {
		for bottomIndex, bottom := range participants {
			if selected[bottomIndex] || seen[bottomIndex] || !canPair(participants[top], bottom, forbidden, criteriaFn) {
				continue
			}
			seen[bottomIndex] = true
			if bottomFor[bottomIndex] == -1 || assign(bottomFor[bottomIndex], seen) {
				bottomFor[bottomIndex] = top
				return true
			}
		}
		return false
	}
	for _, top := range tops {
		if !assign(top, make([]bool, len(participants))) {
			return false
		}
	}
	return true
}

func assignBottomMembers(ctx context.Context, participants []*ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc, tops []int, available []bool, pairIndex int, pairs [][2]*ParticipantState, visit func([][2]*ParticipantState) bool) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if pairIndex == len(tops) {
		return !visit(pairs), nil
	}

	top := participants[tops[pairIndex]]
	for bottomIndex, bottom := range participants {
		if available[bottomIndex] || !canPair(top, bottom, forbidden, criteriaFn) {
			continue
		}
		available[bottomIndex] = true
		pair := [2]*ParticipantState{top, bottom}
		found, err := assignBottomMembers(ctx, participants, forbidden, criteriaFn, tops, available, pairIndex+1, append(pairs, pair), visit)
		available[bottomIndex] = false
		if err != nil || found {
			return found, err
		}
	}
	return false, nil
}

func canPair(top, bottom *ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc) bool {
	return bottom.TPN > top.TPN && !HasPlayed(top, bottom) && !isForbidden(top.ID, bottom.ID, forbidden) && (criteriaFn == nil || criteriaFn(top, bottom))
}

// pairRecursive attempts to find a complete pairing using DFS.
// Returns true if a complete pairing is found.
func pairRecursive(ctx context.Context, participants []*ParticipantState, used []bool, pairs *[][2]*ParticipantState, forbidden map[[2]string]bool, criteriaFn CriteriaFunc) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

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
		return true, nil
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
		return true, nil
	}

	// Odd-sized brackets retain the prior sequence order: try pairing the
	// first unused participant with each subsequent participant by ascending TPN.
	used[firstUnused] = true
	for j := firstUnused + 1; j < n; j++ {
		if err := ctx.Err(); err != nil {
			return false, err
		}
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

		if ok, err := pairRecursive(ctx, participants, used, pairs, forbidden, criteriaFn); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}

		// Backtrack.
		used[j] = false
		*pairs = (*pairs)[:len(*pairs)-1]
	}

	used[firstUnused] = false
	return false, nil
}
