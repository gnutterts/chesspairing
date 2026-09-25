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

func TestParseOptions_TotalRounds(t *testing.T) {
	for _, totalRounds := range []any{9, int64(9), 9.0} {
		opts := ParseOptions(map[string]any{"totalRounds": totalRounds})
		if opts.TotalRounds == nil || *opts.TotalRounds != 9 {
			t.Errorf("totalRounds %T parsed as %v, want 9", totalRounds, opts.TotalRounds)
		}
	}
}

func TestParseOptions_NonIntegralTotalRounds(t *testing.T) {
	opts := ParseOptions(map[string]any{"totalRounds": 9.9})
	if opts.TotalRounds != nil || !opts.totalRoundsInvalid {
		t.Errorf("non-integral totalRounds parsed as %+v, want invalid option", opts)
	}
}

func TestNewFromMap(t *testing.T) {
	p := NewFromMap(map[string]any{"topSeedColor": "white"})
	if p == nil {
		t.Fatal("NewFromMap returned nil")
	}
}
