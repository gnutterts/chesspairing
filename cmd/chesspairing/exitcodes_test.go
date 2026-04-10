// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/exitcodes_test.go
package main

import "testing"

func TestExitCodes(t *testing.T) {
	// Verify constants have expected values (bbpPairings compatible)
	tests := []struct {
		name string
		code int
		want int
	}{
		{"success", ExitSuccess, 0},
		{"no valid pairing", ExitNoPairing, 1},
		{"unexpected error", ExitUnexpected, 2},
		{"invalid input", ExitInvalidInput, 3},
		{"size overflow", ExitSizeOverflow, 4},
		{"file access error", ExitFileAccess, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Errorf("got %d, want %d", tt.code, tt.want)
			}
		})
	}
}
