// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import "github.com/gnutterts/chesspairing/pairing/swisslib"

// ColourExchange performs the Art. 5.2 second scrutiny: after ExchangeMatch
// has established legal pairings, this pass swaps opponents between pairs to
// give each player, if possible, the alternating and equalising colour.
//
// In Maxi-tournaments (Art. 5.7), an exchange of opponents is only allowed
// if the ratings of the exchanged players differ by 100 points or less.
func ColourExchange(pairs [][2]*swisslib.PlayerState, forbidden map[[2]string]bool, isMaxi bool) [][2]*swisslib.PlayerState {
	// Iterate over conflicted pairs, trying to resolve each by swapping
	// with another pair. After a successful swap, restart the scan since
	// the swap may have introduced or resolved other conflicts.
	for {
		improved := false
		for i := range pairs {
			if !hasColourConflict(pairs[i][0], pairs[i][1]) {
				continue
			}
			// Art. 5.2 asks for "a second scrutiny with exchanges where
			// necessary" without limiting them to later pairs: try the later
			// pairs first, then the earlier ones.
			for _, j := range exchangeCandidates(i, len(pairs)) {
				if trySwap(pairs, i, j, forbidden, isMaxi) {
					improved = true
					break
				}
			}
			if improved {
				break
			}
		}
		if !improved {
			break
		}
	}

	return pairs
}

// exchangeCandidates lists the other pairs for pair i: later pairs first, then
// earlier pairs.
func exchangeCandidates(i, n int) []int {
	candidates := make([]int, 0, n-1)
	for j := i + 1; j < n; j++ {
		candidates = append(candidates, j)
	}
	for j := 0; j < i; j++ {
		candidates = append(candidates, j)
	}
	return candidates
}

// hasColourConflict returns true if both players in a pair are due the same
// colour, meaning no colour assignment can satisfy both preferences.
//
// If either player has no preference (nil), there is no conflict.
func hasColourConflict(a, b *swisslib.PlayerState) bool {
	prefA := swisslib.ComputeColorPreference(a.ColorHistory)
	prefB := swisslib.ComputeColorPreference(b.ColorHistory)

	if prefA.Color == nil || prefB.Color == nil {
		return false
	}
	return *prefA.Color == *prefB.Color
}

// trySwap attempts to swap opponents between pairs[i] and pairs[j] to reduce
// the total number of colour conflicts across those two pairs.
//
// For pairs[i] = {a, b} and pairs[j] = {c, d}, two swaps are tried:
//   - Swap b↔c: new pairs {a, c} and {b, d}
//   - Swap b↔d: new pairs {a, d} and {b, c}
//
// A swap is valid only if:
//  1. Both new pairs are compatible (IsCompatible).
//  2. The total colour conflicts for the two pairs decreases.
//  3. In Maxi-tournaments: the exchanged players' ratings differ by ≤100.
//
// If valid, pairs[i] and pairs[j] are mutated in-place and true is returned.
func trySwap(pairs [][2]*swisslib.PlayerState, i, j int, forbidden map[[2]string]bool, isMaxi bool) bool {
	a, b := pairs[i][0], pairs[i][1]
	c, d := pairs[j][0], pairs[j][1]

	oldConflicts := conflictCount2(a, b, c, d)

	// Art. 5.2 permits both opponent exchanges, in the Art. 4 order.
	// b↔c may create S1-S1/S2-S2 pairs; Art. 4.2 expressly permits those.
	if canSwap(a, c, b, d, b, c, forbidden, isMaxi) {
		newConflicts := conflictCount2(a, c, b, d)
		if newConflicts < oldConflicts {
			pairs[i] = [2]*swisslib.PlayerState{a, c}
			pairs[j] = [2]*swisslib.PlayerState{b, d}
			return true
		}
	}

	// b↔d → {a, d} and {b, c}.
	if canSwap(a, d, b, c, b, d, forbidden, isMaxi) {
		newConflicts := conflictCount2(a, d, b, c)
		if newConflicts < oldConflicts {
			pairs[i] = [2]*swisslib.PlayerState{a, d}
			pairs[j] = [2]*swisslib.PlayerState{b, c}
			return true
		}
	}

	return false
}

// canSwap checks whether two proposed new pairs are compatible and whether
// the maxi rating constraint is met for the two exchanged players.
func canSwap(new1a, new1b, new2a, new2b, swapped1, swapped2 *swisslib.PlayerState, forbidden map[[2]string]bool, isMaxi bool) bool {
	if !IsCompatible(new1a, new1b, forbidden) {
		return false
	}
	if !IsCompatible(new2a, new2b, forbidden) {
		return false
	}
	if isMaxi && absInt(swapped1.Rating-swapped2.Rating) > 100 {
		return false
	}
	return true
}

// conflictCount2 returns the number of colour conflicts across two pairs
// {a, b} and {c, d}.
func conflictCount2(a, b, c, d *swisslib.PlayerState) int {
	n := 0
	if hasColourConflict(a, b) {
		n++
	}
	if hasColourConflict(c, d) {
		n++
	}
	return n
}
