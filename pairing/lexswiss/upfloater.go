// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"context"
	"sort"
)

// SelectUpfloater selects the participant to float up from a bracket to the
// bracket above, per Art. 3.5 (shared by Double-Swiss and Team Swiss).
//
// The upfloater is chosen per Art. 3.5.3-3.5.5 (see
// SelectUpfloaterWithPreference): the first candidate, by descending score and
// ascending TPN, for which both brackets can still be completed. Compatibility
// with the target bracket means: (1) not already played, and (2) not a
// forbidden pair.
//
// Parameters:
//   - bracket: participants in the current bracket (odd count)
//   - targetBracket: participants in the bracket above (where floater goes)
//   - forbidden: forbidden pairs map (nil if none)
//
// Returns nil if no valid upfloater can be selected (all have played
// everyone in the target bracket).
func SelectUpfloater(bracket []*ParticipantState, targetBracket []*ParticipantState, forbidden map[[2]string]bool) *ParticipantState {
	return SelectUpfloaterWithContext(context.Background(), bracket, targetBracket, forbidden)
}

// SelectUpfloaterWithContext is SelectUpfloater with a context.
func SelectUpfloaterWithContext(ctx context.Context, bracket []*ParticipantState, targetBracket []*ParticipantState, forbidden map[[2]string]bool) *ParticipantState {
	return SelectUpfloaterWithPreferenceAndContext(ctx, bracket, targetBracket, forbidden, nil)
}

// SelectUpfloaterWithPreference selects the upfloater per Art. 3.5.3-3.5.5.
//
// Candidates are ordered by descending score and then ascending TPN
// (3.5.3, 3.5.4). The first candidate that yields a legal pairing of the
// target bracket and, when the rest of its own bracket is even, a legal
// pairing of that rest (C6, which needs C1 and C3 to hold there) is chosen
// (3.5.5). Among those, a candidate satisfying preferred wins where possible.
// When no candidate passes the pairing checks it falls back to the first
// candidate with a compatible opponent, so C3 (completion) is never
// sacrificed to a quality criterion.
func SelectUpfloaterWithPreference(bracket []*ParticipantState, targetBracket []*ParticipantState, forbidden map[[2]string]bool, preferred func(*ParticipantState) bool) *ParticipantState {
	return SelectUpfloaterWithPreferenceAndContext(context.Background(), bracket, targetBracket, forbidden, preferred)
}

// SelectUpfloaterWithPreferenceAndContext is SelectUpfloaterWithPreference with a context.
func SelectUpfloaterWithPreferenceAndContext(ctx context.Context, bracket []*ParticipantState, targetBracket []*ParticipantState, forbidden map[[2]string]bool, preferred func(*ParticipantState) bool) *ParticipantState {
	if len(bracket) == 0 {
		return nil
	}
	candidates := make([]*ParticipantState, len(bracket))
	copy(candidates, bracket)
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].TPN < candidates[j].TPN
	})
	for _, strict := range []bool{true, false} {
		for pass := 0; pass < 2; pass++ {
			for _, cand := range candidates {
				if pass == 0 && preferred != nil && !preferred(cand) {
					continue
				}
				if !hasCompatibleOpponent(cand, targetBracket, forbidden) {
					continue
				}
				if strict && !completesBrackets(ctx, cand, bracket, targetBracket, forbidden) {
					continue
				}
				return cand
			}
			if preferred == nil {
				break
			}
		}
	}
	return nil
}

// completesBrackets reports whether floating cand up leaves both brackets
// pairable: the target bracket together with cand, and the remainder of the
// candidate's own bracket. A bracket of odd size is not checked, because its
// completion depends on a later floater.
func completesBrackets(ctx context.Context, cand *ParticipantState, bracket, target []*ParticipantState, forbidden map[[2]string]bool) bool {
	up := append(append([]*ParticipantState{}, target...), cand)
	if len(up)%2 == 0 {
		if _, err := PairBracket(ctx, up, forbidden, nil); err != nil {
			return false
		}
	}
	rest := make([]*ParticipantState, 0, len(bracket))
	for _, p := range bracket {
		if p != cand {
			rest = append(rest, p)
		}
	}
	if len(rest) > 0 && len(rest)%2 == 0 {
		if _, err := PairBracket(ctx, rest, forbidden, nil); err != nil {
			return false
		}
	}
	return true
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
