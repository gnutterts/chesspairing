// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import "sort"

// SelectUpfloater selects the participant to float up from a bracket to the
// bracket above, per Art. 3.5 (shared by Double-Swiss and Team Swiss).
//
// The upfloater is the lowest-ranked participant (highest TPN) who has at
// least one compatible opponent in the target bracket. Compatibility means:
// (1) not already played, and (2) not a forbidden pair.
//
// Parameters:
//   - bracket: participants in the current bracket (odd count)
//   - targetBracket: participants in the bracket above (where floater goes)
//   - forbidden: forbidden pairs map (nil if none)
//
// Returns nil if no valid upfloater can be selected (all have played
// everyone in the target bracket).
func SelectUpfloater(bracket []*ParticipantState, targetBracket []*ParticipantState, forbidden map[[2]string]bool) *ParticipantState {
	if len(bracket) == 0 {
		return nil
	}

	// Sort candidates by TPN descending (highest TPN = lowest ranking, tried first).
	candidates := make([]*ParticipantState, len(bracket))
	copy(candidates, bracket)
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].TPN > candidates[j].TPN
	})

	for _, cand := range candidates {
		if hasCompatibleOpponent(cand, targetBracket, forbidden) {
			return cand
		}
	}

	return nil
}

// hasCompatibleOpponent returns true if the participant has at least one
// compatible opponent in the target bracket.
func hasCompatibleOpponent(p *ParticipantState, targets []*ParticipantState, forbidden map[[2]string]bool) bool {
	for _, t := range targets {
		if !HasPlayed(p, t) && !isForbidden(p.ID, t.ID, forbidden) {
			return true
		}
	}
	return false
}

// isForbidden checks if two participants form a forbidden pair.
func isForbidden(id1, id2 string, forbidden map[[2]string]bool) bool {
	if forbidden == nil {
		return false
	}
	return forbidden[[2]string{id1, id2}] || forbidden[[2]string{id2, id1}]
}
