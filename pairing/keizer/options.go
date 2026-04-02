package keizer

import (
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
	return o
}

// ParseOptions converts a map[string]any (from Firestore/JSON) into
// typed Options. Unrecognized keys are ignored.
func ParseOptions(m map[string]any) Options {
	var o Options
	if v, ok := chesspairing.GetBool(m, "allowRepeatPairings"); ok {
		o.AllowRepeatPairings = &v
	}
	if v, ok := chesspairing.GetInt(m, "minRoundsBetweenRepeats"); ok {
		o.MinRoundsBetweenRepeats = &v
	}
	// Pass all keys to the scoring parser for scoring-related options.
	// Only set ScoringOptions if at least one scoring field was present,
	// to preserve the nil-means-default convention.
	scoringOpts := keizerscoring.ParseOptions(m)
	if scoringOpts != (keizerscoring.Options{}) {
		o.ScoringOptions = &scoringOpts
	}
	return o
}
