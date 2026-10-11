// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import "github.com/gnutterts/chesspairing/pairing/swisslib"

// AllocateColor implements Dubov Article 5.2. Article 1.6.4 is applied by
// preference, rather than changing swisslib's shared colour behaviour.
func AllocateColor(a, b *swisslib.PlayerState, _ int, initialColor *swisslib.Color) (string, string) {
	if swisslib.GamesPlayed(a) == 0 && swisslib.GamesPlayed(b) == 0 {
		higher, lower := higherRanked(a, b), a
		if higher == a {
			lower = b
		}
		initialWhite := initialColor == nil || *initialColor == swisslib.ColorWhite
		if (swisslib.EffectivePairingNumber(higher)%2 != 0) == initialWhite {
			return higher.ID, lower.ID
		}
		return lower.ID, higher.ID
	}
	preferenceA, preferenceB := preference(a), preference(b)
	grantA := func(color swisslib.Color) (string, string) {
		if color == swisslib.ColorWhite {
			return a.ID, b.ID
		}
		return b.ID, a.ID
	}
	grantB := func(color swisslib.Color) (string, string) {
		if color == swisslib.ColorWhite {
			return b.ID, a.ID
		}
		return a.ID, b.ID
	}
	if preferenceA.Color == nil || preferenceB.Color == nil || *preferenceA.Color != *preferenceB.Color {
		if preferenceA.Color != nil {
			return grantA(*preferenceA.Color)
		}
		if preferenceB.Color != nil {
			return grantB(*preferenceB.Color)
		}
	} else {
		if preferenceA.AbsolutePreference && (preferenceA.ColorImbalance > preferenceB.ColorImbalance || !preferenceB.AbsolutePreference) {
			return grantA(*preferenceA.Color)
		}
		if preferenceB.AbsolutePreference && (preferenceB.ColorImbalance > preferenceA.ColorImbalance || !preferenceA.AbsolutePreference) {
			return grantB(*preferenceB.Color)
		}
		if preferenceA.StrongPreference && !preferenceB.StrongPreference {
			return grantA(*preferenceA.Color)
		}
		if preferenceB.StrongPreference && !preferenceA.StrongPreference {
			return grantB(*preferenceB.Color)
		}
		if color := dubovColorDifference(preferenceA.PlayedColors, preferenceB.PlayedColors); color != nil {
			return grantA(*color)
		}
	}
	higher := higherRanked(a, b)
	if higher == a {
		return grantA(*preferenceA.Color)
	}
	return grantB(*preferenceB.Color)
}

func higherRanked(a, b *swisslib.PlayerState) *swisslib.PlayerState {
	if b.Score > a.Score || (b.Score == a.Score && swisslib.EffectivePairingNumber(b) < swisslib.EffectivePairingNumber(a)) {
		return b
	}
	return a
}

func dubovColorDifference(a, b []swisslib.Color) *swisslib.Color {
	for i, j := len(a)-1, len(b)-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		if a[i] != b[j] {
			color := b[j]
			return &color
		}
	}
	return nil
}
