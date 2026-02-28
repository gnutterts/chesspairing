package keizer

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
}

// WithDefaults returns a copy of Options with all nil fields filled
// in with system defaults.
func (o Options) WithDefaults() Options {
	if o.AllowRepeatPairings == nil {
		o.AllowRepeatPairings = boolPtr(true)
	}
	if o.MinRoundsBetweenRepeats == nil {
		o.MinRoundsBetweenRepeats = intPtr(3)
	}
	return o
}

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }

// ParseOptions converts a map[string]any (from Firestore/JSON) into
// typed Options. Unrecognized keys are ignored.
func ParseOptions(m map[string]any) Options {
	var o Options
	if v, ok := getBool(m, "allowRepeatPairings"); ok {
		o.AllowRepeatPairings = &v
	}
	if v, ok := getInt(m, "minRoundsBetweenRepeats"); ok {
		o.MinRoundsBetweenRepeats = &v
	}
	return o
}

func getBool(m map[string]any, key string) (bool, bool) {
	v, ok := m[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func getInt(m map[string]any, key string) (int, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		return 0, false
	}
}
