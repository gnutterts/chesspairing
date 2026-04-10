// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import "github.com/gnutterts/chesspairing/pairing/swisslib"

// CanReceiveColor returns true if the player can legally be assigned the
// given colour without violating Art. 2.1 (compatibility statement):
//   - No player shall have the same colour in three successive rounds (Art. 5.1.1)
//   - No player shall have three more of one colour than the other (Art. 5.1.2)
func CanReceiveColor(p *swisslib.PlayerState, c swisslib.Color) bool {
	history := filterPlayed(p.ColorHistory)

	// Check 3 consecutive: would this be the 3rd in a row?
	n := len(history)
	if n >= 2 && history[n-1] == c && history[n-2] == c {
		return false
	}

	// Check imbalance: would the difference become 3+?
	whites, blacks := 0, 0
	for _, h := range history {
		if h == swisslib.ColorWhite {
			whites++
		} else {
			blacks++
		}
	}
	if c == swisslib.ColorWhite {
		whites++
	} else {
		blacks++
	}
	diff := whites - blacks
	if diff < 0 {
		diff = -diff
	}
	if diff >= 3 {
		return false
	}

	return true
}

// IsCompatible returns true if two players can be legally paired per Art. 2.1.
// Two players are compatible if:
//  1. They have not already played each other.
//  2. They are not in the forbidden pairs list.
//  3. There exists at least one legal colour assignment (a=W,b=B or a=B,b=W)
//     that doesn't violate the consecutive or imbalance constraints for either player.
func IsCompatible(a, b *swisslib.PlayerState, forbidden map[[2]string]bool) bool {
	// Already played?
	if swisslib.HasPlayed(a, b) {
		return false
	}

	// Forbidden pair?
	if forbidden != nil {
		if forbidden[[2]string{a.ID, b.ID}] || forbidden[[2]string{b.ID, a.ID}] {
			return false
		}
	}

	// Check if at least one colour assignment works.
	// Assignment 1: a=White, b=Black
	assign1 := CanReceiveColor(a, swisslib.ColorWhite) && CanReceiveColor(b, swisslib.ColorBlack)
	// Assignment 2: a=Black, b=White
	assign2 := CanReceiveColor(a, swisslib.ColorBlack) && CanReceiveColor(b, swisslib.ColorWhite)

	return assign1 || assign2
}

// filterPlayed returns only non-None colours from history.
func filterPlayed(history []swisslib.Color) []swisslib.Color {
	var played []swisslib.Color
	for _, c := range history {
		if c != swisslib.ColorNone {
			played = append(played, c)
		}
	}
	return played
}
