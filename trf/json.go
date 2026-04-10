// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON encodes the Color as its TRF character string ("w", "b", or "-").
func (c Color) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c.Char()))
}

// UnmarshalJSON decodes a Color from its TRF character string.
func (c *Color) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("trf: unmarshal Color: %w", err)
	}
	if len(s) != 1 {
		return fmt.Errorf("trf: invalid color %q: must be a single character", s)
	}
	parsed, ok := parseColorChar(s[0])
	if !ok {
		return fmt.Errorf("trf: invalid color character %q", s)
	}
	*c = parsed
	return nil
}

// MarshalJSON encodes the ResultCode as its TRF character string ("1", "0", "=", etc.).
func (rc ResultCode) MarshalJSON() ([]byte, error) {
	ch := rc.Char()
	if ch == '?' {
		return nil, fmt.Errorf("trf: cannot marshal unknown ResultCode %d", rc)
	}
	return json.Marshal(string(ch))
}

// UnmarshalJSON decodes a ResultCode from its TRF character string.
func (rc *ResultCode) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("trf: unmarshal ResultCode: %w", err)
	}
	if len(s) != 1 {
		return fmt.Errorf("trf: invalid result code %q: must be a single character", s)
	}
	parsed, ok := parseResultChar(s[0])
	if !ok {
		return fmt.Errorf("trf: invalid result character %q", s)
	}
	*rc = parsed
	return nil
}
