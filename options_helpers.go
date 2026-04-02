package chesspairing

// Float64Ptr returns a pointer to v. Used by options packages to set
// pointer-nil-pattern fields.
func Float64Ptr(v float64) *float64 { return &v }

// IntPtr returns a pointer to v.
func IntPtr(v int) *int { return &v }

// BoolPtr returns a pointer to v.
func BoolPtr(v bool) *bool { return &v }

// GetFloat64 extracts a float64 from a map, handling float64, int, and
// int64 value types. Returns (0, false) if the key is missing or has an
// incompatible type.
func GetFloat64(m map[string]any, key string) (float64, bool) {
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

// GetInt extracts an int from a map, handling int, int64, and float64
// value types. Returns (0, false) if the key is missing or has an
// incompatible type.
func GetInt(m map[string]any, key string) (int, bool) {
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

// GetBool extracts a bool from a map. Returns (false, false) if the key
// is missing or has an incompatible type.
func GetBool(m map[string]any, key string) (bool, bool) {
	v, ok := m[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}
