// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"testing"
)

func TestNeedsBye(t *testing.T) {
	if !NeedsBye(5) {
		t.Error("5 participants should need bye")
	}
	if NeedsBye(4) {
		t.Error("4 participants should not need bye")
	}
	if NeedsBye(0) {
		t.Error("0 participants should not need bye")
	}
}

func TestAssignPAB_LowestScoreHighestTPN(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 2.0},
		{ID: "p2", TPN: 2, Score: 1.0},
		{ID: "p3", TPN: 3, Score: 1.0},
		{ID: "p4", TPN: 4, Score: 0.0},
		{ID: "p5", TPN: 5, Score: 0.0},
	}

	got := AssignPAB(participants)
	if got == nil || got.ID != "p5" {
		t.Errorf("expected p5 (lowest score=0, highest TPN=5), got %v", got)
	}
}

func TestAssignPAB_SkipsAlreadyReceivedBye(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 1.0},
		{ID: "p2", TPN: 2, Score: 0.0, ByeReceived: true},
		{ID: "p3", TPN: 3, Score: 0.0},
	}

	got := AssignPAB(participants)
	if got == nil || got.ID != "p3" {
		t.Errorf("expected p3 (TPN=3, no bye yet), got %v", got)
	}
}

func TestAssignPAB_AllHadBye(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 0.0, ByeReceived: true},
	}

	got := AssignPAB(participants)
	if got != nil {
		t.Errorf("expected nil when all have had bye, got %v", got)
	}
}

func TestAssignPAB_TiedScore(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 0.0},
		{ID: "p2", TPN: 5, Score: 0.0},
		{ID: "p3", TPN: 3, Score: 0.0},
	}

	got := AssignPAB(participants)
	if got == nil || got.ID != "p2" {
		t.Errorf("expected p2 (highest TPN=5 at score 0), got %v", got)
	}
}

func TestAssignPAB_Empty(t *testing.T) {
	got := AssignPAB(nil)
	if got != nil {
		t.Errorf("expected nil for empty, got %v", got)
	}
}
