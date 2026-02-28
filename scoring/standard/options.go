package standard

// Options holds configurable settings for standard (1-½-0) scoring.
// All fields are pointers to distinguish "not set" (nil = use default)
// from "explicitly set to zero."
type Options struct {
	// PointWin is the points awarded for a win.
	// Default: 1.0.
	PointWin *float64 `json:"pointWin,omitempty"`

	// PointDraw is the points awarded for a draw.
	// Default: 0.5.
	PointDraw *float64 `json:"pointDraw,omitempty"`

	// PointLoss is the points awarded for a loss.
	// Default: 0.0.
	PointLoss *float64 `json:"pointLoss,omitempty"`

	// PointBye is the points awarded for a bye.
	// Default: 1.0 (full point bye, FIDE default).
	PointBye *float64 `json:"pointBye,omitempty"`

	// PointForfeitWin is the points awarded for a forfeit win.
	// Default: 1.0.
	PointForfeitWin *float64 `json:"pointForfeitWin,omitempty"`

	// PointForfeitLoss is the points awarded for a forfeit loss.
	// Default: 0.0.
	PointForfeitLoss *float64 `json:"pointForfeitLoss,omitempty"`

	// PointAbsent is the points awarded when a player is absent
	// (neither plays nor receives a bye).
	// Default: 0.0.
	PointAbsent *float64 `json:"pointAbsent,omitempty"`
}

// WithDefaults returns a copy of Options with all nil fields filled
// in with standard FIDE defaults (1-½-0).
func (o Options) WithDefaults() Options {
	if o.PointWin == nil {
		o.PointWin = float64Ptr(1.0)
	}
	if o.PointDraw == nil {
		o.PointDraw = float64Ptr(0.5)
	}
	if o.PointLoss == nil {
		o.PointLoss = float64Ptr(0.0)
	}
	if o.PointBye == nil {
		o.PointBye = float64Ptr(1.0)
	}
	if o.PointForfeitWin == nil {
		o.PointForfeitWin = float64Ptr(1.0)
	}
	if o.PointForfeitLoss == nil {
		o.PointForfeitLoss = float64Ptr(0.0)
	}
	if o.PointAbsent == nil {
		o.PointAbsent = float64Ptr(0.0)
	}
	return o
}

func float64Ptr(v float64) *float64 { return &v }

// ParseOptions converts a map[string]any (from Firestore/JSON) into
// typed Options. Unrecognized keys are ignored. Type mismatches use defaults.
func ParseOptions(m map[string]any) Options {
	var o Options
	if v, ok := getFloat64(m, "pointWin"); ok {
		o.PointWin = &v
	}
	if v, ok := getFloat64(m, "pointDraw"); ok {
		o.PointDraw = &v
	}
	if v, ok := getFloat64(m, "pointLoss"); ok {
		o.PointLoss = &v
	}
	if v, ok := getFloat64(m, "pointBye"); ok {
		o.PointBye = &v
	}
	if v, ok := getFloat64(m, "pointForfeitWin"); ok {
		o.PointForfeitWin = &v
	}
	if v, ok := getFloat64(m, "pointForfeitLoss"); ok {
		o.PointForfeitLoss = &v
	}
	if v, ok := getFloat64(m, "pointAbsent"); ok {
		o.PointAbsent = &v
	}
	return o
}

// getFloat64 extracts a float64 from a map, handling both float64 and int values.
func getFloat64(m map[string]any, key string) (float64, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
