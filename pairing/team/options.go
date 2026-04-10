// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package team implements the FIDE Swiss Team Pairing System (C.04.6).
//
// The Team Swiss system (approved Oct 2025, effective Feb 2026) pairs teams
// using lexicographic bracket pairing (Art. 3.6) from the shared
// pairing/lexswiss package. Each PlayerEntry in TournamentState represents
// a team.
//
// Key characteristics:
//   - Configurable primary score: "match" points (default) or "game" points
//   - Two colour preference types: Type A (simple) and Type B (strong + mild)
//   - No absolute colour criteria (C8-C10 are quality criteria only)
//   - C7 and C10 relaxed in last TWO rounds
//   - 9-step colour allocation (Art. 4)
//   - Colour determined by first board assignment (Art. 1.6.1)
package team

import (
	"strings"

	"github.com/gnutterts/chesspairing"
)

// Pairer implements the chesspairing.Pairer interface for the Team Swiss system.
type Pairer struct {
	opts Options
}

// Options holds Team Swiss-specific pairing configuration.
// All pointer fields use nil = use default.
type Options struct {
	// TopSeedColor is the colour determined by drawing of lots before round 1.
	// Values: "white" (default initial-colour), "black".
	// This is called "initial-colour" in Art. 4.1.
	TopSeedColor *string `json:"topSeedColor,omitempty"`

	// ForbiddenPairs lists team ID pairs that must not be paired together.
	ForbiddenPairs [][]string `json:"forbiddenPairs,omitempty"`

	// TotalRounds is the total number of rounds in the tournament.
	// Used to determine "last two rounds" for C7/C10 relaxation,
	// and "last round" for Type B mild preference calculation.
	TotalRounds *int `json:"totalRounds,omitempty"`

	// ColorPreferenceType selects which colour preference rules to use.
	// Values: "A" (default, Type A simple), "B" (Type B strong+mild),
	// "none" (no colour preferences).
	// Corresponds to Art. 1.7.
	ColorPreferenceType *string `json:"colorPreferenceType,omitempty"`

	// PrimaryScore selects which score is used for pairing.
	// Values: "match" (default, match points), "game" (game points).
	// The other score becomes the "secondary score" used for colour allocation
	// (Art. 4.2.2).
	// Corresponds to Art. 1.2.
	PrimaryScore *string `json:"primaryScore,omitempty"`
}

// WithDefaults returns a copy of options with defaults applied for nil fields.
func (o Options) WithDefaults() Options {
	if o.ColorPreferenceType == nil {
		v := "A"
		o.ColorPreferenceType = &v
	}
	if o.PrimaryScore == nil {
		v := "match"
		o.PrimaryScore = &v
	}
	if o.TopSeedColor == nil {
		v := "white"
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
	if v, ok := chesspairing.GetInt(m, "totalRounds"); ok {
		o.TotalRounds = &v
	}
	if v, ok := m["colorPreferenceType"].(string); ok {
		o.ColorPreferenceType = &v
	}
	if v, ok := m["primaryScore"].(string); ok {
		o.PrimaryScore = &v
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

// New creates a new Team Swiss pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts.WithDefaults()}
}

// NewFromMap creates a new Team Swiss pairer from a generic options map.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// resolveColorPrefType converts a string option to ColorPrefType.
func resolveColorPrefType(s string) ColorPrefType {
	switch strings.ToLower(s) {
	case "b":
		return ColorPrefTypeB
	case "none":
		return ColorPrefTypeNone
	default:
		return ColorPrefTypeA
	}
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
