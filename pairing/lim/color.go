// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import "github.com/gnutterts/chesspairing/pairing/swisslib"

// AllocateColor decides which player gets White and which gets Black
// for a specific pairing, implementing Art. 5 with median-aware tiebreaking.
//
// Parameters:
//   - a, b: the two players in the pairing
//   - roundNumber: 1-based round number (used to determine odd/even)
//   - isAboveMedian: true if the scoregroup is at or above the median (affects Art. 5.4)
//   - topSeedColor: override for round 1 top seed colour (nil = White default)
//
// Returns (whiteID, blackID).
func AllocateColor(a, b *swisslib.PlayerState, roundNumber int, isAboveMedian bool, topSeedColor *swisslib.Color) (string, string) {
	histA := filterPlayed(a.ColorHistory)
	histB := filterPlayed(b.ColorHistory)

	// Round 1: no history, alternate by TPN.
	if len(histA) == 0 && len(histB) == 0 {
		return round1Color(a, b, topSeedColor)
	}

	// Art. 5.3: If one player had the same colour in previous two rounds,
	// that player MUST get the alternating colour.
	mustA := mustAlternate(histA)
	mustB := mustAlternate(histB)

	if mustA != nil && mustB != nil {
		// Both must alternate. If they need different colours, grant both.
		// If same → problem (should have been caught by compatibility).
		if *mustA != *mustB {
			return grantColor(a, b, *mustA)
		}
		// Same colour needed — higher ranked gets it.
		return grantColor(a, b, *mustA)
	}
	if mustA != nil {
		return grantColor(a, b, *mustA)
	}
	if mustB != nil {
		opp := (*mustB).Opposite()
		return grantColor(a, b, opp)
	}

	isEven := roundNumber%2 == 0

	// Art. 5.2/5.6: Even round — equalising colour.
	if isEven {
		eqA := equalisingColor(histA)
		eqB := equalisingColor(histB)
		if eqA != nil && eqB != nil && *eqA != *eqB {
			return grantColor(a, b, *eqA)
		}
		if eqA != nil && eqB == nil {
			return grantColor(a, b, *eqA)
		}
		if eqB != nil && eqA == nil {
			opp := (*eqB).Opposite()
			return grantColor(a, b, opp)
		}
		// Both need same equalising colour → use history tiebreak (Art. 5.4).
		if eqA != nil {
			return historyTiebreak(a, b, *eqA, isAboveMedian)
		}
	}

	// Art. 5.5: Odd round — alternating colour.
	altA := alternatingColor(histA)
	altB := alternatingColor(histB)
	if altA != nil && altB != nil && *altA != *altB {
		return grantColor(a, b, *altA)
	}
	if altA != nil && altB == nil {
		return grantColor(a, b, *altA)
	}
	if altB != nil && altA == nil {
		opp := (*altB).Opposite()
		return grantColor(a, b, opp)
	}

	// Both want same colour → history tiebreak (Art. 5.4).
	if altA != nil {
		return historyTiebreak(a, b, *altA, isAboveMedian)
	}

	// No preference from either → default alternation by board.
	return round1Color(a, b, topSeedColor)
}

// round1Color assigns colours for round 1 or when no history exists.
// Per Art. 7.2: odd-numbered players in upper half get the drawn colour.
func round1Color(a, b *swisslib.PlayerState, topSeedColor *swisslib.Color) (string, string) {
	higher, lower := a, b
	if b.TPN < a.TPN {
		higher, lower = b, a
	}

	initialColor := swisslib.ColorWhite
	if topSeedColor != nil {
		initialColor = *topSeedColor
	}

	// Odd TPN in upper half gets initial colour.
	if higher.TPN%2 == 1 {
		return grantColor(higher, lower, initialColor)
	}
	return grantColor(higher, lower, initialColor.Opposite())
}

// mustAlternate returns the colour a player MUST receive if they've had
// the same colour in the previous two rounds (Art. 5.3). Returns nil if no constraint.
func mustAlternate(played []swisslib.Color) *swisslib.Color {
	n := len(played)
	if n < 2 {
		return nil
	}
	if played[n-1] == played[n-2] {
		opp := played[n-1].Opposite()
		return &opp
	}
	return nil
}

// equalisingColor returns the colour needed to equalise the count.
// Returns nil if already equal.
func equalisingColor(played []swisslib.Color) *swisslib.Color {
	w, b := 0, 0
	for _, c := range played {
		if c == swisslib.ColorWhite {
			w++
		} else {
			b++
		}
	}
	if w > b {
		c := swisslib.ColorBlack
		return &c
	}
	if b > w {
		c := swisslib.ColorWhite
		return &c
	}
	return nil
}

// alternatingColor returns the opposite of the last played colour.
// Returns nil if no history.
func alternatingColor(played []swisslib.Color) *swisslib.Color {
	if len(played) == 0 {
		return nil
	}
	opp := played[len(played)-1].Opposite()
	return &opp
}

// historyTiebreak implements Art. 5.4: when both players want the same colour
// and have identical recent history, the median position determines who wins.
//
// Above or at median: higher ranked (lower TPN) gets the desired colour.
// Below median: lower ranked (higher TPN) gets the desired colour.
func historyTiebreak(a, b *swisslib.PlayerState, desiredColorForWinner swisslib.Color, isAboveMedian bool) (string, string) {
	// First try to find a difference in colour history going backwards.
	histA := filterPlayed(a.ColorHistory)
	histB := filterPlayed(b.ColorHistory)
	ia := len(histA) - 1
	ib := len(histB) - 1

	for ia >= 0 && ib >= 0 {
		if histA[ia] != histB[ib] {
			// Found a difference. The player who had the same colour as desired
			// in that round should get the opposite (to alternate), and the other
			// gets the desired colour.
			if histA[ia] == desiredColorForWinner {
				// a had this colour before → b gets it now (alternation principle).
				return grantColor(b, a, desiredColorForWinner)
			}
			return grantColor(a, b, desiredColorForWinner)
		}
		ia--
		ib--
	}

	// Identical histories → use median tiebreak.
	if isAboveMedian {
		// Higher ranked (lower TPN) gets the colour.
		if a.TPN < b.TPN {
			return grantColor(a, b, desiredColorForWinner)
		}
		return grantColor(b, a, desiredColorForWinner)
	}
	// Below median: lower ranked (higher TPN) gets the colour.
	if a.TPN > b.TPN {
		return grantColor(a, b, desiredColorForWinner)
	}
	return grantColor(b, a, desiredColorForWinner)
}

// grantColor returns (whiteID, blackID) granting player p the given colour.
func grantColor(p, other *swisslib.PlayerState, c swisslib.Color) (string, string) {
	if c == swisslib.ColorWhite {
		return p.ID, other.ID
	}
	return other.ID, p.ID
}
