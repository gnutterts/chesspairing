package roundrobin

import (
	"github.com/gnutterts/chesspairing"
)

// Options holds configurable settings for round-robin pairing.
// All fields are pointers to distinguish "not set" (nil = use default)
// from "explicitly set."
type Options struct {
	// Cycles is the number of complete round-robins.
	// 1 = single round-robin (each pair plays once).
	// 2 = double round-robin (each pair plays twice, colors reversed).
	// Default: 1.
	Cycles *int `json:"cycles,omitempty"`

	// ColorBalance controls whether colors are swapped in even cycles
	// of a double (or multi-cycle) round-robin.
	// Default: true.
	ColorBalance *bool `json:"colorBalance,omitempty"`

	// SwapLastTwoRounds controls whether the last two rounds of cycle 1
	// are swapped in a double round-robin (Cycles=2), per the FIDE
	// recommendation (C.05 Annex 1) to avoid three consecutive games
	// with the same colour at the cycle boundary.
	// Only applies when Cycles == 2 and roundsPerCycle >= 2.
	// Default: true.
	SwapLastTwoRounds *bool `json:"swapLastTwoRounds,omitempty"`
}

// WithDefaults returns a copy of Options with all nil fields filled
// in with system defaults.
func (o Options) WithDefaults() Options {
	if o.Cycles == nil {
		o.Cycles = chesspairing.IntPtr(1)
	}
	if o.ColorBalance == nil {
		o.ColorBalance = chesspairing.BoolPtr(true)
	}
	if o.SwapLastTwoRounds == nil {
		o.SwapLastTwoRounds = chesspairing.BoolPtr(true)
	}
	return o
}

// ParseOptions converts a map[string]any (from Firestore/JSON) into
// typed Options. Unrecognized keys are ignored.
func ParseOptions(m map[string]any) Options {
	var o Options
	if v, ok := chesspairing.GetInt(m, "cycles"); ok {
		o.Cycles = &v
	}
	if v, ok := chesspairing.GetBool(m, "colorBalance"); ok {
		o.ColorBalance = &v
	}
	if v, ok := chesspairing.GetBool(m, "swapLastTwoRounds"); ok {
		o.SwapLastTwoRounds = &v
	}
	return o
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
