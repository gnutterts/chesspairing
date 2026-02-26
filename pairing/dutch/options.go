// Package dutch implements the FIDE Dutch Swiss pairing system (C.04.3).
//
// The Dutch system is the most widely used Swiss pairing system in chess.
// Players are grouped by score, then paired within brackets using
// transposition and exchange algorithms subject to 21 quality criteria.
package dutch

import (
	"github.com/gnutterts/chesspairing"
)

// Pairer implements the chesspairing.Pairer interface for FIDE Dutch Swiss pairing.
type Pairer struct {
	opts Options
}

// Options holds Dutch-specific pairing configuration.
// All fields use pointer-nil pattern: nil = use default.
type Options struct {
	// Acceleration selects Baku acceleration mode.
	// Values: "none" (default), "baku".
	Acceleration *string `json:"acceleration,omitempty"`

	// TopSeedColor forces the top seed's color in round 1.
	// Values: "auto" (default), "white", "black".
	TopSeedColor *string `json:"topSeedColor,omitempty"`

	// ForbiddenPairs lists player ID pairs that must not be paired together.
	ForbiddenPairs [][]string `json:"forbiddenPairs,omitempty"`
}

// WithDefaults returns a copy of options with defaults applied for nil fields.
func (o Options) WithDefaults() Options {
	if o.Acceleration == nil {
		v := "none"
		o.Acceleration = &v
	}
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

	if v, ok := m["acceleration"].(string); ok {
		o.Acceleration = &v
	}
	if v, ok := m["topSeedColor"].(string); ok {
		o.TopSeedColor = &v
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

// New creates a new Dutch pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts.WithDefaults()}
}

// NewFromMap creates a new Dutch pairer from a generic options map.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
