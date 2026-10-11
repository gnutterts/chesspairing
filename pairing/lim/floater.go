// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// FloaterType classifies a floater per Art. 3.9.
// Lower values indicate more disadvantage (worse).
type FloaterType int

const (
	FloaterTypeA FloaterType = iota // already floated + no compatible opponent in adjacent
	FloaterTypeB                    // already floated + has compatible opponent in adjacent
	FloaterTypeC                    // not floated + no compatible opponent in adjacent
	FloaterTypeD                    // not floated + has compatible opponent in adjacent
)

// String returns the floater type name.
func (ft FloaterType) String() string {
	switch ft {
	case FloaterTypeA:
		return "A"
	case FloaterTypeB:
		return "B"
	case FloaterTypeC:
		return "C"
	case FloaterTypeD:
		return "D"
	default:
		return "?"
	}
}

// ClassifyFloater determines the floater type for a player per Art. 3.9.
//
// Parameters:
//   - p: the player being classified
//   - alreadyFloated: true if the player was already floated into this scoregroup
//   - adjacentPlayers: players in the adjacent scoregroup (the one p would float to)
//   - forbidden: forbidden pairs map (nil if none)
//
// This applies the plain compatibility rule. Pair uses the last-round-aware
// variant internally when Article 6 relaxes compatibility.
func ClassifyFloater(p *swisslib.PlayerState, alreadyFloated bool, adjacentPlayers []*swisslib.PlayerState, forbidden map[[2]string]bool) FloaterType {
	return classifyFloaterWithCompatibility(p, alreadyFloated, adjacentPlayers, func(a, b *swisslib.PlayerState) bool {
		return IsCompatible(a, b, forbidden)
	})
}

// SelectDownFloater selects a down-floater per Arts. 3.2-3.4.
// Art. 3.10 is applied only through Pair, which supplies previous-round floaters.
func SelectDownFloater(players []*swisslib.PlayerState, adjacent []*swisslib.PlayerState, forbidden map[[2]string]bool, floatedIDs map[string]bool, maxiTournament bool) *swisslib.PlayerState {
	return selectFloater(players, adjacent, forbidden, floatedIDs, maxiTournament, true, nil, false, nil)
}

// SelectUpFloater selects an up-floater per Arts. 3.2 and 3.4.
// Art. 3.10 is applied only through Pair, which supplies previous-round floaters.
func SelectUpFloater(players []*swisslib.PlayerState, adjacent []*swisslib.PlayerState, forbidden map[[2]string]bool, floatedIDs map[string]bool, maxiTournament bool) *swisslib.PlayerState {
	return selectFloater(players, adjacent, forbidden, floatedIDs, maxiTournament, false, nil, false, nil)
}

// selectFloater implements Arts. 3.2-3.5 and 3.9. Article 3.3 says "the
// player chosen is the lowest numbered player ... who has a compatible
// opponent ... after excluding the opponents of other floaters"; upward 3.4
// says "the highest numbered player ... who has a compatible opponent ...
// after excluding the opponents of other floaters". Article 3.5 says a
// proposed floater without one "shall, if possible, be exchanged". Within
// that requirement, 3.9.2 says floaters "should be chosen to minimise the
// disadvantages": D, B, C, A.
//
// protectPrevious applies Art. 3.10 solely to a 2.3.4 transfer. A protected
// floater is replaced only by type D and only when the exchange engine makes
// no fewer pairs in this scoregroup (Arts. 3.10.1 and 3.10.2).
func selectFloater(players, adjacent []*swisslib.PlayerState, forbidden map[[2]string]bool, floatedIDs map[string]bool, maxiTournament, downward bool, previousFloaters map[string]bool, protectPrevious bool, compatible func(a, b *swisslib.PlayerState) bool) *swisslib.PlayerState {
	return selectFloaterWithEarlier(players, adjacent, forbidden, floatedIDs, maxiTournament, downward, previousFloaters, protectPrevious, nil, compatible)
}

