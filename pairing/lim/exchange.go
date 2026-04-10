// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// ExchangeMatch pairs players in a scoregroup using the Lim exchange
// algorithm (Art. 4). Players are split into top half (S1) and bottom half
// (S2), with initial proposed pairings S1[i] vs S2[i]. When a pairing is
// incompatible, the S2 player is exchanged per Art. 4.2.
//
// Parameters:
//   - players: sorted by TPN ascending within the scoregroup
//   - pairingDownward: true when pairing above the median (scrutiny starts
//     from highest-numbered in top half); false when pairing upward
//   - forbidden: forbidden pairs map (nil if none)
//
// Returns:
//   - pairs: successfully matched pairs [top, bottom]
//   - unpaired: players that could not be paired (must float)
func ExchangeMatch(players []*swisslib.PlayerState, pairingDownward bool, forbidden map[[2]string]bool) (pairs [][2]*swisslib.PlayerState, unpaired []*swisslib.PlayerState) {
	n := len(players)
	if n == 0 {
		return nil, nil
	}

	// Sort by TPN ascending.
	sorted := make([]*swisslib.PlayerState, n)
	copy(sorted, players)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].TPN < sorted[j].TPN
	})

	// Handle odd number: remove last player as unpaired.
	if n%2 == 1 {
		unpaired = append(unpaired, sorted[n-1])
		sorted = sorted[:n-1]
		n = len(sorted)
	}

	if n == 0 {
		return nil, unpaired
	}

	half := n / 2
	s1 := sorted[:half] // top half (lower TPNs)
	s2 := sorted[half:] // bottom half (higher TPNs)

	// Try to pair using exchange algorithm.
	result := tryExchangePairing(s1, s2, pairingDownward, forbidden)
	if result != nil {
		return result, unpaired
	}

	// If complete pairing failed, try to pair as many as possible.
	// Use a greedy fallback: try each S1 player against all available S2 players.
	return greedyPair(sorted, pairingDownward, forbidden, unpaired)
}

