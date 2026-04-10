// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"math"
	"testing"
)

func TestDpFromP(t *testing.T) {
	tests := []struct {
		p    float64 // fractional score (0.0 to 1.0)
		want float64 // expected dp
	}{
		{1.0, 800},   // 100% → dp = 800
		{0.99, 677},  // 99% → dp = 677
		{0.92, 401},  // 92% → dp = 401
		{0.83, 273},  // 83% → dp = 273
		{0.78, 220},  // 78% → dp = 220
		{0.73, 175},  // 73% → dp = 175
		{0.68, 133},  // 68% → dp = 133
		{0.63, 95},   // 63% → dp = 95
		{0.58, 57},   // 58% → dp = 57
		{0.53, 21},   // 53% → dp = 21
		{0.50, 0},    // 50% → dp = 0
		{0.47, -21},  // 47% → dp = -21
		{0.42, -57},  // 42% → dp = -57
		{0.32, -133}, // 32% → dp = -133
		{0.0, -800},  // 0% → dp = -800
	}

	for _, tt := range tests {
		got := dpFromP(tt.p)
		if got != tt.want {
			t.Errorf("dpFromP(%.2f) = %v, want %v", tt.p, got, tt.want)
		}
	}
}

func TestDpFromPInterpolation(t *testing.T) {
	// 75% → dp = 193, 76% → dp = 202. 75.5% should interpolate to 197.5.
	got := dpFromP(0.755)
	if math.Abs(got-197.5) > 0.01 {
		t.Errorf("dpFromP(0.755) = %v, want ~197.5", got)
	}
}

func TestExpectedScore(t *testing.T) {
	// dp = 0 → expected = 0.50
	got := expectedScore(0)
	if got != 0.50 {
		t.Errorf("expectedScore(0) = %v, want 0.50", got)
	}

	// dp = 100 → expected: look up in table. dp=95 → p=0.63, dp=102 → p=0.64
	// So dp=100 is between: 0.63 + (100-95)/(102-95) * (0.64-0.63) ≈ 0.6371
	got = expectedScore(100)
	if got < 0.63 || got > 0.65 {
		t.Errorf("expectedScore(100) = %v, want ~0.64", got)
	}

	// dp = -200 → look up: dp=-202 → p=0.24, dp=-193 → p=0.25
	// So dp=-200 is between: 0.24 + (-200-(-202))/(-193-(-202)) * (0.25-0.24) ≈ 0.2422
	got = expectedScore(-200)
	if got < 0.24 || got > 0.25 {
		t.Errorf("expectedScore(-200) = %v, want ~0.24", got)
	}
}

func TestDpFromPBoundaries(t *testing.T) {
	// Values outside [0,1] should be clamped.
	if got := dpFromP(-0.5); got != -800 {
		t.Errorf("dpFromP(-0.5) = %v, want -800 (clamped)", got)
	}
	if got := dpFromP(1.5); got != 800 {
		t.Errorf("dpFromP(1.5) = %v, want 800 (clamped)", got)
	}
	// Exact boundaries.
	if got := dpFromP(0.0); got != -800 {
		t.Errorf("dpFromP(0.0) = %v, want -800", got)
	}
	if got := dpFromP(1.0); got != 800 {
		t.Errorf("dpFromP(1.0) = %v, want 800", got)
	}
	if got := dpFromP(0.50); got != 0 {
		t.Errorf("dpFromP(0.50) = %v, want 0", got)
	}
}

func TestExpectedScoreBoundaries(t *testing.T) {
	// Values outside [-800,800] should be clamped.
	if got := expectedScore(-1000); got != 0.0 {
		t.Errorf("expectedScore(-1000) = %v, want 0.0 (clamped)", got)
	}
	if got := expectedScore(1000); got != 1.0 {
		t.Errorf("expectedScore(1000) = %v, want 1.0 (clamped)", got)
	}
	// Exact boundaries.
	if got := expectedScore(-800); got != 0.0 {
		t.Errorf("expectedScore(-800) = %v, want 0.0", got)
	}
	if got := expectedScore(800); got != 1.0 {
		t.Errorf("expectedScore(800) = %v, want 1.0", got)
	}
	if got := expectedScore(0); got != 0.50 {
		t.Errorf("expectedScore(0) = %v, want 0.50", got)
	}
}

func TestDpFromPSymmetry(t *testing.T) {
	// dpFromP should be antisymmetric: dpFromP(p) = -dpFromP(1-p).
	for p := 0.01; p < 0.50; p += 0.01 {
		dp1 := dpFromP(p)
		dp2 := dpFromP(1.0 - p)
		if math.Abs(dp1+dp2) > 0.1 {
			t.Errorf("dpFromP(%.2f) + dpFromP(%.2f) = %.2f, want 0 (antisymmetry)", p, 1.0-p, dp1+dp2)
		}
	}
}

func TestExpectedScoreSymmetry(t *testing.T) {
	// expectedScore should satisfy: es(dp) + es(-dp) = 1.0.
	for dp := 0.0; dp <= 800; dp += 50 {
		es1 := expectedScore(dp)
		es2 := expectedScore(-dp)
		if math.Abs(es1+es2-1.0) > 0.01 {
			t.Errorf("expectedScore(%.0f) + expectedScore(-%.0f) = %.4f, want 1.0", dp, dp, es1+es2)
		}
	}
}

func TestDpFromPRoundTrip(t *testing.T) {
	// dpFromP and expectedScore should be approximate inverses.
	// For each p in the interior of the table, expectedScore(dpFromP(p)) ≈ p.
	for p := 0.05; p <= 0.95; p += 0.05 {
		dp := dpFromP(p)
		pBack := expectedScore(dp)
		if math.Abs(pBack-p) > 0.02 {
			t.Errorf("expectedScore(dpFromP(%.2f)) = %.4f, want ≈%.2f", p, pBack, p)
		}
	}
}

func TestDpFromPAllTableEntries(t *testing.T) {
	// Verify every exact table entry returns without interpolation.
	for _, entry := range fideTable {
		got := dpFromP(entry.p)
		if got != entry.dp {
			t.Errorf("dpFromP(%.2f) = %v, want %v (exact table entry)", entry.p, got, entry.dp)
		}
	}
}

func TestExpectedScoreAllTableEntries(t *testing.T) {
	// Verify every exact dp in the table returns the corresponding p.
	for _, entry := range fideTable {
		got := expectedScore(entry.dp)
		if math.Abs(got-entry.p) > 0.001 {
			t.Errorf("expectedScore(%.0f) = %v, want %v (exact table entry)", entry.dp, got, entry.p)
		}
	}
}