// selectFloaterWithEarlier excludes the opponents that floaters which precede
// each candidate under Arts. 3.6/3.7 will take under Art. 3.8.
func selectFloaterWithEarlier(players, adjacent []*swisslib.PlayerState, forbidden map[[2]string]bool, floatedIDs map[string]bool, maxiTournament, downward bool, previousFloaters map[string]bool, protectPrevious bool, earlier []FloaterEntry, compatible func(a, b *swisslib.PlayerState) bool) *swisslib.PlayerState {
	if compatible == nil {
		compatible = func(a, b *swisslib.PlayerState) bool {
			return IsCompatible(a, b, forbidden)
		}
	}
	if len(players) == 0 {
		return nil
	}
	dueWhite, dueBlack := countDueColors(players)
	var preferDue *swisslib.Color
	if dueWhite > dueBlack {
		w := swisslib.ColorWhite
		preferDue = &w
	} else if dueBlack > dueWhite {
		b := swisslib.ColorBlack
		preferDue = &b
	}
	type candidate struct {
		player      *swisslib.PlayerState
		floaterType FloaterType
		dueColor    *swisslib.Color
	}
	candidates := make([]candidate, 0, len(players))
	for _, p := range players {
		available := adjacentAfterEarlierFloaters(p, adjacent, earlier, downward, maxiTournament, forbidden)
		pref := swisslib.ComputeColorPreference(p.ColorHistory)
		candidates = append(candidates, candidate{
			player:      p,
			floaterType: classifyFloaterWithCompatibility(p, floatedIDs[p.ID], available, compatible),
			dueColor:    pref.Color,
		})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].floaterType != candidates[j].floaterType {
			return floaterPriority(candidates[i].floaterType) > floaterPriority(candidates[j].floaterType)
		}
		matchI := preferDue != nil && candidates[i].dueColor != nil && *candidates[i].dueColor == *preferDue
		matchJ := preferDue != nil && candidates[j].dueColor != nil && *candidates[j].dueColor == *preferDue
		if matchI != matchJ {
			return matchI
		}
		if downward {
			return candidates[i].player.TPN > candidates[j].player.TPN
		}
		return candidates[i].player.TPN < candidates[j].player.TPN
	})

	selected := candidates[0].player
	if protectPrevious && previousFloaters[selected.ID] {
		selector := func(pairable []*swisslib.PlayerState) *swisslib.PlayerState {
			return selectFloaterWithEarlier(pairable, adjacent, forbidden, floatedIDs, maxiTournament, downward, previousFloaters, protectPrevious, earlier, compatible)
		}
		type outcome struct {
			pairCount int
			floaters  []*swisslib.PlayerState
		}
		// A simulation can be requested more than once as nested 2.3.4
		// selections protect prior floaters. Cache its result for this exact
		// scoregroup and candidate.
		outcomes := make(map[string]outcome)
		simulate := func(floater *swisslib.PlayerState) outcome {
			if result, ok := outcomes[floater.ID]; ok {
				return result
			}
			pairCount, floaters := pairingAfterFloating(players, floater, downward, compatible, selector)
			result := outcome{pairCount: pairCount, floaters: floaters}
			outcomes[floater.ID] = result
			return result
		}
		selectedPairCount := simulate(selected).pairCount
		for _, replacement := range candidates[1:] {
			if replacement.floaterType != FloaterTypeD || previousFloaters[replacement.player.ID] {
				continue
			}
			// Art. 3.10.1 requires that the replacement produce no other type
			// a, b or c floater; 3.10.2 says it must not decrease this group's
			// pair count. Simulate the real compatibility and 2.3.4 selector.
			result := simulate(replacement.player)
			if result.pairCount >= selectedPairCount && !hasABCFloater(result.floaters, adjacent, floatedIDs, compatible) {
				selected = replacement.player
				break
			}
		}
	}

	// Art. 3.2.3 is an unconditional Maxi override.
	if maxiTournament {
		reference := highestTPNPlayer(players)
		if !downward {
			reference = lowestTPNPlayer(players)
		}
		if reference != nil && selected.ID != reference.ID && absInt(selected.Rating-reference.Rating) > 100 {
			selected = reference
		}
	}
	return selected
}

