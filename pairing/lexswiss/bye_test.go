// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"context"
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

	got := assignPAB(t, participants)
	if got == nil || got.ID != "p5" {
		t.Errorf("expected p5 (lowest score=0, highest TPN=5), got %v", got)
	}
}

func TestAssignPAB_SkipsAlreadyReceivedBye(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 1.0},
		{ID: "p2", TPN: 2, Score: 0.0, PABIneligible: PABIneligibility{PriorPAB: true}},
		{ID: "p3", TPN: 3, Score: 0.0},
	}

	got := assignPAB(t, participants)
	if got == nil || got.ID != "p3" {
		t.Errorf("expected p3 (TPN=3, no bye yet), got %v", got)
	}
}

func TestAssignPAB_AllHadBye(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 0.0, PABIneligible: PABIneligibility{PriorPAB: true}},
	}

	got := assignPAB(t, participants)
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

	got := assignPAB(t, participants)
	if got == nil || got.ID != "p2" {
		t.Errorf("expected p2 (highest TPN=5 at score 0), got %v", got)
	}
}

func TestAssignPAB_TieUsesFixedTPNNotRunningRank(t *testing.T) {
	ranked := SortParticipants([]ParticipantState{
		{ID: "running-second", TPN: 8, Score: 0},
		{ID: "running-first", TPN: 2, Score: 0},
		{ID: "leader", TPN: 1, Score: 1},
	})
	participants := make([]*ParticipantState, len(ranked))
	for i := range ranked {
		participants[i] = &ranked[i]
	}
	got := assignPAB(t, participants)
	if got == nil || got.ID != "running-second" {
		t.Errorf("PAB = %v; want highest fixed TPN in tie", got)
	}
}

func TestAssignPAB_LeavesCompletePairing(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "a", TPN: 1, Score: 1, Opponents: []string{"c", "d"}},
		{ID: "b", TPN: 2, Score: 1},
		{ID: "c", TPN: 3, Score: 1, Opponents: []string{"d"}},
		{ID: "d", TPN: 4, Score: 0},
		{ID: "e", TPN: 5, Score: 0},
	}

	got := assignPAB(t, participants)
	if got == nil || got.ID != "d" {
		t.Errorf("PAB = %v, want d because removing e leaves no C1 pairing", got)
	}
}

func TestAssignPAB_MostMatchesPlayedThenLargestTPN(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Score: 1},
		{ID: "p2", TPN: 2, Score: 0, MatchesPlayed: 3},
		{ID: "p3", TPN: 3, Score: 0, MatchesPlayed: 2},
	}

	got := assignPAB(t, participants)
	if got == nil || got.ID != "p2" {
		t.Errorf("PAB = %v, want p2 with most played matches", got)
	}
	participants[2].MatchesPlayed = 3
	got = assignPAB(t, participants)
	if got == nil || got.ID != "p3" {
		t.Errorf("PAB = %v, want p3 with largest TPN", got)
	}
}

func TestAssignPAB_Empty(t *testing.T) {
	got := assignPAB(t, nil)
	if got != nil {
		t.Errorf("expected nil for empty, got %v", got)
	}
}

func assignPAB(t *testing.T, participants []*ParticipantState) *ParticipantState {
	t.Helper()
	got, err := AssignPAB(context.Background(), participants, nil)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
