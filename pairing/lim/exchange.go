package lim

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// ExchangeMatch applies FIDE C.04.4.3 Articles 2.3 and 4.1-4.4.
func ExchangeMatch(players []*swisslib.PlayerState, pairingDownward bool, forbidden map[[2]string]bool) ([][2]*swisslib.PlayerState, []*swisslib.PlayerState) {
	return ExchangeMatchWithCompatibility(players, pairingDownward, forbidden, func(a, b *swisslib.PlayerState) bool {
		return IsCompatible(a, b, forbidden)
	})
}

// ExchangeMatchWithCompatibility is ExchangeMatch with the compatibility rule
// supplied by the caller. It is used for Article 6 in the last round.
func ExchangeMatchWithCompatibility(players []*swisslib.PlayerState, pairingDownward bool, _ map[[2]string]bool, compatible func(a, b *swisslib.PlayerState) bool) ([][2]*swisslib.PlayerState, []*swisslib.PlayerState) {
	return exchangeMatchWithFloaterSelector(players, pairingDownward, compatible, nil)
}

// exchangeMatchWithFloaterSelector permits Pair to apply the complete Arts.
// 3.2-3.5, 3.9 and 3.10 selection to every Art. 2.3.4 choice, including the
// one after Arts. 2.3.1-2.3.3 have transferred incompatible players.
func exchangeMatchWithFloaterSelector(players []*swisslib.PlayerState, pairingDownward bool, compatible func(a, b *swisslib.PlayerState) bool, selectInitial func([]*swisslib.PlayerState) *swisslib.PlayerState) ([][2]*swisslib.PlayerState, []*swisslib.PlayerState) {
	sorted := append([]*swisslib.PlayerState(nil), players...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].TPN < sorted[j].TPN })

	// Art. 2.3.1-2.3.3: a player with no suitable opponent is transferred
	// before the scoregroup is divided into S1 and S2.
	unpaired := make([]*swisslib.PlayerState, 0)
	pairable := make([]*swisslib.PlayerState, 0, len(sorted))
	for _, player := range sorted {
		for _, opponent := range sorted {
			if player.ID != opponent.ID && compatible(player, opponent) {
				pairable = append(pairable, player)
				break
			}
		}
		if len(pairable) == 0 || pairable[len(pairable)-1] != player {
			unpaired = append(unpaired, player)
		}
	}
	sorted = pairable

	var initialFloater *swisslib.PlayerState
	// This records an actual Art. 2.3.4 choice, rather than merely an
	// unpaired player transferred under Arts. 2.3.1-2.3.3.
	hadInitial23Floater := false
	if len(sorted)%2 != 0 {
		hadInitial23Floater = true
		// Art. 2.3.4.  In Lim terminology the lowest numbered player is the
		// bottom-ranked player (the largest TPN).
		if selectInitial != nil {
			initialFloater = selectInitial(sorted)
		}
		if initialFloater == nil {
			if pairingDownward {
				initialFloater = sorted[len(sorted)-1]
			} else {
				initialFloater = sorted[0]
			}
		}
		sorted = removePtrs(sorted, initialFloater)
		unpaired = append(unpaired, initialFloater)
	}
	if len(sorted) == 0 {
		return nil, unpaired
	}

	proposal := initialProposal(len(sorted))
	half := len(sorted) / 2
	order := make([]int, half)
	for i := range half {
		order[i] = i // Art. 4.1.1: downward starts with #1.
		if !pairingDownward {
			order[i] = len(sorted) - 1 - i // Art. 4.1.2: upward starts with #n.
		}
	}

	for position, player := range order {
		if compatible(sorted[player], sorted[proposal[player]]) {
			continue
		}
		found := false
		for _, candidate := range exchangeOrderFor(player, half, pairingDownward) {
			if candidate == player || candidate == proposal[player] {
				continue
			}
			next := fideExchangedProposal(proposal, player, candidate)
			if !compatible(sorted[player], sorted[next[player]]) || !earlierPairsCompatible(sorted, next, order[:position], compatible) {
				continue
			}
			proposal = next
			found = true
			break
		}
		if found {
			continue
		}

		// Art. 4.4: retain the earlier proposed pairs; only the player for
		// which no exchange works floats.  Art. 4.4.1 uses the 2.3.4 floater;
		// Art. 4.4.2 floats the lowest numbered remaining player in company.
		retained := make(map[int]bool, 2*position)
		for _, earlier := range order[:position] {
			retained[earlier] = true
			retained[proposal[earlier]] = true
		}
		var companion *swisslib.PlayerState
		if hadInitial23Floater {
			// Arts. 4.4/4.4.1/4.4.2 say that when the lowest numbered player
			// was floated to make the scoregroup even, the failed player is
			// exchanged with that floater; otherwise that player floats with
			// the lowest numbered remaining companion. We apply that reading
			// even when 2.3.1-2.3.3 made an originally even group odd.
			// The 2.3.4 floater returns to this scoregroup; only the failed
			// player floats.
			unpaired = removePtrs(unpaired, initialFloater)
			unpaired = appendUniqueFloaters(unpaired, sorted[player])
		} else {
			for i := len(sorted) - 1; i >= 0; i-- {
				// Art. 4.4.2 permits the failed player's proposed partner to be
				// the companion: it is the lowest numbered player remaining.
				if i != player && !retained[i] {
					companion = sorted[i]
					break
				}
			}
			unpaired = appendUniqueFloaters(unpaired, sorted[player], companion)
		}

		// Keep the earlier pairs, then apply the exchange rules anew to all
		// players left after the Art. 4.4 floaters.  Dropping their proposed
		// pairs would leave their former partners unaccounted for.
		retainedPairs := pairsAt(sorted, proposal, retained)
		remaining := playersExcept(sorted, retained, unpaired)
		if hadInitial23Floater {
			remaining = append(remaining, initialFloater)
		}
		// Art. 3.10 only protects a floater selected under Art. 2.3.4. The
		// companion added by Art. 4.4.2 is not such a selection.
		remainingPairs, remainingFloaters := exchangeMatchWithFloaterSelector(remaining, pairingDownward, compatible, nil)
		unpaired = appendUniqueFloaters(unpaired, remainingFloaters...)
		return append(retainedPairs, remainingPairs...), unpaired
	}

	return proposalPairs(sorted, proposal, unpaired, compatible), unpaired
}

