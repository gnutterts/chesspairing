package swisslib

// ColorPreference holds the computed color preference for a player.
//
// This matches bbpPairings' three-tier system (tournament.cpp computePlayerData):
//   - AbsolutePreference: colorImbalance > 1 OR 2+ consecutive same color
//   - StrongPreference: colorImbalance > 0 AND NOT absolute
//   - Otherwise: mild preference (alternation) or no preference
//
// The Color field holds the preferred color direction regardless of strength.
// ColorImbalance and HasConsecutive are used by AllocateColor for tiebreaking.
type ColorPreference struct {
	Color              *Color  // preferred color (nil = no preference)
	ColorImbalance     int     // abs(whites - blacks), always >= 0
	AbsolutePreference bool    // must play this color (imbalance > 1 OR 2+ consecutive)
	StrongPreference   bool    // should play this color (imbalance > 0, NOT absolute)
	PlayedColors       []Color // played color history (no byes), for findFirstColorDifference
}

// Strength returns a numeric strength value for comparison.
// 3 = absolute, 2 = strong, 1 = mild, 0 = none.
func (cp ColorPreference) Strength() int {
	if cp.AbsolutePreference {
		return 3
	}
	if cp.StrongPreference {
		return 2
	}
	if cp.Color != nil {
		return 1
	}
	return 0
}

// PreferredColor returns the color this player prefers, or nil if none.
func (cp ColorPreference) PreferredColor() *Color {
	return cp.Color
}

// ComputeColorPreference derives a player's color preference from their
// color history, matching bbpPairings' computePlayerData exactly.
//
// bbpPairings logic (tournament.cpp lines 43-93):
//
//	colorPreference =
//	  colorImbalance > 1 ? lowerColor                    // absolute (imbalance)
//	  : consecutiveCount > 1 ? invert(repeatedColor)     // absolute (consecutive)
//	  : colorImbalance > 0 ? lowerColor                  // strong
//	  : consecutiveCount > 0 ? invert(repeatedColor)     // mild
//	  : COLOR_NONE
//
//	absoluteColorPreference = colorImbalance > 1 || repeatedColor != NONE
//	  (repeatedColor is cleared to NONE when consecutiveCount <= 1)
//	  So effectively: colorImbalance > 1 || consecutiveCount > 1
//
//	strongColorPreference = !absoluteColorPreference && colorImbalance > 0
//
// ColorNone (byes/absences) is skipped for all calculations.
func ComputeColorPreference(history []Color) ColorPreference {
	var pref ColorPreference

	// Filter out ColorNone (byes/absences).
	played := filterPlayed(history)
	pref.PlayedColors = played
	if len(played) == 0 {
		return pref
	}

	// Count colors.
	whites, blacks := countColors(played)

	// Compute colorImbalance (always >= 0) and lowerColor.
	var lowerColor Color
	if whites > blacks {
		pref.ColorImbalance = whites - blacks
		lowerColor = ColorBlack
	} else {
		pref.ColorImbalance = blacks - whites
		lowerColor = ColorWhite
	}

	// Track consecutive count (same as bbpPairings).
	var consecutiveCount int
	var repeatedColor Color
	for _, c := range played {
		if consecutiveCount == 0 || c != repeatedColor {
			consecutiveCount = 1
		} else {
			consecutiveCount++
		}
		repeatedColor = c
	}

	// Compute colorPreference BEFORE clearing repeatedColor.
	// bbpPairings lines 80-85 use the uncleaned repeatedColor for the
	// consecutiveCount > 0 case (mild preference = invert of last played color).
	switch {
	case pref.ColorImbalance > 1:
		pref.Color = &lowerColor
	case consecutiveCount > 1:
		opp := repeatedColor.Opposite()
		pref.Color = &opp
	case pref.ColorImbalance > 0:
		pref.Color = &lowerColor
	case consecutiveCount > 0:
		opp := repeatedColor.Opposite()
		pref.Color = &opp
	default:
		// No preference.
	}

	// Clear repeatedColor if no consecutive run (matches bbpPairings line 86-88).
	// This MUST happen AFTER computing colorPreference but BEFORE computing
	// absoluteColorPreference and strongColorPreference.
	if consecutiveCount <= 1 {
		repeatedColor = ColorNone
	}

	// Compute absoluteColorPreference (bbpPairings lines 215-218).
	// absoluteColorImbalance() = colorImbalance > 1
	// absoluteColorPreference() = absoluteColorImbalance() || repeatedColor != COLOR_NONE
	// After clearing above, repeatedColor is COLOR_NONE when consecutiveCount <= 1.
	pref.AbsolutePreference = pref.ColorImbalance > 1 || repeatedColor != ColorNone

	// Compute strongColorPreference (bbpPairings line 90-91).
	// = !absoluteColorPreference() && colorImbalance > 0
	pref.StrongPreference = !pref.AbsolutePreference && pref.ColorImbalance > 0

	return pref
}

