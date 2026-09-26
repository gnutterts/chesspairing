// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import "testing"

func TestParseOptionsStrictRejectsUnknownOption(t *testing.T) {
	_, err := ParseOptionsStrict(map[string]any{"unknownOption": true})
	if err == nil {
		t.Fatal("ParseOptionsStrict accepted an unknown option")
	}
}

func TestParseOptionsIgnoresUnknownOption(t *testing.T) {
	opts := ParseOptions(map[string]any{"winFraction": 0.5, "unknownOption": true})
	if opts.WinFraction == nil || *opts.WinFraction != 0.5 {
		t.Fatalf("WinFraction = %v, want 0.5", opts.WinFraction)
	}
}