func initialProposal(n int) []int {
	proposal := make([]int, n)
	half := n / 2
	for i := range half {
		proposal[i], proposal[half+i] = half+i, i
	}
	return proposal
}

func fideExchangedProposal(proposal []int, player, candidate int) []int {
	// The last two columns of the six-player tables use a three-pair
	// displacement.  These are the literal proposals printed in Arts. 4.2
	// and 4.3, rather than a bipartite-matching approximation.
	if len(proposal) == 6 && player == 0 {
		switch candidate {
		case 5:
			return proposalFromPairs(6, [][2]int{{0, 5}, {1, 3}, {2, 4}})
		case 1:
			return proposalFromPairs(6, [][2]int{{0, 1}, {2, 4}, {3, 5}})
		}
	}
	if len(proposal) == 6 && player == 1 && sameProposal(proposal, [][2]int{{0, 5}, {1, 3}, {2, 4}}) {
		switch candidate {
		case 5:
			return proposalFromPairs(6, [][2]int{{0, 2}, {1, 5}, {3, 4}})
		case 0:
			return proposalFromPairs(6, [][2]int{{0, 1}, {2, 4}, {3, 5}})
		}
	}
	return exchangedProposal(proposal, player, candidate)
}

func exchangedProposal(proposal []int, player, candidate int) []int {
	next := append([]int(nil), proposal...)
	playerPartner, candidatePartner := next[player], next[candidate]
	next[player], next[candidate] = candidate, player
	next[playerPartner], next[candidatePartner] = candidatePartner, playerPartner
	return next
}

func proposalFromPairs(n int, pairs [][2]int) []int {
	proposal := make([]int, n)
	for _, pair := range pairs {
		proposal[pair[0]], proposal[pair[1]] = pair[1], pair[0]
	}
	return proposal
}

