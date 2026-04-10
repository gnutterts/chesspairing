// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"testing"
)

func TestBuildScoreGroups_BasicGrouping(t *testing.T) {
	participants := []ParticipantState{
		{ID: "p1", TPN: 1, Score: 2.0},
		{ID: "p2", TPN: 2, Score: 2.0},
		{ID: "p3", TPN: 3, Score: 1.0},
		{ID: "p4", TPN: 4, Score: 0.0},
	}

	groups := BuildScoreGroups(participants)
	if len(groups) != 3 {
		t.Fatalf("expected 3 score groups, got %d", len(groups))
	}

	// Descending score order.
	if groups[0].Score != 2.0 {
		t.Errorf("group 0 score: expected 2.0, got %f", groups[0].Score)
	}
	if groups[1].Score != 1.0 {
		t.Errorf("group 1 score: expected 1.0, got %f", groups[1].Score)
	}
	if groups[2].Score != 0.0 {
		t.Errorf("group 2 score: expected 0.0, got %f", groups[2].Score)
	}

	// Group 0 should have 2 players.
	if len(groups[0].Participants) != 2 {
		t.Errorf("group 0: expected 2 participants, got %d", len(groups[0].Participants))
	}
}

func TestBuildScoreGroups_TPNOrdering(t *testing.T) {
	participants := []ParticipantState{
		{ID: "p3", TPN: 3, Score: 1.0},
		{ID: "p1", TPN: 1, Score: 1.0},
		{ID: "p2", TPN: 2, Score: 1.0},
	}

	groups := BuildScoreGroups(participants)
	if len(groups) != 1 {
		t.Fatalf("expected 1 score group, got %d", len(groups))
	}

	// Within the group, participants should be ordered by TPN ascending.
	if groups[0].Participants[0].TPN != 1 {
		t.Errorf("first participant TPN: expected 1, got %d", groups[0].Participants[0].TPN)
	}
	if groups[0].Participants[2].TPN != 3 {
		t.Errorf("third participant TPN: expected 3, got %d", groups[0].Participants[2].TPN)
	}
}

func TestBuildScoreGroups_Empty(t *testing.T) {
	groups := BuildScoreGroups(nil)
	if groups != nil {
		t.Errorf("expected nil for empty input, got %v", groups)
	}
}

func TestBuildScoreGroups_SinglePlayer(t *testing.T) {
	participants := []ParticipantState{
		{ID: "p1", TPN: 1, Score: 0.0},
	}

	groups := BuildScoreGroups(participants)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Participants) != 1 {
		t.Errorf("expected 1 participant in group, got %d", len(groups[0].Participants))
	}
}

func TestBuildScoreGroups_HalfPointScores(t *testing.T) {
	participants := []ParticipantState{
		{ID: "p1", TPN: 1, Score: 1.5},
		{ID: "p2", TPN: 2, Score: 1.0},
		{ID: "p3", TPN: 3, Score: 1.5},
		{ID: "p4", TPN: 4, Score: 0.5},
	}

	groups := BuildScoreGroups(participants)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].Score != 1.5 || len(groups[0].Participants) != 2 {
		t.Errorf("group 0: expected score 1.5 with 2 players, got %f with %d", groups[0].Score, len(groups[0].Participants))
	}
}
