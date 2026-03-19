// Package lim implements the Lim Swiss pairing system (C.04.4.3).
//
// The Lim system (approved 1987, amended through 1999) is a Swiss variant
// that processes scoregroups in median-first order and uses exchange-based
// matching within scoregroups. It has four floater types (A-D) with priority
// ordering and median-aware colour allocation.
//
// Key differences from Dutch/Burstein/Dubov:
//   - Scoregroups processed: highest -> above median, then lowest -> up to median, median last
//   - Exchange-based matching (Art. 4), not Blossom or transposition matching
//   - Four floater types (A-D) with priority ordering (Art. 3.9)
//   - Compatibility: no 3 consecutive same colour, no 3+ colour imbalance (Art. 2.1)
//   - Colour allocation with median-aware tiebreaking (Art. 5.4)
//   - Optional "Maxi-tournament" 100-point rating constraint (Art. 3.2.3)
package lim

import (
	chesspairing "github.com/gnutterts/chesspairing"
)

// Pairer implements the chesspairing.Pairer interface for the Lim Swiss system.
type Pairer struct {
	opts Options
}

// Options holds Lim-specific pairing configuration.
// All pointer fields use nil = use default.
type Options struct {
	// TopSeedColor forces the top seed's color in round 1.
	// Values: "auto" (default), "white", "black".
	TopSeedColor *string `json:"topSeedColor,omitempty"`

	// ForbiddenPairs lists player ID pairs that must not be paired together.
	ForbiddenPairs [][]string `json:"forbiddenPairs,omitempty"`

	// MaxiTournament enables the 100-point rating constraint for exchanges
	// and floater selection (Art. 3.2.3, 3.8, 5.7).
	// Default: false.
	MaxiTournament *bool `json:"maxiTournament,omitempty"`
}

// WithDefaults returns a copy of options with defaults applied for nil fields.
func (o Options) WithDefaults() Options {
	if o.TopSeedColor == nil {
		v := "auto"
		o.TopSeedColor = &v
	}
	if o.MaxiTournament == nil {
		v := false
		o.MaxiTournament = &v
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
	if v, ok := m["maxiTournament"].(bool); ok {
		o.MaxiTournament = &v
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

// New creates a new Lim pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts.WithDefaults()}
}

// NewFromMap creates a new Lim pairer from a generic options map.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