// tryExchangePairing implements the Art. 4 exchange algorithm with full
// cross-half exchange support.
//
// Players are arranged in a unified slice [S1 | S2] (indices 0..2*half-1).
// S1 players are processed in scrutiny order. For each, candidates are tried
// in Art. 4.2 order: S2 partners first, then S1 partners (cross-half). When
// an S1 player is consumed as another's cross-half partner, it is skipped
// during scrutiny. After S1 scrutiny, any remaining unpaired players (typically
// S2-S2 leftovers from cross-half displacement) are paired among themselves.
//
// Returns nil if no complete pairing is possible.
func tryExchangePairing(s1, s2 []*swisslib.PlayerState, pairingDownward bool, forbidden map[[2]string]bool) [][2]*swisslib.PlayerState {
	half := len(s1)
	if half == 0 || len(s2) != half {
		return nil
	}

	// Build unified player slice: S1 (0..half-1) + S2 (half..2*half-1).
	allPlayers := make([]*swisslib.PlayerState, 0, 2*half)
	allPlayers = append(allPlayers, s1...)
	allPlayers = append(allPlayers, s2...)

	// Generate the order of S1 players to scrutinise.
	// When pairing downward: start with highest-numbered (Art. 4.1.1 says
	// "scrutiny begins with the highest numbered player").
	// When pairing upward: start with lowest-numbered (Art. 4.1.2).
	scrutinyOrder := make([]int, half)
	if pairingDownward {
		for i := range half {
			scrutinyOrder[i] = half - 1 - i // highest first
		}
	} else {
		for i := range half {
			scrutinyOrder[i] = i // lowest first
		}
	}

	used := make([]bool, 2*half)
	var pairs [][2]*swisslib.PlayerState

	for _, si := range scrutinyOrder {
		if used[si] {
			// Already consumed as another S1 player's cross-half partner.
			continue
		}

		candidates := generateExchangeOrder(si, half, pairingDownward)
		found := false
		for _, ci := range candidates {
			if used[ci] {
				continue
			}
			if IsCompatible(allPlayers[si], allPlayers[ci], forbidden) {
				pairs = append(pairs, [2]*swisslib.PlayerState{allPlayers[si], allPlayers[ci]})
				used[si] = true
				used[ci] = true
				found = true
				break
			}
		}
		if !found {
			// Cannot pair this S1 player — complete pairing impossible.
			return nil
		}
	}

	// After S1 scrutiny, pair any remaining unpaired players (S2-S2 leftovers
	// from cross-half displacement, or S1-S1 leftovers).
	var remaining []int
	for i, u := range used {
		if !u {
			remaining = append(remaining, i)
		}
	}
	if len(remaining)%2 != 0 {
		return nil // odd leftovers — complete pairing impossible
	}
	for i := 0; i < len(remaining); i++ {
		found := false
		for j := i + 1; j < len(remaining); j++ {
			ri, rj := remaining[i], remaining[j]
			if IsCompatible(allPlayers[ri], allPlayers[rj], forbidden) {
				pairs = append(pairs, [2]*swisslib.PlayerState{allPlayers[ri], allPlayers[rj]})
				// Remove j from remaining (swap with last, shrink).
				remaining[j] = remaining[len(remaining)-1]
				remaining = remaining[:len(remaining)-1]
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	// Sort pairs by the lower TPN in each pair for consistent ordering
	// (matches the old behavior of returning pairs in S1 index order).
	sort.SliceStable(pairs, func(i, j int) bool {
		minI := pairs[i][0].TPN
		if pairs[i][1].TPN < minI {
			minI = pairs[i][1].TPN
		}
		minJ := pairs[j][0].TPN
		if pairs[j][1].TPN < minJ {
			minJ = pairs[j][1].TPN
		}
		return minI < minJ
	})

	return pairs
}

// generateExchangeOrder produces the full Art. 4.2 exchange sequence for S1[si].
//
// Returns unified indices into the [S1 | S2] slice (0..half-1 = S1,
// half..2*half-1 = S2). The sequence is:
//
//  1. S2 partners: proposed partner (half+si) first, then remaining S2 in
//     exchange order (upward then wrap when pairing down, downward then wrap
//     when pairing up).
//  2. S1 cross-half partners: other S1 players excluding self, in scrutiny-
//     direction order (highest-first when pairing down, lowest-first when up).
//
// Example: player #1 in a 6-player group (S1={1,2,3}, S2={4,5,6}), pairing
// downward: exchange sequence is 4, 5, 6, 3, 2.
func generateExchangeOrder(si, half int, pairingDownward bool) []int {
	order := make([]int, 0, 2*half-1) // all players except self

	// Phase 1: S2 partners (offset by half into unified index space).
	order = append(order, half+si) // proposed partner
	if pairingDownward {
		for j := si + 1; j < half; j++ {
			order = append(order, half+j)
		}
		for j := si - 1; j >= 0; j-- {
			order = append(order, half+j)
		}
	} else {
		for j := si - 1; j >= 0; j-- {
			order = append(order, half+j)
		}
		for j := si + 1; j < half; j++ {
			order = append(order, half+j)
		}
	}

	// Phase 2: S1 cross-half partners (excluding self).
	if pairingDownward {
		// Highest to lowest (same direction as scrutiny).
		for j := half - 1; j >= 0; j-- {
			if j != si {
				order = append(order, j)
			}
		}
	} else {
		// Lowest to highest.
		for j := range half {
			if j != si {
				order = append(order, j)
			}
		}
	}

	return order
}

// greedyPair pairs as many players as possible using a greedy approach.
// Returns matched pairs and any remaining unpaired players (appended to existing unpaired).
func greedyPair(players []*swisslib.PlayerState, _ bool, forbidden map[[2]string]bool, existingUnpaired []*swisslib.PlayerState) ([][2]*swisslib.PlayerState, []*swisslib.PlayerState) {
	n := len(players)
	used := make([]bool, n)
	var pairs [][2]*swisslib.PlayerState

	// Try to pair each unused player with the best available partner.
	for i := range n {
		if used[i] {
			continue
		}
		for j := i + 1; j < n; j++ {
			if used[j] {
				continue
			}
			if IsCompatible(players[i], players[j], forbidden) {
				pairs = append(pairs, [2]*swisslib.PlayerState{players[i], players[j]})
				used[i] = true
				used[j] = true
				break
			}
		}
	}

	// Collect unpaired.
	unpaired := existingUnpaired
	for i, u := range used {
		if !u {
			unpaired = append(unpaired, players[i])
		}
	}

	return pairs, unpaired
}
