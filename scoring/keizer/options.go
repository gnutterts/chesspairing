package keizer

// Options holds configurable settings for Keizer point scoring.
// All fields are pointers to distinguish "not set" (nil = use default)
// from "explicitly set to zero."
type Options struct {
	// ValueNumberBase is the top-ranked player's value number.
	// Default: player count (N).
	ValueNumberBase *int `json:"valueNumberBase,omitempty"`

	// ValueNumberStep is the decrement per rank position.
	// Player at rank r gets: ValueNumberBase - (r-1) * ValueNumberStep.
	// Default: 1.
	ValueNumberStep *int `json:"valueNumberStep,omitempty"`

	// AbsentPenaltyFraction is the fraction of own value number awarded
	// when a player is absent (neither plays nor receives a bye).
	// Default: 0.5 (50% of own value).
	AbsentPenaltyFraction *float64 `json:"absentPenaltyFraction,omitempty"`

	// ByeValueFraction is the fraction of own value number awarded
	// when a player receives a bye.
	// Default: 0.5 (50% of own value).
	ByeValueFraction *float64 `json:"byeValueFraction,omitempty"`

	// LateJoinHandicap is points deducted from a player's total
	// for each round missed before they joined.
	// Default: 0.
	LateJoinHandicap *float64 `json:"lateJoinHandicap,omitempty"`

	// WinFraction is the multiplier applied to the opponent's value number
	// for a win. Points for win = opponent_value × WinFraction.
	// Default: 1.0.
	WinFraction *float64 `json:"winFraction,omitempty"`

	// DrawFraction is the multiplier applied to the opponent's value number
	// for a draw. Points for draw = opponent_value × DrawFraction.
	// Default: 0.5.
	DrawFraction *float64 `json:"drawFraction,omitempty"`

	// LossFraction is the multiplier applied to the opponent's value number
	// for a loss. Points for loss = opponent_value × LossFraction.
	// Default: 0.0.
	LossFraction *float64 `json:"lossFraction,omitempty"`
}

// WithDefaults returns a copy of Options with all nil fields filled
// in with system defaults. playerCount is the number of active players
// in the tournament.
func (o Options) WithDefaults(playerCount int) Options {
	if o.ValueNumberBase == nil {
		o.ValueNumberBase = intPtr(playerCount)
	}
	if o.ValueNumberStep == nil {
		o.ValueNumberStep = intPtr(1)
	}
	if o.AbsentPenaltyFraction == nil {
		o.AbsentPenaltyFraction = float64Ptr(0.5)
	}
	if o.ByeValueFraction == nil {
		o.ByeValueFraction = float64Ptr(0.5)
	}
	if o.LateJoinHandicap == nil {
		o.LateJoinHandicap = float64Ptr(0)
	}
	if o.WinFraction == nil {
		o.WinFraction = float64Ptr(1.0)
	}
	if o.DrawFraction == nil {
		o.DrawFraction = float64Ptr(0.5)
	}
	if o.LossFraction == nil {
		o.LossFraction = float64Ptr(0.0)
	}
	return o
}

// ValueNumber calculates the value number for a player at the given rank.
// Rank is 1-based (rank 1 = strongest player).
func (o Options) ValueNumber(rank int) int {
	return *o.ValueNumberBase - (rank-1)**o.ValueNumberStep
}

func intPtr(v int) *int             { return &v }
func float64Ptr(v float64) *float64 { return &v }

// ParseOptions converts a map[string]any (from Firestore/JSON) into
// typed Options. Unrecognized keys are ignored. Type mismatches use defaults.
func ParseOptions(m map[string]any) Options {
	var o Options
	if v, ok := getInt(m, "valueNumberBase"); ok {
		o.ValueNumberBase = &v
	}
	if v, ok := getInt(m, "valueNumberStep"); ok {
		o.ValueNumberStep = &v
	}
	if v, ok := getFloat64(m, "absentPenaltyFraction"); ok {
		o.AbsentPenaltyFraction = &v
	}
	if v, ok := getFloat64(m, "byeValueFraction"); ok {
		o.ByeValueFraction = &v
	}
	if v, ok := getFloat64(m, "lateJoinHandicap"); ok {
		o.LateJoinHandicap = &v
	}
	if v, ok := getFloat64(m, "winFraction"); ok {
		o.WinFraction = &v
	}
	if v, ok := getFloat64(m, "drawFraction"); ok {
		o.DrawFraction = &v
	}
	if v, ok := getFloat64(m, "lossFraction"); ok {
		o.LossFraction = &v
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

// getInt extracts an int from a map, handling both int and float64 values.
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
