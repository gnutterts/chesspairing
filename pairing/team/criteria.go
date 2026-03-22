package team

import "github.com/gnutterts/chesspairing/pairing/lexswiss"

// BuildCriteriaFunc creates a lexswiss.CriteriaFunc that checks Team Swiss
// quality criteria C8 and C9 for a proposed pair.
//
// C8: Minimise teams whose colour preference is not fulfilled.
//
//	Two teams with the same colour preference cannot both be satisfied,
//	so such pairings violate C8.
//
// C9 (Type B only): Minimise teams whose strong colour preference is not
//
//	fulfilled. Two teams with the same strong preference violate C9.
//
// C10 (upfloater opponents) is handled at the bracket/upfloater level,
// not per-pair, so it is not checked here.
//
// Parameters:
//   - prefType: colour preference type (A, B, or None)
//   - isLastTwoRounds: true if pairing one of the last two rounds (C7/C10 relaxation)
//   - isLastRound: true if pairing the last round (affects Type B mild preferences)
//
// Returns nil if colour preferences are disabled (ColorPrefTypeNone).
func BuildCriteriaFunc(prefType ColorPrefType, isLastTwoRounds bool, isLastRound bool) lexswiss.CriteriaFunc {
	if prefType == ColorPrefTypeNone {
		return nil
	}

	return func(a, b *lexswiss.ParticipantState) bool {
		prefA := ComputeColorPreference(a, prefType, isLastRound)
		prefB := ComputeColorPreference(b, prefType, isLastRound)

		// C8: both have a colour preference for the same colour → violation.
		// One of them won't get their preference.
		if prefA.IsWhite() && prefB.IsWhite() {
			return false
		}
		if prefA.IsBlack() && prefB.IsBlack() {
			return false
		}

		// C9 (Type B only): both have a strong preference for the same colour.
		// This is already covered by C8 above since strong prefs imply same-colour.
		// But C9 has lower priority — if C8 passes, C9 automatically passes for
		// opposite-colour cases. The same-colour case is caught by C8.
		// No additional check needed here.

		return true
	}
}
