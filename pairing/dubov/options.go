// Package dubov implements the Dubov Swiss pairing system (C.04.4.1).
//
// The Dubov system is an ARO-equalization Swiss variant that splits score
// groups by colour preference (G1=White-seekers, G2=Black-seekers), sorts
// G1 by ascending ARO, and uses transposition-based matching. It has 10
// criteria (C1-C10) and tracks MaxT upfloater limits.
//
// Key differences from Dutch:
//   - No S1/S2 half-split; instead G1/G2 by colour preference
//   - S1 (=G1) sorted by ascending ARO, not descending score then TPN
//   - Transposition-only matching (no exchanges)
//   - MaxT upfloater limit = 2 + floor(Rnds/5)
//   - 10 criteria (C1-C10), not 13 (C8-C21)
//   - Distinct 5-rule colour allocation
//   - Distinct bye selection (Art. 2.3)
package dubov

import (
	chesspairing "github.com/gnutterts/chesspairing"
)

// Pairer implements the chesspairing.Pairer interface for the Dubov Swiss system.
type Pairer struct {
	opts Options
}

// Options holds Dubov-specific pairing configuration.
// All pointer fields use nil = use default.
type Options struct {
	// TopSeedColor forces the top seed's color in round 1.
	// Values: "auto" (default), "white", "black".
	TopSeedColor *string `json:"topSeedColor,omitempty"`

	// ForbiddenPairs lists player ID pairs that must not be paired together.
	ForbiddenPairs [][]string `json:"forbiddenPairs,omitempty"`

	// TotalRounds is the planned number of rounds in the tournament.
	// If nil, derived from state.
	TotalRounds *int `json:"totalRounds,omitempty"`
}

// WithDefaults returns a copy of options with defaults applied for nil fields.
func (o Options) WithDefaults() Options {
	if o.TopSeedColor == nil {
		v := "auto"
		o.TopSeedColor = &v
	}
	return o
}

// ParseOptions converts a generic map[string]any into typed Options.
func ParseOptions(m map[string]any) Options {
	var o Options
	if m == nil {
		return o
	}

	if v, ok := m["topSeedColor"].(string); ok {
		o.TopSeedColor = &v
	}
	if v, ok := m["totalRounds"].(float64); ok {
		n := int(v)
		o.TotalRounds = &n
	}
	if v, ok := m["forbiddenPairs"].([]any); ok {
		for _, pair := range v {
			if arr, ok := pair.([]any); ok && len(arr) == 2 {
				s1, ok1 := arr[0].(string)
				s2, ok2 := arr[1].(string)
				if ok1 && ok2 {
					o.ForbiddenPairs = append(o.ForbiddenPairs, []string{s1, s2})
				}
			}
		}
	}

	return o
}

// New creates a new Dubov pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts.WithDefaults()}
}

// NewFromMap creates a new Dubov pairer from a generic options map.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
