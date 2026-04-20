// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package factory_test

import (
	"sort"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/factory"
)

func TestNewPairer_AllNames(t *testing.T) {
	for _, name := range factory.PairerNames() {
		t.Run(name, func(t *testing.T) {
			p, err := factory.NewPairer(name, nil)
			if err != nil {
				t.Fatalf("NewPairer(%q): %v", name, err)
			}
			if p == nil {
				t.Fatalf("NewPairer(%q): nil pairer", name)
			}
		})
	}
}

func TestNewPairer_AcceptsAliases(t *testing.T) {
	cases := []struct{ name, want string }{
		{"FIDE-Dutch", "dutch"},
		{"Round-Robin", "roundrobin"},
		{"rr", "roundrobin"},
		{"DOUBLE-SWISS", "doubleswiss"},
	}
	for _, c := range cases {
		p, err := factory.NewPairer(c.name, nil)
		if err != nil {
			t.Errorf("NewPairer(%q): %v", c.name, err)
			continue
		}
		if p == nil {
			t.Errorf("NewPairer(%q): nil", c.name)
		}
	}
}

func TestNewPairer_Unknown(t *testing.T) {
	_, err := factory.NewPairer("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown pairer")
	}
}

func TestNewPairer_NilOpts(t *testing.T) {
	p, err := factory.NewPairer("dutch", nil)
	if err != nil {
		t.Fatalf("NewPairer with nil opts: %v", err)
	}
	if p == nil {
		t.Fatal("nil pairer")
	}
}

func TestNewScorer_AllNames(t *testing.T) {
	for _, name := range factory.ScorerNames() {
		t.Run(name, func(t *testing.T) {
			s, err := factory.NewScorer(name, nil)
			if err != nil {
				t.Fatalf("NewScorer(%q): %v", name, err)
			}
			if s == nil {
				t.Fatalf("NewScorer(%q): nil scorer", name)
			}
		})
	}
}

func TestNewScorer_Unknown(t *testing.T) {
	_, err := factory.NewScorer("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown scorer")
	}
}

func TestNewTieBreaker_KnownIDs(t *testing.T) {
	ids := factory.TieBreakerIDs()
	if len(ids) == 0 {
		t.Fatal("no registered tiebreakers")
	}
	for _, id := range ids {
		t.Run(id, func(t *testing.T) {
			tb, err := factory.NewTieBreaker(id)
			if err != nil {
				t.Fatalf("NewTieBreaker(%q): %v", id, err)
			}
			if tb == nil {
				t.Fatalf("NewTieBreaker(%q): nil", id)
			}
			if tb.ID() != id {
				t.Errorf("ID() = %q, want %q", tb.ID(), id)
			}
		})
	}
}

func TestNewTieBreaker_Unknown(t *testing.T) {
	_, err := factory.NewTieBreaker("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown tiebreaker")
	}
}

func TestPairerNames_Sorted(t *testing.T) {
	names := factory.PairerNames()
	if !sort.StringsAreSorted(names) {
		t.Errorf("PairerNames not sorted: %v", names)
	}
}

func TestScorerNames_Sorted(t *testing.T) {
	names := factory.ScorerNames()
	if !sort.StringsAreSorted(names) {
		t.Errorf("ScorerNames not sorted: %v", names)
	}
}

func TestTieBreakerIDs_Sorted(t *testing.T) {
	ids := factory.TieBreakerIDs()
	if !sort.StringsAreSorted(ids) {
		t.Errorf("TieBreakerIDs not sorted: %v", ids)
	}
}

// TestNewScorer_StandardExcusedAndClubCommitment confirms the new
// pointExcused and pointClubCommitment option keys round-trip through
// the factory into the standard scorer.
func TestNewScorer_StandardExcusedAndClubCommitment(t *testing.T) {
	opts := map[string]any{
		"pointExcused":        0.25,
		"pointClubCommitment": 0.75,
	}
	s, err := factory.NewScorer("standard", opts)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	if s == nil {
		t.Fatal("nil scorer")
	}

	excused := chesspairing.ByeExcused
	pts := s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{ByeType: &excused})
	if pts != 0.25 {
		t.Errorf("excused points = %v, want 0.25", pts)
	}

	club := chesspairing.ByeClubCommitment
	pts = s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{ByeType: &club})
	if pts != 0.75 {
		t.Errorf("club commitment points = %v, want 0.75", pts)
	}
}
