// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"testing"
)

func TestSelectUpfloater_LowestTPNFirst(t *testing.T) {
	bracket := []*ParticipantState{
		{ID: "p3", TPN: 3, Score: 0.0},
		{ID: "p4", TPN: 4, Score: 0.0},
		{ID: "p5", TPN: 5, Score: 0.0},
	}
	targetBracket := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 1.0},
		{ID: "p2", TPN: 2, Score: 1.0},
	}

	floater := SelectUpfloater(bracket, targetBracket, nil)
	if floater == nil || floater.ID != "p3" {
		t.Errorf("expected p3 (lowest TPN first, Art. 3.5.4), got %v", floater)
	}
}

func TestSelectUpfloater_SkipsPlayedAll(t *testing.T) {
	// p3 has already played both members of the target bracket.
	bracket := []*ParticipantState{
		{ID: "p3", TPN: 3, Score: 0.0, Opponents: []string{"p1", "p2"}},
		{ID: "p4", TPN: 4, Score: 0.0},
		{ID: "p5", TPN: 5, Score: 0.0},
	}
	targetBracket := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 1.0},
		{ID: "p2", TPN: 2, Score: 1.0},
	}

	floater := SelectUpfloater(bracket, targetBracket, nil)
	// p3 can't go (played both), so p4 should be selected.
	if floater == nil || floater.ID != "p4" {
		t.Errorf("expected p4 (p3 played all), got %v", floater)
	}
}

func TestSelectUpfloater_ForbiddenPairs(t *testing.T) {
	bracket := []*ParticipantState{
		{ID: "p3", TPN: 3, Score: 0.0},
		{ID: "p4", TPN: 4, Score: 0.0},
	}
	targetBracket := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 1.0},
	}
	// p4 is forbidden from playing p1.
	forbidden := map[[2]string]bool{
		{"p4", "p1"}: true,
	}

	floater := SelectUpfloater(bracket, targetBracket, forbidden)
	// p4 can't play p1, so p3 should be selected.
	if floater == nil || floater.ID != "p3" {
		t.Errorf("expected p3 (p4 forbidden), got %v", floater)
	}
}

func TestSelectUpfloater_NoValidFloater(t *testing.T) {
	// Both have played all members of target bracket.
	bracket := []*ParticipantState{
		{ID: "p3", TPN: 3, Score: 0.0, Opponents: []string{"p1"}},
		{ID: "p4", TPN: 4, Score: 0.0, Opponents: []string{"p1"}},
	}
	targetBracket := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 1.0},
	}

	floater := SelectUpfloater(bracket, targetBracket, nil)
	if floater != nil {
		t.Errorf("expected nil (no valid floater), got %v", floater)
	}
}

func TestSelectUpfloater_Empty(t *testing.T) {
	floater := SelectUpfloater(nil, nil, nil)
	if floater != nil {
		t.Errorf("expected nil for empty bracket, got %v", floater)
	}
}

// TestSelectUpfloater_LeavesPairableRemainder is the C.04.5 D7 case: floating
// a (the lowest TPN) would leave b and c, who have already met, so the next
// candidate floats up (Art. 3.5.5 with C6).
func TestSelectUpfloater_LeavesPairableRemainder(t *testing.T) {
	bracket := []*ParticipantState{
		{ID: "a", TPN: 1, Score: 1.0, Opponents: []string{"d"}},
		{ID: "b", TPN: 2, Score: 1.0, Opponents: []string{"c"}},
		{ID: "c", TPN: 3, Score: 1.0, Opponents: []string{"b"}},
	}
	targetBracket := []*ParticipantState{{ID: "e", TPN: 5, Score: 1.5}}

	floater := SelectUpfloater(bracket, targetBracket, nil)
	if floater == nil || floater.ID != "b" {
		t.Errorf("expected b (a would leave the played pair b-c), got %v", floater)
	}
}
