// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestAllocateColor_BothNoHistory(t *testing.T) {
	a := &swisslib.PlayerState{ID: "a", TPN: 1, ColorHistory: nil}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, ColorHistory: nil}

	// Board 1, no topSeedColor -> higher ranked gets white.
	wID, bID := AllocateColor(a, b, 1, nil)
	if wID != "a" || bID != "b" {
		t.Errorf("expected a=white, b=black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_CompatiblePreferences(t *testing.T) {
	a := &swisslib.PlayerState{
		ID: "a", TPN: 1,
		ColorHistory: []swisslib.Color{swisslib.ColorWhite}, // prefers black
	}
	b := &swisslib.PlayerState{
		ID: "b", TPN: 2,
		ColorHistory: []swisslib.Color{swisslib.ColorBlack}, // prefers white
	}

	wID, bID := AllocateColor(a, b, 1, nil)
	if wID != "b" || bID != "a" {
		t.Errorf("expected b=white, a=black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_StrongerPreferenceWins(t *testing.T) {
	a := &swisslib.PlayerState{
		ID: "a", TPN: 1,
		// Played W,W,B -> whites=2, blacks=1 -> prefers Black (strong, imbalance=1)
		ColorHistory: []swisslib.Color{swisslib.ColorWhite, swisslib.ColorWhite, swisslib.ColorBlack},
	}
	b := &swisslib.PlayerState{
		ID: "b", TPN: 2,
		// Played W,B -> balanced -> mild preference for White (alternate last)
		ColorHistory: []swisslib.Color{swisslib.ColorWhite, swisslib.ColorBlack},
	}

	wID, bID := AllocateColor(a, b, 1, nil)
	// a has strong black preference, b has mild white preference -> compatible.
	if wID != "b" || bID != "a" {
		t.Errorf("expected b=white, a=black; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_Board2BlackTopSeed(t *testing.T) {
	a := &swisslib.PlayerState{ID: "a", TPN: 1, ColorHistory: nil}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, ColorHistory: nil}

	c := swisslib.ColorBlack
	wID, bID := AllocateColor(a, b, 2, &c)
	// topSeedColor=black, board 2 -> alternated -> higher ranked gets white.
	if wID != "a" || bID != "b" {
		t.Errorf("expected a=white, b=black on board 2 with black top seed; got w=%s, b=%s", wID, bID)
	}
}
