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
	if len(pairs) < 3 {
		// With only 2 pairs in a scoregroup, swapping opponents changes every
		// pairing in the group. In later rounds where the opponent pool is
		// already constrained, this can make future rounds infeasible (e.g.,
		// the swap creates a combination where some player has played all
		// remaining opponents). AllocateColor handles colour conflicts for
		// small groups by giving one player a suboptimal colour, which is
		// less disruptive than rearranging all pairings.
		return pairs
	}

	// Iterate over conflicted pairs, trying to resolve each by swapping
	// with another pair. After a successful swap, restart the scan since
	// the swap may have introduced or resolved other conflicts.
	for {
		improved := false
		for i := range pairs {
			if !hasColourConflict(pairs[i][0], pairs[i][1]) {
				continue
			}
			for j := i + 1; j < len(pairs); j++ {
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

	// Swap 1: b↔c → {a, c} and {b, d}.
	if canSwap(a, c, b, d, b, c, forbidden, isMaxi) {
		newConflicts := conflictCount2(a, c, b, d)
		if newConflicts < oldConflicts {
			pairs[i] = [2]*swisslib.PlayerState{a, c}
			pairs[j] = [2]*swisslib.PlayerState{b, d}
			return true
		}
	}

	// Swap 2: b↔d → {a, d} and {b, c}.
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
