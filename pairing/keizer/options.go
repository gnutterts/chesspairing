// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"fmt"
	"sort"

	"github.com/gnutterts/chesspairing"
	keizerscoring "github.com/gnutterts/chesspairing/scoring/keizer"
)

// Options holds configurable settings for Keizer pairing.
// All fields are pointers to distinguish "not set" (nil = use default)
// from "explicitly set."
type Options struct {
	// AllowRepeatPairings controls whether players can be paired against
	// the same opponent again within the tournament.
	// Default: true (Keizer tournaments often span many rounds).
	AllowRepeatPairings *bool `json:"allowRepeatPairings,omitempty"`

	// MinRoundsBetweenRepeats is the minimum number of rounds that must
	// pass before two players can be paired again.
	// Only applies when AllowRepeatPairings is true.
	// Default: 3.
	MinRoundsBetweenRepeats *int `json:"minRoundsBetweenRepeats,omitempty"`

	// InitialOrder controls tiebreaking when players have the same score/rating.
	// "rating-name" (default) sorts alphabetically.
	// "rating-entry" sorts by entry sequence (slice index).
	InitialOrder *string `json:"initialOrder,omitempty"`

	// ScoringOptions configures the internal Keizer scorer used for ranking.
	// When nil, the scorer uses its own defaults.
	ScoringOptions *keizerscoring.Options `json:"scoringOptions,omitempty"`
}

// WithDefaults returns a copy of Options with all nil fields filled
// in with system defaults.
func (o Options) WithDefaults() Options {
	if o.AllowRepeatPairings == nil {
		o.AllowRepeatPairings = chesspairing.BoolPtr(true)
	}
	if o.MinRoundsBetweenRepeats == nil {
		o.MinRoundsBetweenRepeats = chesspairing.IntPtr(3)
	}
	if o.InitialOrder == nil {
		o.InitialOrder = chesspairing.StringPtr("rating-name")
	}
	return o
}

// ParseOptions converts a map[string]any (from Firestore/JSON) into
// typed Options. Unrecognized keys are ignored.
func ParseOptions(m map[string]any) Options {
	o, err := ParseOptionsStrict(m)
	if err == nil {
		return o
	}
	return parseOptions(m)
}

// ParseOptionsStrict converts a map[string]any into typed Options and
// reports unrecognized keys, including keys in scoringOptions.
func ParseOptionsStrict(m map[string]any) (Options, error) {
	if err := validateOptionKeys(m); err != nil {
		return Options{}, err
	}
	if _, err := keizerscoring.ParseOptionsStrict(scoringOptionsMap(m)); err != nil {
		return Options{}, err
	}
	o := parseOptions(m)
	if o.InitialOrder != nil && *o.InitialOrder != "rating-name" && *o.InitialOrder != "rating-entry" {
		return Options{}, fmt.Errorf("invalid Keizer initialOrder %q", *o.InitialOrder)
	}
	return o, nil
}

func parseOptions(m map[string]any) Options {
	var o Options
	if v, ok := chesspairing.GetBool(m, "allowRepeatPairings"); ok {
		o.AllowRepeatPairings = &v
	}
	if v, ok := chesspairing.GetInt(m, "minRoundsBetweenRepeats"); ok {
		o.MinRoundsBetweenRepeats = &v
	}
	if v, ok := chesspairing.GetString(m, "initialOrder"); ok {
		o.InitialOrder = &v
	}
	// Prefer nested scoringOptions. Top-level scoring options remain supported
	// for compatibility. Only set ScoringOptions if at least one scoring field
	// was present, to preserve the nil-means-default convention.
	scoringOpts := keizerscoring.ParseOptions(scoringOptionsMap(m))
	if scoringOpts != (keizerscoring.Options{}) {
		o.ScoringOptions = &scoringOpts
	}
	return o
}

func scoringOptionsMap(m map[string]any) map[string]any {
	result := make(map[string]any)
	for key := range scoringOptionKeys {
		if value, ok := m[key]; ok {
			result[key] = value
		}
	}
	if nested, ok := m["scoringOptions"].(map[string]any); ok {
		for key, value := range nested {
			result[key] = value
		}
	}
	return result
}

// optionKeys lists every accepted pairing option. topSeedColor and
// totalRounds are shared Swiss/tournament config keys; Keizer ignores them,
// but the CLI/TRF path always populates them for every pairing system, so
// they must remain accepted.
var optionKeys = map[string]struct{}{
	"allowRepeatPairings":     {},
	"minRoundsBetweenRepeats": {},
	"initialOrder":            {},
	"scoringOptions":          {},
	"topSeedColor":            {},
	"totalRounds":             {},
}

var scoringOptionKeys = map[string]struct{}{
	"valueNumberBase":          {},
	"valueNumberStep":          {},
	"winFraction":              {},
	"drawFraction":             {},
	"lossFraction":             {},
	"forfeitWinFraction":       {},
	"forfeitLossFraction":      {},
	"doubleForfeitFraction":    {},
	"byeValueFraction":         {},
	"halfByeFraction":          {},
	"zeroByeFraction":          {},
	"absentPenaltyFraction":    {},
	"excusedAbsentFraction":    {},
	"clubCommitmentFraction":   {},
	"byeFixedValue":            {},
	"halfByeFixedValue":        {},
	"zeroByeFixedValue":        {},
	"absentFixedValue":         {},
	"excusedAbsentFixedValue":  {},
	"clubCommitmentFixedValue": {},
	"selfVictory":              {},
	"absenceLimit":             {},
	"absenceDecay":             {},
	"frozen":                   {},
	"lateJoinHandicap":         {},
}

func validateOptionKeys(m map[string]any) error {
	unknown := make([]string, 0)
	for key := range m {
		if _, ok := optionKeys[key]; ok {
			continue
		}
		if _, ok := scoringOptionKeys[key]; !ok {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return fmt.Errorf("unknown Keizer pairing option %q", unknown[0])
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
