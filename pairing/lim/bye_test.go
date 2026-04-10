// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestLimByeSelector_LowestScoreHighestTPN(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 2.0},
		{ID: "b", TPN: 2, Score: 1.0},
		{ID: "c", TPN: 3, Score: 1.0},
		{ID: "d", TPN: 4, Score: 0.0},
		{ID: "e", TPN: 5, Score: 0.0},
	}

	sel := LimByeSelector{}
	got := sel.SelectBye(players)
	if got == nil || got.ID != "e" {
		t.Errorf("expected player 'e' (lowest score=0, highest TPN=5), got %v", got)
	}
}

func TestLimByeSelector_SkipsAlreadyReceivedBye(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 1.0},
		{ID: "b", TPN: 2, Score: 0.0, ByeReceived: true},
		{ID: "c", TPN: 3, Score: 0.0},
	}

	sel := LimByeSelector{}
	got := sel.SelectBye(players)
	if got == nil || got.ID != "c" {
		t.Errorf("expected player 'c' (TPN=3, no bye), got %v", got)
	}
}

func TestLimByeSelector_AllHadBye(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 0.0, ByeReceived: true},
	}

	sel := LimByeSelector{}
	got := sel.SelectBye(players)
	if got != nil {
		t.Errorf("expected nil when all have had bye, got %v", got)
	}
}

func TestLimByeSelector_TiedScoreUsesHighestTPN(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 0.0},
		{ID: "b", TPN: 5, Score: 0.0},
		{ID: "c", TPN: 3, Score: 0.0},
	}

	sel := LimByeSelector{}
	got := sel.SelectBye(players)
	if got == nil || got.ID != "b" {
		t.Errorf("expected player 'b' (highest TPN=5 at score 0), got %v", got)
	}
}