// AllocateColor decides which player gets White and which gets Black
// for a specific pairing, implementing bbpPairings' choosePlayerNeutralColor
// and choosePlayerColor (dutch.cpp lines 488-516, common.cpp lines 250-315).
//
// topSeedColor controls round 1 board alternation when no player has color
// history. nil = default (higher-ranked gets White on odd boards).
// Non-nil overrides the color assigned to the higher-ranked player on board 1;
// subsequent boards alternate from there.
//
// Priority (from choosePlayerNeutralColor):
//  1. Compatible preferences → grant first player's preference
//  2. One absolute, stronger imbalance or opponent not absolute → grant
//  3. One strong, other not → grant strong player's preference
//  4. findFirstColorDifference → swap from most recent differing round
//
// Fallback (from choosePlayerColor, when neutral returns COLOR_NONE):
//  5. If both have preferences but same color → higher-ranked gets preferred
//  6. If neither has preference → alternate by board (FIDE C.04.3 A.6.e)
//
// Returns (whiteID, blackID).
func AllocateColor(a, b *PlayerState, topScorerRules bool, boardNumber int, topSeedColor *Color) (string, string) {
	prefA := ComputeColorPreference(a.ColorHistory)
	prefB := ComputeColorPreference(b.ColorHistory)

	colorA := prefA.PreferredColor()
	colorB := prefB.PreferredColor()

	// Helper: return player a with given color.
	grantA := func(c Color) (string, string) {
		if c == ColorWhite {
			return a.ID, b.ID
		}
		return b.ID, a.ID
	}
	grantB := func(c Color) (string, string) {
		if c == ColorWhite {
			return b.ID, a.ID
		}
		return a.ID, b.ID
	}

	// Step 1: Check if preferences are compatible.
	// Compatible = at least one has no preference, or they want different colors.
	compatible := colorPreferencesAreCompatible(colorA, colorB)

	if compatible {
		if colorA != nil {
			return grantA(*colorA)
		}
		if colorB != nil {
			// Grant opponent's (b's) preference by inverting.
			return grantB(*colorB)
		}
		// Both have no preference → fall through to choosePlayerColor fallback.
	} else {
		// Incompatible preferences (both want same color).

		// Step 2: Absolute preference wins.
		if prefA.AbsolutePreference &&
			(prefA.ColorImbalance > prefB.ColorImbalance || !prefB.AbsolutePreference) {
			return grantA(*colorA)
		}
		if prefB.AbsolutePreference &&
			(prefB.ColorImbalance > prefA.ColorImbalance || !prefA.AbsolutePreference) {
			return grantB(*colorB)
		}

		// Step 3: Strong preference beats non-strong.
		if prefA.StrongPreference && !prefB.StrongPreference {
			return grantA(*colorA)
		}
		if prefB.StrongPreference && !prefA.StrongPreference {
			return grantB(*colorB)
		}

		// Step 4: findFirstColorDifference.
		colorForA := findFirstColorDifference(prefA.PlayedColors, prefB.PlayedColors)
		if colorForA != nil {
			return grantA(*colorForA)
		}
		// Both had identical color patterns → fall through.
	}

	// choosePlayerColor fallback (dutch.cpp lines 488-516).
	// choosePlayerNeutralColor returned COLOR_NONE.
	//
	// If both have color preferences (same color, equal strength):
	// Higher-ranked player (by acceleratedScoreRankCompare) gets preference.
	if colorA != nil {
		// Both want the same color. Higher-ranked (lower TPN) gets it.
		if a.TPN < b.TPN {
			return grantA(*colorA)
		}
		return grantB(*colorB)
	}

	// No preferences at all: alternate by board number per FIDE C.04.3 A.6.e.
	// topSeedColor controls which color the higher-ranked player gets on board 1.
	// Default (nil or White): odd boards → higher-ranked gets White.
	// Black: odd boards → higher-ranked gets Black.
	higherRanked, lowerRanked := a, b
	if b.TPN < a.TPN {
		higherRanked, lowerRanked = b, a
	}

	// Determine if the board alternation pattern is inverted.
	invertPattern := topSeedColor != nil && *topSeedColor == ColorBlack

	if (boardNumber%2 == 1) != invertPattern {
		// Odd board (normal) or even board (inverted): higher-ranked gets White.
		return higherRanked.ID, lowerRanked.ID
	}
	// Even board (normal) or odd board (inverted): lower-ranked gets White.
	return lowerRanked.ID, higherRanked.ID
}

// colorPreferencesAreCompatible returns true when the two preferences can
// both be satisfied (at least one is nil, or they want different colors).
// Matches bbpPairings colorPreferencesAreCompatible.
func colorPreferencesAreCompatible(a, b *Color) bool {
	if a == nil || b == nil {
		return true
	}
	return *a != *b
}

// findFirstColorDifference walks backwards through both players' played-color
// histories simultaneously, skipping rounds where they played the same color,
// and returns the color for player A from the first round where they differed.
// Returns nil if no difference found (identical color patterns).
//
// Matches bbpPairings findFirstColorDifference (common.cpp lines 216-242).
// When a difference is found, bbpPairings returns opponentColor to the caller
// (line 309), which means: give player the color the OPPONENT had in that
// round (i.e., swap). We return the swap color for player A directly.
func findFirstColorDifference(playedA, playedB []Color) *Color {
	ia := len(playedA) - 1
	ib := len(playedB) - 1

	for ia >= 0 && ib >= 0 {
		if playedA[ia] != playedB[ib] {
			// Found a difference. bbpPairings returns opponentColor (line 309),
			// which is the color of player1 at this position. The logic is:
			// return the opposite of what player A had → give player A the
			// color that player B had (swap).
			c := playedB[ib] // opponent's color at this round
			return &c
		}
		ia--
		ib--
	}
	return nil
}

// filterPlayed returns only non-None colors from history.
func filterPlayed(history []Color) []Color {
	var played []Color
	for _, c := range history {
		if c != ColorNone {
			played = append(played, c)
		}
	}
	return played
}

// countColors counts White and Black in a played-color slice.
func countColors(played []Color) (whites, blacks int) {
	for _, c := range played {
		switch c {
		case ColorWhite:
			whites++
		case ColorBlack:
			blacks++
		}
	}
	return
}