func sameProposal(proposal []int, pairs [][2]int) bool {
	want := proposalFromPairs(len(proposal), pairs)
	for i := range proposal {
		if proposal[i] != want[i] {
			return false
		}
	}
	return true
}

func earlierPairsCompatible(players []*swisslib.PlayerState, proposal, earlier []int, compatible func(a, b *swisslib.PlayerState) bool) bool {
	for _, player := range earlier {
		if !compatible(players[player], players[proposal[player]]) {
			return false
		}
	}
	return true
}

func appendUniqueFloaters(floaters []*swisslib.PlayerState, players ...*swisslib.PlayerState) []*swisslib.PlayerState {
	seen := make(map[string]bool, len(floaters)+len(players))
	for _, player := range floaters {
		seen[player.ID] = true
	}
	for _, player := range players {
		if player != nil && !seen[player.ID] {
			floaters = append(floaters, player)
			seen[player.ID] = true
		}
	}
	return floaters
}

func pairsAt(players []*swisslib.PlayerState, proposal []int, retained map[int]bool) [][2]*swisslib.PlayerState {
	pairs := make([][2]*swisslib.PlayerState, 0, len(retained)/2)
	for player := range players {
		partner := proposal[player]
		if retained[player] && player < partner {
			pairs = append(pairs, [2]*swisslib.PlayerState{players[player], players[partner]})
		}
	}
	return pairs
}

func playersExcept(players []*swisslib.PlayerState, retained map[int]bool, floaters []*swisslib.PlayerState) []*swisslib.PlayerState {
	floating := make(map[string]bool, len(floaters))
	for _, player := range floaters {
		floating[player.ID] = true
	}
	remaining := make([]*swisslib.PlayerState, 0, len(players))
	for i, player := range players {
		if !retained[i] && !floating[player.ID] {
			remaining = append(remaining, player)
		}
	}
	return remaining
}

func proposalPairs(players []*swisslib.PlayerState, proposal []int, unpaired []*swisslib.PlayerState, compatible func(a, b *swisslib.PlayerState) bool) [][2]*swisslib.PlayerState {
	floating := make(map[string]bool, len(unpaired))
	for _, player := range unpaired {
		floating[player.ID] = true
	}
	pairs := make([][2]*swisslib.PlayerState, 0, len(players)/2)
	for i, partner := range proposal {
		if i < partner && !floating[players[i].ID] && !floating[players[partner].ID] && compatible(players[i], players[partner]) {
			pairs = append(pairs, [2]*swisslib.PlayerState{players[i], players[partner]})
		}
	}
	return pairs
}

// exchangeOrderFor implements the Art. 4.2 order: S2 candidates first, then
// S1 candidates.  The upward order is the corresponding reversal of ranks.
func exchangeOrderFor(player, half int, downward bool) []int {
	if downward {
		// For #2 the printed Art. 4.3 table tries #4 (its own partner), #5,
		// #3, #6, #1.  The caller skips the own-partner entry.
		if half == 3 && player == 1 {
			return []int{3, 4, 2, 5, 0}
		}
		return generateExchangeOrder(player, half)
	}
	// Upward scrutiny reflects the generic S2-then-S1 sequence across ranks.
	mirror := half*2 - 1 - player
	order := generateExchangeOrder(mirror, half)
	for i := range order {
		order[i] = half*2 - 1 - order[i]
	}
	return order
}

// generateExchangeOrder returns the Art. 4.2 S2-then-S1 candidate sequence
// for a downward pairing (including the five columns for six players).
func generateExchangeOrder(si, half int) []int {
	order := make([]int, 0, 2*half-1)
	for sj := si; sj < half; sj++ {
		order = append(order, half+sj)
	}
	for sj := 0; sj < si; sj++ {
		order = append(order, half+sj)
	}
	for sj := half - 1; sj >= 0; sj-- {
		if sj != si {
			order = append(order, sj)
		}
	}
	return order
}
