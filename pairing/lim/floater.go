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
// The floatedIDs set tracks players that have already been floated from a
// previous scoregroup; these are classified as "already floated" (Type A/B)
// per Art. 3.9.
//
// Returns nil if no valid floater can be selected.
func SelectDownFloater(players []*swisslib.PlayerState, adjacent []*swisslib.PlayerState, forbidden map[[2]string]bool, floatedIDs map[string]bool, maxiTournament bool) *swisslib.PlayerState {
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
		ft := ClassifyFloater(p, floatedIDs[p.ID], adjacent, forbidden)
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

	// Find the best candidate that has a compatible opponent in adjacent.
	var selected *swisslib.PlayerState
	for _, c := range candidates {
		for _, adj := range adjacent {
			if IsCompatible(c.player, adj, forbidden) {
				selected = c.player
				break
			}
		}
		if selected != nil {
			break
		}
	}

	// Fallback: no compatible opponent.
	if selected == nil && len(candidates) > 0 {
		selected = candidates[0].player
	}

	// Maxi-tournament rating cap (Art. 3.2.3): unconditional override.
	// The spec mandates that "the lowest numbered player IS chosen" when the
	// rating difference exceeds 100 points. This takes absolute priority over
	// floater type (Art. 3.9) and compatibility — the caller handles the case
	// where the overridden floater has no compatible opponent by floating further.
	if maxiTournament && selected != nil {
		ref := lowestTPNPlayer(players)
		if ref != nil && selected.ID != ref.ID {
			if absInt(selected.Rating-ref.Rating) > 100 {
				selected = ref
			}
		}
	}

	return selected
}

// SelectUpFloater selects the player to float up from a scoregroup per Art. 3.2, 3.4.
//
// When pairing upwards, the highest numbered player (highest TPN) is chosen.
// The floatedIDs set tracks players already floated from a previous scoregroup.
func SelectUpFloater(players []*swisslib.PlayerState, adjacent []*swisslib.PlayerState, forbidden map[[2]string]bool, floatedIDs map[string]bool, maxiTournament bool) *swisslib.PlayerState {
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
		ft := ClassifyFloater(p, floatedIDs[p.ID], adjacent, forbidden)
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

	var selected *swisslib.PlayerState
	for _, c := range candidates {
		for _, adj := range adjacent {
			if IsCompatible(c.player, adj, forbidden) {
				selected = c.player
				break
			}
		}
		if selected != nil {
			break
		}
	}

	if selected == nil && len(candidates) > 0 {
		selected = candidates[0].player
	}

	// Maxi-tournament rating cap (Art. 3.2.3): unconditional override.
	// The spec mandates that "the highest numbered player IS chosen" when the
	// rating difference exceeds 100 points. This takes absolute priority over
	// floater type (Art. 3.9) and compatibility — the caller handles the case
	// where the overridden floater has no compatible opponent by floating further.
	if maxiTournament && selected != nil {
		ref := highestTPNPlayer(players)
		if ref != nil && selected.ID != ref.ID {
			if absInt(selected.Rating-ref.Rating) > 100 {
				selected = ref
			}
		}
	}

	return selected
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

// absInt returns the absolute value of x.
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// lowestTPNPlayer returns the player with the lowest TPN in the slice.
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

// highestTPNPlayer returns the player with the highest TPN in the slice.
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
