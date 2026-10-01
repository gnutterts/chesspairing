// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestDubovByeSelector(t *testing.T) {
	sel := DubovByeSelector{}

	t.Run("selects lowest-ranked eligible", func(t *testing.T) {
		players := []*swisslib.PlayerState{
			{ID: "a", TPN: 1, Score: 2.0},
			{ID: "b", TPN: 2, Score: 1.0},
			{ID: "c", TPN: 3, Score: 1.0},
		}
		got := sel.SelectBye(players)
		if got == nil || got.ID != "c" {
			t.Errorf("expected c (highest TPN in lowest score), got %v", got)
		}
	})

	t.Run("skips players who already had bye", func(t *testing.T) {
		players := []*swisslib.PlayerState{
			{ID: "a", TPN: 1, Score: 1.0},
			{ID: "b", TPN: 2, Score: 0.0, PABIneligible: swisslib.PABIneligibility{PriorPAB: true}},
			{ID: "c", TPN: 3, Score: 0.0},
		}
		got := sel.SelectBye(players)
		if got == nil || got.ID != "c" {
			t.Errorf("expected c (lowest score, eligible), got %v", got)
		}
	})

	t.Run("all have had bye returns nil", func(t *testing.T) {
		players := []*swisslib.PlayerState{
			{ID: "a", TPN: 1, Score: 1.0, PABIneligible: swisslib.PABIneligibility{PriorPAB: true}},
			{ID: "b", TPN: 2, Score: 0.0, PABIneligible: swisslib.PABIneligibility{PriorPAB: true}},
		}
		got := sel.SelectBye(players)
		if got != nil {
			t.Errorf("expected nil, got %v", got.ID)
		}
	})

	t.Run("single player", func(t *testing.T) {
		players := []*swisslib.PlayerState{
			{ID: "a", TPN: 1, Score: 0.0},
		}
		got := sel.SelectBye(players)
		if got == nil || got.ID != "a" {
			t.Errorf("expected a, got %v", got)
		}
	})
}

func TestDubovByeSelector_Interface(t *testing.T) {
	var _ swisslib.ByeSelector = DubovByeSelector{}
}
