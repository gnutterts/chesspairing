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
func ClassifyFloater(p *swisslib.PlayerState, alreadyFloated bool, adjacentPlayers []*swisslib.PlayerState, forbidden map[[2]string]bool) FloaterType {
	hasCompatible := false
	for _, adj := range adjacentPlayers {
		if IsCompatible(p, adj, forbidden) {
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

// SelectDownFloater selects the player to float down from a scoregroup per Art. 3.2-3.4.
//
// Rules (Art. 3.2):
//  1. Select to equalise due colours in the remaining group (Art. 3.2.2)
//  2. If equal, lowest TPN when pairing downward (Art. 3.2.4)
//  3. Must have compatible opponent in adjacent group (Art. 3.3)
//  4. Minimise floater disadvantage type (Art. 3.9.2)
//
// Returns nil if no valid floater can be selected.
func SelectDownFloater(players []*swisslib.PlayerState, adjacent []*swisslib.PlayerState, forbidden map[[2]string]bool) *swisslib.PlayerState {
	if len(players) == 0 {
		return nil
	}

	// Count players due each colour.
	dueWhite, dueBlack := countDueColors(players)

	// Determine which due-colour group should provide the floater.
	var preferDue *swisslib.Color
	if dueWhite > dueBlack {
		w := swisslib.ColorWhite
		preferDue = &w
	} else if dueBlack > dueWhite {
		b := swisslib.ColorBlack
		preferDue = &b
	}

	// Build candidates with their floater types.
	type candidate struct {
		player     *swisslib.PlayerState
		floaterTyp FloaterType
		dueColor   *swisslib.Color
	}
	var candidates []candidate
	for _, p := range players {
		ft := ClassifyFloater(p, false, adjacent, forbidden)
		pref := swisslib.ComputeColorPreference(p.ColorHistory)
		candidates = append(candidates, candidate{
			player:     p,
			floaterTyp: ft,
			dueColor:   pref.Color,
		})
	}

	// Sort candidates by:
	// 1. Best floater type (highest = D, least disadvantage) first
	// 2. Colour match (matches preferDue) first
	// 3. Lowest TPN first (Art. 3.2.4: lowest numbered when pairing downward)
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].floaterTyp != candidates[j].floaterTyp {
			return candidates[i].floaterTyp > candidates[j].floaterTyp
		}
		matchI := preferDue != nil && candidates[i].dueColor != nil && *candidates[i].dueColor == *preferDue
		matchJ := preferDue != nil && candidates[j].dueColor != nil && *candidates[j].dueColor == *preferDue
		if matchI != matchJ {
			return matchI
		}
		return candidates[i].player.TPN < candidates[j].player.TPN
	})

	// Return the best candidate that has a compatible opponent in adjacent.
	for _, c := range candidates {
		for _, adj := range adjacent {
			if IsCompatible(c.player, adj, forbidden) {
				return c.player
			}
		}
	}

	// No compatible opponent — return the first candidate (caller handles further floating).
	if len(candidates) > 0 {
		return candidates[0].player
	}
	return nil
}

// SelectUpFloater selects the player to float up from a scoregroup per Art. 3.2, 3.4.
//
// When pairing upwards, the highest numbered player (highest TPN) is chosen.
func SelectUpFloater(players []*swisslib.PlayerState, adjacent []*swisslib.PlayerState, forbidden map[[2]string]bool) *swisslib.PlayerState {
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
		player     *swisslib.PlayerState
		floaterTyp FloaterType
		dueColor   *swisslib.Color
	}
	var candidates []candidate
	for _, p := range players {
		ft := ClassifyFloater(p, false, adjacent, forbidden)
		pref := swisslib.ComputeColorPreference(p.ColorHistory)
		candidates = append(candidates, candidate{
			player:     p,
			floaterTyp: ft,
			dueColor:   pref.Color,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].floaterTyp != candidates[j].floaterTyp {
			return candidates[i].floaterTyp > candidates[j].floaterTyp
		}
		matchI := preferDue != nil && candidates[i].dueColor != nil && *candidates[i].dueColor == *preferDue
		matchJ := preferDue != nil && candidates[j].dueColor != nil && *candidates[j].dueColor == *preferDue
		if matchI != matchJ {
			return matchI
		}
		// Highest TPN first when pairing upwards (Art. 3.2.4).
		return candidates[i].player.TPN > candidates[j].player.TPN
	})

	for _, c := range candidates {
		for _, adj := range adjacent {
			if IsCompatible(c.player, adj, forbidden) {
				return c.player
			}
		}
	}

	if len(candidates) > 0 {
		return candidates[0].player
	}
	return nil
}

// countDueColors counts how many players are due White vs Black.
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
