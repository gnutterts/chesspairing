// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import "github.com/gnutterts/chesspairing/pairing/lexswiss"

// AllocateColor decides which participant gets White in Game 1 of the
// match and which gets Black, implementing Art. 4 of the Double-Swiss system.
//
// In Double-Swiss, "colour" means who gets White in Game 1 of the 2-game match.
// The other participant gets White in Game 2 (colours alternate within the match).
//
// Priority:
//  1. Hard constraint: no participant plays Game 1 as the same colour 3 times in a row
//  2. Equalise: participant with more white Game-1s gets Black in Game 1
//  3. Alternate: participant who had White in Game 1 last round gets Black
//  4. Rank tiebreak: higher ranked (lower TPN) gets White in Game 1
//  5. Round 1: board alternation (odd board → higher ranked White, even board → Black)
//
// Parameters:
//   - a, b: the two participants in the pairing
//   - roundNumber: 1-based round number
//   - boardNumber: 1-based board number (for round 1 alternation)
//   - topSeedColor: override for round 1 top seed colour (nil or "auto" = default)
//
// Returns (whiteID, blackID) indicating who plays White in Game 1.
func AllocateColor(a, b *lexswiss.ParticipantState, roundNumber, boardNumber int, topSeedColor *string) (string, string) {
	histA := filterPlayed(a.ColorHistory)
	histB := filterPlayed(b.ColorHistory)

	// Step 1: Hard constraint — no 3 consecutive same colour in Game 1.
	mustNotWhiteA := len(histA) >= 2 && histA[len(histA)-1] == lexswiss.ColorWhite && histA[len(histA)-2] == lexswiss.ColorWhite
	mustNotWhiteB := len(histB) >= 2 && histB[len(histB)-1] == lexswiss.ColorWhite && histB[len(histB)-2] == lexswiss.ColorWhite
	mustNotBlackA := len(histA) >= 2 && histA[len(histA)-1] == lexswiss.ColorBlack && histA[len(histA)-2] == lexswiss.ColorBlack
	mustNotBlackB := len(histB) >= 2 && histB[len(histB)-1] == lexswiss.ColorBlack && histB[len(histB)-2] == lexswiss.ColorBlack

	if mustNotWhiteA && !mustNotBlackB {
		return b.ID, a.ID // a must NOT get White → a=Black
	}
	if mustNotWhiteB && !mustNotBlackA {
		return a.ID, b.ID // b must NOT get White → b=Black
	}
	if mustNotBlackA && !mustNotWhiteB {
		return a.ID, b.ID // a must NOT get Black → a=White
	}
	if mustNotBlackB && !mustNotWhiteA {
		return b.ID, a.ID // b must NOT get Black → b=White
	}

	// Step 2: Equalise — participant with more whites gets Black.
	whitesA := countColor(histA, lexswiss.ColorWhite)
	whitesB := countColor(histB, lexswiss.ColorWhite)

	if whitesA > whitesB {
		return b.ID, a.ID // a has more whites → a gets Black
	}
	if whitesB > whitesA {
		return a.ID, b.ID // b has more whites → b gets Black
	}

	// Step 3: Alternate — participant who had White last gets Black.
	if len(histA) > 0 && len(histB) > 0 {
		lastA := histA[len(histA)-1]
		lastB := histB[len(histB)-1]
		if lastA != lastB {
			if lastA == lexswiss.ColorWhite {
				return b.ID, a.ID // a had White → a gets Black
			}
			return a.ID, b.ID // b had White → b gets Black
		}
		// Both had same colour last → fall through to rank tiebreak.
	}

	// Step 4: Round 1 or no history — board alternation.
	if len(histA) == 0 && len(histB) == 0 {
		return round1Color(a, b, boardNumber, topSeedColor)
	}

	// Step 5: Rank tiebreak — higher ranked (lower TPN) gets White.
	if a.TPN < b.TPN {
		return a.ID, b.ID
	}
	return b.ID, a.ID
}

// round1Color assigns colours for round 1 using board alternation.
// Odd boards: higher ranked gets White (or reversed by topSeedColor).
func round1Color(a, b *lexswiss.ParticipantState, boardNumber int, topSeedColor *string) (string, string) {
	higherRanked, lowerRanked := a, b
	if b.TPN < a.TPN {
		higherRanked, lowerRanked = b, a
	}

	invertPattern := topSeedColor != nil && *topSeedColor == "black"

	if (boardNumber%2 == 1) != invertPattern {
		// Odd board (normal) or even board (inverted): higher ranked gets White.
		return higherRanked.ID, lowerRanked.ID
	}
	return lowerRanked.ID, higherRanked.ID
}

// filterPlayed returns only non-None colours from history.
func filterPlayed(history []lexswiss.Color) []lexswiss.Color {
	var played []lexswiss.Color
	for _, c := range history {
		if c != lexswiss.ColorNone {
			played = append(played, c)
		}
	}
	return played
}

// countColor counts occurrences of a specific colour in a played-colour slice.
func countColor(played []lexswiss.Color, target lexswiss.Color) int {
	count := 0
	for _, c := range played {
		if c == target {
			count++
		}
	}
	return count
}
