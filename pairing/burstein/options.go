// Package burstein implements the Burstein Swiss pairing system (C.04.4.2).
//
// The Burstein system uses seeding rounds (delegating to Dutch matching)
// followed by opposition-index-based matching for post-seeding rounds.
// Seeding rounds = min(floor(totalRounds/2), 4).
//
// Key differences from Dutch:
//   - No float tracking (C9-C21 criteria not used)
//   - Bye selection: lowest score → most games played → lowest ranking
//   - Post-seeding rounds re-rank by opposition index before bracket building
//   - Criteria set is C1-C8 only
package burstein

import (
	chesspairing "github.com/gnutterts/chesspairing"
)

// Pairer implements the chesspairing.Pairer interface for the Burstein Swiss system.
type Pairer struct {
	opts Options
}

// Options holds Burstein-specific pairing configuration.
// All fields use pointer-nil pattern: nil = use default.
type Options struct {
	// TopSeedColor forces the top seed's color in round 1.
	// Values: "auto" (default), "white", "black".
	TopSeedColor *string `json:"topSeedColor,omitempty"`

	// ForbiddenPairs lists player ID pairs that must not be paired together.
	ForbiddenPairs [][]string `json:"forbiddenPairs,omitempty"`

	// TotalRounds is the planned number of rounds in the tournament.
	// Used to compute seeding round count. If nil, derived from state.
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

// New creates a new Burstein pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts.WithDefaults()}
}

// NewFromMap creates a new Burstein pairer from a generic options map.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// SeedingRounds returns the number of seeding rounds for the given total.
// Seeding rounds = min(floor(totalRounds/2), 4).
func SeedingRounds(totalRounds int) int {
	n := totalRounds / 2
	if n > 4 {
		return 4
	}
	return n
}

// IsSeedingRound returns true if the given round number (1-based) is a
// seeding round for a tournament with the given total rounds.
func IsSeedingRound(roundNumber, totalRounds int) bool {
	return roundNumber <= SeedingRounds(totalRounds)
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
