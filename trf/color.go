// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"fmt"
	"strings"
)

// normalizeInitialColor converts TRF initial-colour values to pairing option values.
func normalizeInitialColor(raw string) (string, error) {
	switch strings.ToLower(raw) {
	case "w", "white", "white1":
		return "white", nil
	case "b", "black", "black1":
		return "black", nil
	default:
		return "", fmt.Errorf("invalid initial color %q", raw)
	}
}
