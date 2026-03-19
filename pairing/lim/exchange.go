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

// tryExchangePairing implements the Art. 4 exchange algorithm.
// Returns nil if no complete pairing is possible.
func tryExchangePairing(s1, s2 []*swisslib.PlayerState, pairingDownward bool, forbidden map[[2]string]bool) [][2]*swisslib.PlayerState {
	half := len(s1)
	if half == 0 || len(s2) != half {
		return nil
	}

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

	// paired[i] = index into s2 that s1[i] is paired with, -1 if not yet.
	paired := make([]int, half)
	for i := range paired {
		paired[i] = -1
	}
	usedS2 := make([]bool, half)

	for _, si := range scrutinyOrder {
		found := false
		// Try the proposed partner first (same index), then exchange.
		// Exchange order per Art. 4.2: try s2 partners in order from
		// proposed index, wrapping through all available.
		candidates := generateExchangeOrder(si, half, pairingDownward)
		for _, ci := range candidates {
			if usedS2[ci] {
				continue
			}
			if IsCompatible(s1[si], s2[ci], forbidden) {
				paired[si] = ci
				usedS2[ci] = true
				found = true
				break
			}
		}
		if !found {
			// Cannot pair s1[si] — complete pairing impossible.
			return nil
		}
	}

	// Build result.
	result := make([][2]*swisslib.PlayerState, half)
	for i := range half {
		result[i] = [2]*swisslib.PlayerState{s1[i], s2[paired[i]]}
	}
	return result
}

// generateExchangeOrder produces the sequence of S2 indices to try for S1[si].
// Per Art. 4.2: first try the proposed partner (si), then remaining S2 indices
// in order, then try S1 players as exchange partners.
//
// The Art. 4.2 exchange sequence for player #1 in a 6-player group is:
// 4, 5, 6, 3, 2 — which maps to S2 indices first, then S1 indices (excluding self).
//
// For this implementation, we only exchange within S2 (the bottom half).
// The full cross-half exchange is more complex and handled by the
// greedy fallback when this simpler approach fails.
func generateExchangeOrder(si, half int, pairingDownward bool) []int {
	order := make([]int, 0, half)

	// Start with the proposed partner (same index).
	order = append(order, si)

	// Then try remaining S2 indices.
	if pairingDownward {
		// Going up from proposed, then wrapping.
		for j := si + 1; j < half; j++ {
			order = append(order, j)
		}
		for j := si - 1; j >= 0; j-- {
			order = append(order, j)
		}
	} else {
		// Going down from proposed, then wrapping.
		for j := si - 1; j >= 0; j-- {
			order = append(order, j)
		}
		for j := si + 1; j < half; j++ {
			order = append(order, j)
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
