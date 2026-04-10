// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import "testing"

func TestParseOptions_Empty(t *testing.T) {
	opts := ParseOptions(nil)
	defaults := opts.WithDefaults()
	if defaults.Acceleration == nil || *defaults.Acceleration != "none" {
		t.Error("default acceleration should be 'none'")
	}
	if defaults.TopSeedColor == nil || *defaults.TopSeedColor != "auto" {
		t.Error("default top seed color should be 'auto'")
	}
}

func TestParseOptions_WithValues(t *testing.T) {
	m := map[string]any{
		"acceleration":   "baku",
		"topSeedColor":   "black",
		"forbiddenPairs": []any{[]any{"p1", "p2"}, []any{"p3", "p4"}},
	}
	opts := ParseOptions(m)
	if opts.Acceleration == nil || *opts.Acceleration != "baku" {
		t.Error("acceleration should be 'baku'")
	}
	if opts.TopSeedColor == nil || *opts.TopSeedColor != "black" {
		t.Error("top seed color should be 'black'")
	}
	if len(opts.ForbiddenPairs) != 2 {
		t.Errorf("expected 2 forbidden pairs, got %d", len(opts.ForbiddenPairs))
	}
}

func TestNewFromMap(t *testing.T) {
	p := NewFromMap(map[string]any{"topSeedColor": "white"})
	if p == nil {
		t.Fatal("NewFromMap returned nil")
	}
}
