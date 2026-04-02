// Package doubleswiss implements the FIDE Double-Swiss pairing system (C.04.5).
//
// The Double-Swiss system (approved Oct 2025, effective Feb 2026) treats each
// round as a 2-game match. Scores are cumulative game points (win=2, draw=1,
// loss=0 per match). The system uses lexicographic bracket pairing (Art. 3.6)
// from the shared pairing/lexswiss package.
//
// Key characteristics:
//   - Each round is a 2-game match (colours alternate within the match)
//   - PAB awards 1.5 points (Art. 3.4)
//   - Lexicographic enumeration of pairings (no Blossom matching)
//   - Simplified criteria: C1 (absolute) + C8 (colour, relaxable in last round)
//   - Colour allocation with 5-step priority (Art. 4)
package doubleswiss

import "github.com/gnutterts/chesspairing"

// Pairer implements the chesspairing.Pairer interface for the Double-Swiss system.
type Pairer struct {
	opts Options
}

// Options holds Double-Swiss-specific pairing configuration.
// All pointer fields use nil = use default.
type Options struct {
	// TopSeedColor forces the top seed's colour in round 1.
	// Values: "auto" (default), "white", "black".
	TopSeedColor *string `json:"topSeedColor,omitempty"`

	// ForbiddenPairs lists participant ID pairs that must not be paired together.
	ForbiddenPairs [][]string `json:"forbiddenPairs,omitempty"`

	// TotalRounds is the total number of rounds in the tournament.
	// Used to determine "last round" for criteria relaxation (C8).
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
		iv := int(v)
		o.TotalRounds = &iv
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

// New creates a new Double-Swiss pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts.WithDefaults()}
}

// NewFromMap creates a new Double-Swiss pairer from a generic options map.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
