package lim

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// FloatDirection indicates whether a player floated down or up.
type FloatDirection int

const (
	FloatDown FloatDirection = iota
	FloatUp
)

// FloaterEntry tracks a floater with metadata about where they came from.
type FloaterEntry struct {
	Player      *swisslib.PlayerState
	Direction   FloatDirection
	SourceScore float64 // score of the scoregroup they floated from
}

// PairFloaters pairs incoming floaters with targets from the current scoregroup
// per Art. 3.6-3.8, before exchange matching.
//
// Parameters:
//   - floaters: players that floated into this scoregroup from other groups
//   - targets: the native players of this scoregroup (the floater opponents)
//   - isUpperHalf: true if this scoregroup is at or above the median (affects DF/UF priority)
//   - isMaxi: true if maxi-tournament mode (100-point rating constraint on exchanges)
//   - forbidden: forbidden pairs map
//
// Returns:
//   - pairs: floater-target pairs formed
//   - remainingFloaters: floaters that could not be paired (will float further)
//   - remainingTargets: targets not consumed by floater pairing (go into ExchangeMatch)
func PairFloaters(floaters []FloaterEntry, targets []*swisslib.PlayerState, isUpperHalf bool, isMaxi bool, forbidden map[[2]string]bool) (pairs [][2]*swisslib.PlayerState, remainingFloaters []FloaterEntry, remainingTargets []*swisslib.PlayerState) {
	if len(floaters) == 0 {
		return nil, nil, targets
	}

	// Sort floaters per Art. 3.6.3 / 3.7.3 ordering.
	sorted := sortFloaters(floaters, isUpperHalf)

	// Track which targets are still available.
	available := make([]*swisslib.PlayerState, len(targets))
	copy(available, targets)

	for _, fe := range sorted {
		opponent := selectFloaterOpponent(fe.Player, fe.Direction, available, isMaxi, forbidden)
		if opponent != nil {
			pairs = append(pairs, [2]*swisslib.PlayerState{fe.Player, opponent})
			available = removePtrs(available, opponent)
		} else {
			remainingFloaters = append(remainingFloaters, fe)
		}
	}

	remainingTargets = available
	return
}

// sortFloaters orders floaters per Art. 3.6.3 and 3.7.3.
//
// In upper half/median groups (isUpperHalf=true): DF first, then UF (Art. 3.6.3).
// In lower half groups (isUpperHalf=false): UF first, then DF (Art. 3.7.3).
//
// Within each direction group:
//   - Down-floaters: highest source score first (3.6.2), then highest TPN first (3.6.1)
//   - Up-floaters: lowest source score first (3.7.2), then lowest TPN first (3.7.1)
func sortFloaters(floaters []FloaterEntry, isUpperHalf bool) []FloaterEntry {
	sorted := make([]FloaterEntry, len(floaters))
	copy(sorted, floaters)

	sort.SliceStable(sorted, func(i, j int) bool {
		fi, fj := sorted[i], sorted[j]

		// Primary: direction priority depends on half.
		if fi.Direction != fj.Direction {
			if isUpperHalf {
				// Upper half: DF (FloatDown=0) before UF (FloatUp=1).
				return fi.Direction < fj.Direction
			}
			// Lower half: UF (FloatUp=1) before DF (FloatDown=0).
			return fi.Direction > fj.Direction
		}

		// Same direction: sort within group.
		if fi.Direction == FloatDown {
			// DF: highest source score first (Art. 3.6.2).
			if fi.SourceScore != fj.SourceScore {
				return fi.SourceScore > fj.SourceScore
			}
			// Then highest TPN first (Art. 3.6.1).
			return fi.Player.TPN > fj.Player.TPN
		}

		// UF: lowest source score first (Art. 3.7.2).
		if fi.SourceScore != fj.SourceScore {
			return fi.SourceScore < fj.SourceScore
		}
		// Then lowest TPN first (Art. 3.7.1).
		return fi.Player.TPN < fj.Player.TPN
	})

	return sorted
}

// selectFloaterOpponent selects the opponent for a floater per Art. 3.8.
//
// Down-floater: paired with highest TPN available player due the alternate colour.
// Up-floater: paired with lowest TPN available player due the alternate colour.
//
// In Maxi-tournaments (Art. 3.8 + Art. 5.7): an opponent is only eligible if
// the rating difference between the floater and the candidate is 100 points or
// less. This constraint is applied before any colour or compatibility checks.
//
// "Due the alternate colour" means the target's colour preference is the
// opposite of the floater's colour preference.
func selectFloaterOpponent(floater *swisslib.PlayerState, dir FloatDirection, available []*swisslib.PlayerState, isMaxi bool, forbidden map[[2]string]bool) *swisslib.PlayerState {
	if len(available) == 0 {
		return nil
	}

	// Determine floater's colour preference.
	floaterPref := swisslib.ComputeColorPreference(floater.ColorHistory)

	// Sort candidates by TPN in direction order.
	candidates := make([]*swisslib.PlayerState, len(available))
	copy(candidates, available)

	if dir == FloatDown {
		// Highest TPN first (Art. 3.8: "highest numbered player available").
		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].TPN > candidates[j].TPN
		})
	} else {
		// Lowest TPN first (Art. 3.8: "lowest numbered player available").
		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].TPN < candidates[j].TPN
		})
	}

	// First pass: find compatible candidate who is due the alternate colour.
	for _, c := range candidates {
		if isMaxi && absInt(floater.Rating-c.Rating) > 100 {
			continue
		}
		if !IsCompatible(floater, c, forbidden) {
			continue
		}
		cPref := swisslib.ComputeColorPreference(c.ColorHistory)
		if isDueAlternateColour(floaterPref.Color, cPref.Color) {
			return c
		}
	}

	// Second pass: find any compatible candidate (colour preference not met).
	for _, c := range candidates {
		if isMaxi && absInt(floater.Rating-c.Rating) > 100 {
			continue
		}
		if IsCompatible(floater, c, forbidden) {
			return c
		}
	}

	return nil
}

// isDueAlternateColour returns true if the target's due colour is the
// opposite of the floater's due colour (i.e., they want different colours).
// If either has no preference (nil), it counts as compatible.
func isDueAlternateColour(floaterDue, targetDue *swisslib.Color) bool {
	if floaterDue == nil || targetDue == nil {
		return true
	}
	return *floaterDue != *targetDue
}
