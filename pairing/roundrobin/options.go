package roundrobin

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
}

// WithDefaults returns a copy of Options with all nil fields filled
// in with system defaults.
func (o Options) WithDefaults() Options {
	if o.Cycles == nil {
		o.Cycles = intPtr(1)
	}
	if o.ColorBalance == nil {
		o.ColorBalance = boolPtr(true)
	}
	return o
}

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }

// ParseOptions converts a map[string]any (from Firestore/JSON) into
// typed Options. Unrecognized keys are ignored.
func ParseOptions(m map[string]any) Options {
	var o Options
	if v, ok := getInt(m, "cycles"); ok {
		o.Cycles = &v
	}
	if v, ok := getBool(m, "colorBalance"); ok {
		o.ColorBalance = &v
	}
	return o
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

func getBool(m map[string]any, key string) (bool, bool) {
	v, ok := m[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}
