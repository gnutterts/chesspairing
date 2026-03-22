package team

import "github.com/gnutterts/chesspairing/pairing/lexswiss"

// AllocateColor decides which team gets White and which gets Black,
// implementing the 9-step colour allocation algorithm of Art. 4.
//
// Parameters:
//   - a, b: the two teams in the pairing
//   - prefType: colour preference type (A, B, or None)
//   - isLastRound: true if this is the last round (affects Type B mild prefs)
//   - initialColor: the initial-colour from drawing of lots (Art. 4.1)
//   - secondaryScores: secondary score map (Art. 4.2.2), nil if not used
//
// Returns (whiteID, blackID).
func AllocateColor(a, b *lexswiss.ParticipantState, prefType ColorPrefType, isLastRound bool, initialColor *string, secondaryScores map[string]float64) (string, string) {
	// Determine the first-team (Art. 4.2).
	first, second := determineFirstTeam(a, b, secondaryScores)

	prefFirst := ComputeColorPreference(first, prefType, isLastRound)
	prefSecond := ComputeColorPreference(second, prefType, isLastRound)

	// Art. 4.3.1: Both teams have yet to play a match.
	if len(filterPlayed(first.ColorHistory)) == 0 && len(filterPlayed(second.ColorHistory)) == 0 {
		ic := lexswiss.ColorWhite
		if initialColor != nil && *initialColor == "black" {
			ic = lexswiss.ColorBlack
		}
		if first.TPN%2 == 1 {
			// Odd TPN → initial-colour.
			return assignColor(first, second, ic)
		}
		// Even TPN → opposite of initial-colour.
		return assignColor(first, second, ic.Opposite())
	}

	// Art. 4.3.2: Only one team has a colour preference → grant it.
	if prefFirst != ColorPrefNone && prefSecond == ColorPrefNone {
		return grantPreference(first, second, prefFirst)
	}
	if prefSecond != ColorPrefNone && prefFirst == ColorPrefNone {
		return grantPreference(second, first, prefSecond)
	}

	// Art. 4.3.3: Two teams have opposite colour preferences → grant both.
	if prefFirst.IsWhite() && prefSecond.IsBlack() {
		return first.ID, second.ID
	}
	if prefFirst.IsBlack() && prefSecond.IsWhite() {
		return second.ID, first.ID
	}

	// Art. 4.3.4 (Type B only): Only one team has a strong preference → grant it.
	if prefType == ColorPrefTypeB {
		if prefFirst.IsStrong() && !prefSecond.IsStrong() {
			return grantPreference(first, second, prefFirst)
		}
		if prefSecond.IsStrong() && !prefFirst.IsStrong() {
			return grantPreference(second, first, prefSecond)
		}
	}

	// Art. 4.3.5: Give White to the team with the lower colour difference.
	cdFirst := colorDifference(first.ColorHistory)
	cdSecond := colorDifference(second.ColorHistory)
	if cdFirst < cdSecond {
		return first.ID, second.ID // lower CD gets White
	}
	if cdSecond < cdFirst {
		return second.ID, first.ID
	}

	// Art. 4.3.6: Alternate from most recent time one had White and other Black.
	if alternateID := findAlternation(first, second); alternateID != "" {
		// alternateID is the team that should get White (the one who had Black at the alternation point).
		if alternateID == first.ID {
			return first.ID, second.ID
		}
		return second.ID, first.ID
	}

	// Art. 4.3.7: Grant the colour preference of the first-team.
	if prefFirst != ColorPrefNone {
		return grantPreference(first, second, prefFirst)
	}

	// Art. 4.3.8: Alternate the colour of the first-team from its last played round.
	lastFirst := lastPlayedColor(first.ColorHistory)
	if lastFirst != lexswiss.ColorNone {
		return assignColor(first, second, lastFirst.Opposite())
	}

	// Art. 4.3.9: Alternate the colour of the other team from its last played round.
	lastSecond := lastPlayedColor(second.ColorHistory)
	if lastSecond != lexswiss.ColorNone {
		return assignColor(second, first, lastSecond.Opposite())
	}

	// Ultimate fallback: first-team gets White.
	return first.ID, second.ID
}

// determineFirstTeam returns (first, second) per Art. 4.2:
// higher primary score, then higher secondary score, then smaller TPN.
func determineFirstTeam(a, b *lexswiss.ParticipantState, secondaryScores map[string]float64) (*lexswiss.ParticipantState, *lexswiss.ParticipantState) {
	// Art. 4.2.1: higher primary score.
	if a.Score > b.Score {
		return a, b
	}
	if b.Score > a.Score {
		return b, a
	}

	// Art. 4.2.2: higher secondary score (if available).
	if secondaryScores != nil {
		secA := secondaryScores[a.ID]
		secB := secondaryScores[b.ID]
		if secA > secB {
			return a, b
		}
		if secB > secA {
			return b, a
		}
	}

	// Art. 4.2.3: smaller TPN.
	if a.TPN < b.TPN {
		return a, b
	}
	return b, a
}

// grantPreference assigns colours to grant the preferred team's colour preference.
func grantPreference(preferred, other *lexswiss.ParticipantState, pref ColorPref) (string, string) {
	if pref.IsWhite() {
		return preferred.ID, other.ID
	}
	return other.ID, preferred.ID
}

// assignColor assigns White to the team that should get the given colour.
// If color is White, team gets White; if Black, team gets Black.
func assignColor(team, other *lexswiss.ParticipantState, color lexswiss.Color) (string, string) {
	if color == lexswiss.ColorWhite {
		return team.ID, other.ID
	}
	return other.ID, team.ID
}

// findAlternation implements Art. 4.3.6: find the most recent round where
// one team had White and the other had Black, then alternate.
// Returns the ID of the team that should get White (the one that had Black
// at the alternation point), or "" if no such point exists.
func findAlternation(a, b *lexswiss.ParticipantState) string {
	histA := a.ColorHistory
	histB := b.ColorHistory
	minLen := len(histA)
	if len(histB) < minLen {
		minLen = len(histB)
	}

	// Walk backward from the most recent round.
	for i := minLen - 1; i >= 0; i-- {
		ca := histA[i]
		cb := histB[i]
		// Skip rounds where either had no game (None).
		if ca == lexswiss.ColorNone || cb == lexswiss.ColorNone {
			continue
		}
		if ca == lexswiss.ColorWhite && cb == lexswiss.ColorBlack {
			// a had White → alternate: a gets Black, b gets White.
			return b.ID
		}
		if ca == lexswiss.ColorBlack && cb == lexswiss.ColorWhite {
			// a had Black → alternate: a gets White, b gets Black.
			return a.ID
		}
		// Same colour → continue looking further back.
	}

	return ""
}

// lastPlayedColor returns the most recent non-None colour from history.
func lastPlayedColor(history []lexswiss.Color) lexswiss.Color {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i] != lexswiss.ColorNone {
			return history[i]
		}
	}
	return lexswiss.ColorNone
}
