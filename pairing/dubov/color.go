package dubov

import (
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// AllocateColor decides which player gets White and which gets Black
// for a Dubov pairing. Returns (whiteID, blackID).
//
// Per Art. 5, the 5 rules are equivalent to swisslib.AllocateColor
// with topScorerRules=false. Dubov does not have topscorer-specific
// colour rules.
func AllocateColor(a, b *swisslib.PlayerState, boardNumber int, topSeedColor *swisslib.Color) (string, string) {
	return swisslib.AllocateColor(a, b, false, boardNumber, topSeedColor)
}