func adjacentAfterEarlierFloaters(proposed *swisslib.PlayerState, adjacent []*swisslib.PlayerState, earlier []FloaterEntry, downward, maxiTournament bool, forbidden map[[2]string]bool) []*swisslib.PlayerState {
	preceding := make([]FloaterEntry, 0, len(earlier))
	for _, floater := range earlier {
		if floater.Direction != FloatDown && floater.Direction != FloatUp {
			continue
		}
		if floaterPrecedes(floater, proposed, downward) {
			preceding = append(preceding, floater)
		}
	}
	if len(preceding) == 0 {
		return adjacent
	}
	pairs, _, _ := PairFloaters(preceding, adjacent, downward, maxiTournament, forbidden)
	taken := make(map[string]bool, len(pairs))
	for _, pair := range pairs {
		taken[pair[1].ID] = true
	}
	available := make([]*swisslib.PlayerState, 0, len(adjacent)-len(taken))
	for _, opponent := range adjacent {
		if !taken[opponent.ID] {
			available = append(available, opponent)
		}
	}
	return available
}

func floaterPrecedes(floater FloaterEntry, proposed *swisslib.PlayerState, downward bool) bool {
	if downward {
		return floater.SourceScore > proposed.Score ||
			(floater.SourceScore == proposed.Score && floater.Player.TPN < proposed.TPN)
	}
	return floater.SourceScore < proposed.Score ||
		(floater.SourceScore == proposed.Score && floater.Player.TPN > proposed.TPN)
}

func pairingAfterFloating(players []*swisslib.PlayerState, floater *swisslib.PlayerState, downward bool, compatible func(a, b *swisslib.PlayerState) bool, selector func([]*swisslib.PlayerState) *swisslib.PlayerState) (int, []*swisslib.PlayerState) {
	remaining := removePtrs(players, floater)
	pairs, floaters := exchangeMatchWithFloaterSelector(remaining, downward, compatible, selector)
	return len(pairs), floaters
}

func hasABCFloater(floaters, adjacent []*swisslib.PlayerState, floatedIDs map[string]bool, compatible func(a, b *swisslib.PlayerState) bool) bool {
	for _, floater := range floaters {
		if classifyFloaterWithCompatibility(floater, floatedIDs[floater.ID], adjacent, compatible) != FloaterTypeD {
			return true
		}
	}
	return false
}

func classifyFloaterWithCompatibility(p *swisslib.PlayerState, alreadyFloated bool, adjacentPlayers []*swisslib.PlayerState, compatible func(a, b *swisslib.PlayerState) bool) FloaterType {
	hasCompatible := false
	for _, adjacent := range adjacentPlayers {
		if compatible(p, adjacent) {
			hasCompatible = true
			break
		}
	}
	switch {
	case alreadyFloated && !hasCompatible:
		return FloaterTypeA
	case alreadyFloated && hasCompatible:
		return FloaterTypeB
	case !alreadyFloated && !hasCompatible:
		return FloaterTypeC
	default:
		return FloaterTypeD
	}
}

func floaterPriority(floaterType FloaterType) int {
	switch floaterType {
	case FloaterTypeD:
		return 4
	case FloaterTypeB:
		return 3
	case FloaterTypeC:
		return 2
	case FloaterTypeA:
		return 1
	default:
		return 0
	}
}

func countDueColors(players []*swisslib.PlayerState) (dueWhite, dueBlack int) {
	for _, p := range players {
		pref := swisslib.ComputeColorPreference(p.ColorHistory)
		if pref.Color != nil {
			if *pref.Color == swisslib.ColorWhite {
				dueWhite++
			} else {
				dueBlack++
			}
		}
	}
	return
}

// absInt returns the absolute value of x.
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// lowestTPNPlayer returns the player with the lowest PairingNumber in the slice.
func lowestTPNPlayer(players []*swisslib.PlayerState) *swisslib.PlayerState {
	if len(players) == 0 {
		return nil
	}
	best := players[0]
	for _, p := range players[1:] {
		if p.TPN < best.TPN {
			best = p
		}
	}
	return best
}

// highestTPNPlayer returns the player with the highest PairingNumber in the slice.
func highestTPNPlayer(players []*swisslib.PlayerState) *swisslib.PlayerState {
	if len(players) == 0 {
		return nil
	}
	best := players[0]
	for _, p := range players[1:] {
		if p.TPN > best.TPN {
			best = p
		}
	}
	return best
}
