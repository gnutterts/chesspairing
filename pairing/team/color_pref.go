package team

import "github.com/gnutterts/chesspairing/pairing/lexswiss"

// ColorPrefType identifies which colour preference rules to use.
type ColorPrefType int

const (
	ColorPrefTypeA    ColorPrefType = iota // Simple: preference or none (default)
	ColorPrefTypeB                         // Strong + mild + none
	ColorPrefTypeNone                      // No colour preferences
)

// String returns the type name for debugging.
func (t ColorPrefType) String() string {
	switch t {
	case ColorPrefTypeA:
		return "TypeA"
	case ColorPrefTypeB:
		return "TypeB"
	case ColorPrefTypeNone:
		return "None"
	default:
		return "Unknown"
	}
}

// ColorPref represents a team's colour preference strength and direction.
type ColorPref int

const (
	ColorPrefNone        ColorPref = iota // No preference
	ColorPrefWhite                        // Type A: simple preference for White
	ColorPrefBlack                        // Type A: simple preference for Black
	ColorPrefStrongWhite                  // Type B: strong preference for White
	ColorPrefStrongBlack                  // Type B: strong preference for Black
	ColorPrefMildWhite                    // Type B: mild preference for White
	ColorPrefMildBlack                    // Type B: mild preference for Black
)

// String returns the preference name for debugging.
func (p ColorPref) String() string {
	switch p {
	case ColorPrefNone:
		return "None"
	case ColorPrefWhite:
		return "White"
	case ColorPrefBlack:
		return "Black"
	case ColorPrefStrongWhite:
		return "StrongWhite"
	case ColorPrefStrongBlack:
		return "StrongBlack"
	case ColorPrefMildWhite:
		return "MildWhite"
	case ColorPrefMildBlack:
		return "MildBlack"
	default:
		return "Unknown"
	}
}

// IsWhite returns true if the preference is for White (any strength).
func (p ColorPref) IsWhite() bool {
	return p == ColorPrefWhite || p == ColorPrefStrongWhite || p == ColorPrefMildWhite
}

// IsBlack returns true if the preference is for Black (any strength).
func (p ColorPref) IsBlack() bool {
	return p == ColorPrefBlack || p == ColorPrefStrongBlack || p == ColorPrefMildBlack
}

// IsStrong returns true if this is a strong preference (Type B only).
func (p ColorPref) IsStrong() bool {
	return p == ColorPrefStrongWhite || p == ColorPrefStrongBlack
}

// IsMild returns true if this is a mild preference (Type B only).
func (p ColorPref) IsMild() bool {
	return p == ColorPrefMildWhite || p == ColorPrefMildBlack
}

// Opposite returns the opposite colour preference (same strength).
func (p ColorPref) Opposite() ColorPref {
	switch p {
	case ColorPrefWhite:
		return ColorPrefBlack
	case ColorPrefBlack:
		return ColorPrefWhite
	case ColorPrefStrongWhite:
		return ColorPrefStrongBlack
	case ColorPrefStrongBlack:
		return ColorPrefStrongWhite
	case ColorPrefMildWhite:
		return ColorPrefMildBlack
	case ColorPrefMildBlack:
		return ColorPrefMildWhite
	default:
		return ColorPrefNone
	}
}

// ComputeColorPreference calculates a team's colour preference based on
// its match history, the preference type, and whether this is the last round.
//
// Implements Art. 1.7 of the Team Swiss system:
//   - Type A (1.7.1): Simple colour preferences (preference for White/Black, or none)
//   - Type B (1.7.2): Strong and mild colour preferences
//   - Type None: No colour preferences (always returns ColorPrefNone)
//
// Parameters:
//   - p: participant state with colour history
//   - prefType: which colour preference rules to use
//   - isLastRound: true if pairing the last round (affects Type B mild preferences)
func ComputeColorPreference(p *lexswiss.ParticipantState, prefType ColorPrefType, isLastRound bool) ColorPref {
	if prefType == ColorPrefTypeNone {
		return ColorPrefNone
	}

	played := filterPlayed(p.ColorHistory)
	if len(played) == 0 {
		return ColorPrefNone
	}

	cd := colorDifference(p.ColorHistory)
	last2 := lastNPlayed(p.ColorHistory, 2)

	// Check strong/simple White conditions (identical for A and B).
	// Preference for White if:
	//   CD < -1, OR
	//   CD in {0, -1} and last two played matches were Black.
	isWhitePref := cd < -1 || ((cd == 0 || cd == -1) && len(last2) >= 2 && last2[0] == lexswiss.ColorBlack && last2[1] == lexswiss.ColorBlack)

	// Check strong/simple Black conditions (identical for A and B).
	// Preference for Black if:
	//   CD > +1, OR
	//   CD in {0, +1} and last two played matches were White.
	isBlackPref := cd > 1 || ((cd == 0 || cd == 1) && len(last2) >= 2 && last2[0] == lexswiss.ColorWhite && last2[1] == lexswiss.ColorWhite)

	if prefType == ColorPrefTypeA {
		if isWhitePref {
			return ColorPrefWhite
		}
		if isBlackPref {
			return ColorPrefBlack
		}
		return ColorPrefNone
	}

	// Type B: strong preferences first (same conditions as Type A).
	if isWhitePref {
		return ColorPrefStrongWhite
	}
	if isBlackPref {
		return ColorPrefStrongBlack
	}

	// Type B: mild preferences.
	last1 := lastNPlayed(p.ColorHistory, 1)

	// Mild White if CD is -1, or CD is zero and not last round and last played was Black.
	if cd == -1 {
		return ColorPrefMildWhite
	}
	if cd == 0 && !isLastRound && len(last1) >= 1 && last1[0] == lexswiss.ColorBlack {
		return ColorPrefMildWhite
	}

	// Mild Black if CD is +1, or CD is zero and not last round and last played was White.
	if cd == 1 {
		return ColorPrefMildBlack
	}
	if cd == 0 && !isLastRound && len(last1) >= 1 && last1[0] == lexswiss.ColorWhite {
		return ColorPrefMildBlack
	}

	// No preference: CD is zero in last round.
	return ColorPrefNone
}

// colorDifference computes the colour difference (whites - blacks),
// ignoring ColorNone entries (byes, absences).
func colorDifference(history []lexswiss.Color) int {
	cd := 0
	for _, c := range history {
		switch c {
		case lexswiss.ColorWhite:
			cd++
		case lexswiss.ColorBlack:
			cd--
		}
	}
	return cd
}

// lastNPlayed returns the last N played colours (non-None) from history,
// in chronological order (oldest first).
func lastNPlayed(history []lexswiss.Color, n int) []lexswiss.Color {
	var played []lexswiss.Color
	for _, c := range history {
		if c != lexswiss.ColorNone {
			played = append(played, c)
		}
	}
	if len(played) <= n {
		return played
	}
	return played[len(played)-n:]
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
