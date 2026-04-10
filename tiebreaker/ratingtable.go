// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import "sort"

// FIDE B.02 conversion table: fractional score (p) → rating difference (dp).
// p is expressed as a fraction 0.0 to 1.0 (e.g., 0.75 = 75%).
// Entries are sorted by ascending p.
//
// Source: FIDE Handbook, B.02 Rating Regulations, Table 8.1b.
var fideTable = []struct {
	p  float64
	dp float64
}{
	{0.00, -800},
	{0.01, -677},
	{0.02, -589},
	{0.03, -538},
	{0.04, -501},
	{0.05, -470},
	{0.06, -444},
	{0.07, -422},
	{0.08, -401},
	{0.09, -383},
	{0.10, -366},
	{0.11, -351},
	{0.12, -336},
	{0.13, -322},
	{0.14, -309},
	{0.15, -296},
	{0.16, -284},
	{0.17, -273},
	{0.18, -262},
	{0.19, -251},
	{0.20, -240},
	{0.21, -230},
	{0.22, -220},
	{0.23, -211},
	{0.24, -202},
	{0.25, -193},
	{0.26, -184},
	{0.27, -175},
	{0.28, -166},
	{0.29, -158},
	{0.30, -149},
	{0.31, -141},
	{0.32, -133},
	{0.33, -125},
	{0.34, -117},
	{0.35, -110},
	{0.36, -102},
	{0.37, -95},
	{0.38, -87},
	{0.39, -80},
	{0.40, -72},
	{0.41, -65},
	{0.42, -57},
	{0.43, -50},
	{0.44, -43},
	{0.45, -36},
	{0.46, -29},
	{0.47, -21},
	{0.48, -14},
	{0.49, -7},
	{0.50, 0},
	{0.51, 7},
	{0.52, 14},
	{0.53, 21},
	{0.54, 29},
	{0.55, 36},
	{0.56, 43},
	{0.57, 50},
	{0.58, 57},
	{0.59, 65},
	{0.60, 72},
	{0.61, 80},
	{0.62, 87},
	{0.63, 95},
	{0.64, 102},
	{0.65, 110},
	{0.66, 117},
	{0.67, 125},
	{0.68, 133},
	{0.69, 141},
	{0.70, 149},
	{0.71, 158},
	{0.72, 166},
	{0.73, 175},
	{0.74, 184},
	{0.75, 193},
	{0.76, 202},
	{0.77, 211},
	{0.78, 220},
	{0.79, 230},
	{0.80, 240},
	{0.81, 251},
	{0.82, 262},
	{0.83, 273},
	{0.84, 284},
	{0.85, 296},
	{0.86, 309},
	{0.87, 322},
	{0.88, 336},
	{0.89, 351},
	{0.90, 366},
	{0.91, 383},
	{0.92, 401},
	{0.93, 422},
	{0.94, 444},
	{0.95, 470},
	{0.96, 501},
	{0.97, 538},
	{0.98, 589},
	{0.99, 677},
	{1.00, 800},
}

// dpFromP returns the FIDE rating difference for a given fractional score.
// Exact table entries are returned directly; intermediate values are linearly
// interpolated between adjacent entries.
//
// p must be in [0.0, 1.0]. Values outside this range are clamped.
func dpFromP(p float64) float64 {
	if p <= 0.0 {
		return -800
	}
	if p >= 1.0 {
		return 800
	}

	// Binary search for the position in the table.
	idx := sort.Search(len(fideTable), func(i int) bool {
		return fideTable[i].p >= p
	})

	// Exact match.
	if fideTable[idx].p == p {
		return fideTable[idx].dp
	}

	// Interpolate between idx-1 and idx.
	lo := fideTable[idx-1]
	hi := fideTable[idx]
	fraction := (p - lo.p) / (hi.p - lo.p)
	return lo.dp + fraction*(hi.dp-lo.dp)
}

// expectedScore returns the expected fractional score for a given rating
// difference. This is the inverse lookup of dpFromP.
//
// dp is clamped to [-800, 800].
func expectedScore(dp float64) float64 {
	if dp <= -800 {
		return 0.0
	}
	if dp >= 800 {
		return 1.0
	}

	// Binary search by dp value.
	idx := sort.Search(len(fideTable), func(i int) bool {
		return fideTable[i].dp >= dp
	})

	if idx >= len(fideTable) {
		return 1.0
	}

	// Exact match.
	if fideTable[idx].dp == dp {
		return fideTable[idx].p
	}

	if idx == 0 {
		return 0.0
	}

	// Interpolate.
	lo := fideTable[idx-1]
	hi := fideTable[idx]
	fraction := (dp - lo.dp) / (hi.dp - lo.dp)
	return lo.p + fraction*(hi.p-lo.p)
}
