// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "testing"

func TestBuildScoreGroups_Empty(t *testing.T) {
	groups := BuildScoreGroups(nil)
	if groups != nil {
		t.Fatalf("expected nil, got %v", groups)
	}
}

func TestBuildScoreGroups_AllSameScore(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", TPN: 1, Score: 0.0, PairingScore: 0.0},
		{ID: "p2", TPN: 2, Score: 0.0, PairingScore: 0.0},
		{ID: "p3", TPN: 3, Score: 0.0, PairingScore: 0.0},
		{ID: "p4", TPN: 4, Score: 0.0, PairingScore: 0.0},
	}
	groups := BuildScoreGroups(players)
	if len(groups) != 1 {
		t.Fatalf("expected 1 score group, got %d", len(groups))
	}
	if groups[0].Score != 0.0 {
		t.Errorf("expected score 0.0, got %.1f", groups[0].Score)
	}
	if len(groups[0].Players) != 4 {
		t.Errorf("expected 4 players, got %d", len(groups[0].Players))
	}
}

func TestBuildScoreGroups_MultipleBrackets(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", TPN: 1, Score: 2.0, PairingScore: 2.0},
		{ID: "p2", TPN: 2, Score: 1.5, PairingScore: 1.5},
		{ID: "p3", TPN: 3, Score: 1.5, PairingScore: 1.5},
		{ID: "p4", TPN: 4, Score: 1.0, PairingScore: 1.0},
		{ID: "p5", TPN: 5, Score: 0.0, PairingScore: 0.0},
	}
	groups := BuildScoreGroups(players)
	if len(groups) != 4 {
		t.Fatalf("expected 4 score groups, got %d", len(groups))
	}
	// Groups should be in descending score order.
	wantScores := []float64{2.0, 1.5, 1.0, 0.0}
	wantCounts := []int{1, 2, 1, 1}
	for i, want := range wantScores {
		if groups[i].Score != want {
			t.Errorf("group %d: want score %.1f, got %.1f", i, want, groups[i].Score)
		}
		if len(groups[i].Players) != wantCounts[i] {
			t.Errorf("group %d: want %d players, got %d", i, wantCounts[i], len(groups[i].Players))
		}
	}
}

func TestBuildScoreGroups_PlayersOrderedByTPN(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", TPN: 1, Score: 1.0, PairingScore: 1.0},
		{ID: "p3", TPN: 3, Score: 1.0, PairingScore: 1.0},
		{ID: "p2", TPN: 2, Score: 1.0, PairingScore: 1.0},
	}
	groups := BuildScoreGroups(players)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	// Players within a group should be ordered by TPN ascending.
	for i := 0; i < len(groups[0].Players)-1; i++ {
		if groups[0].Players[i].TPN > groups[0].Players[i+1].TPN {
			t.Errorf("players not sorted by TPN: %d > %d",
				groups[0].Players[i].TPN, groups[0].Players[i+1].TPN)
		}
	}
}

func TestBuildScoreGroups_UsesPairingScore(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", TPN: 1, Score: 1.0, PairingScore: 2.0},
		{ID: "p2", TPN: 2, Score: 1.0, PairingScore: 1.0},
	}
	groups := BuildScoreGroups(players)
	if len(groups) != 2 {
		t.Fatalf("expected 2 score groups (by PairingScore), got %d", len(groups))
	}
	if groups[0].Score != 2.0 {
		t.Errorf("first group score: got %.1f, want 2.0", groups[0].Score)
	}
	if groups[1].Score != 1.0 {
		t.Errorf("second group score: got %.1f, want 1.0", groups[1].Score)
	}
}

func TestBuildBrackets_HomogeneousOnly(t *testing.T) {
	groups := []ScoreGroup{
		{Score: 2.0, Players: makePlayers("p1", "p2")},
		{Score: 1.0, Players: makePlayers("p3", "p4")},
		{Score: 0.0, Players: makePlayers("p5", "p6")},
	}
	brackets := BuildBrackets(groups)
	if len(brackets) != 3 {
		t.Fatalf("expected 3 brackets, got %d", len(brackets))
	}
	for i, b := range brackets {
		if !b.Homogeneous {
			t.Errorf("bracket %d should be homogeneous", i)
		}
	}
}

func TestBuildBrackets_ScoreDescending(t *testing.T) {
	groups := []ScoreGroup{
		{Score: 2.0, Players: makePlayers("p1")},
		{Score: 0.0, Players: makePlayers("p2")},
		{Score: 1.0, Players: makePlayers("p3")},
	}
	brackets := BuildBrackets(groups)
	if brackets[0].OriginalScore != 2.0 {
		t.Errorf("first bracket score should be 2.0, got %.1f", brackets[0].OriginalScore)
	}
	if brackets[1].OriginalScore != 1.0 {
		t.Errorf("second bracket score should be 1.0, got %.1f", brackets[1].OriginalScore)
	}
	if brackets[2].OriginalScore != 0.0 {
		t.Errorf("third bracket score should be 0.0, got %.1f", brackets[2].OriginalScore)
	}
}

// makePlayers creates PlayerState pointers with given IDs for testing.
func makePlayers(ids ...string) []*PlayerState {
	players := make([]*PlayerState, len(ids))
	for i, id := range ids {
		players[i] = &PlayerState{ID: id, TPN: i + 1}
	}
	return players
}

func TestMergeIntoHeterogeneous(t *testing.T) {
	native := Bracket{
		Players:       makePlayers("p3", "p4"),
		Homogeneous:   true,
		OriginalScore: 1.0,
	}
	floaters := makePlayers("p1", "p2")

	merged := MergeIntoHeterogeneous(native, floaters)

	// Floaters should appear before native players.
	if len(merged.Players) != 4 {
		t.Fatalf("expected 4 players, got %d", len(merged.Players))
	}
	if merged.Players[0].ID != "p1" || merged.Players[1].ID != "p2" {
		t.Errorf("floaters should be first: got %s, %s", merged.Players[0].ID, merged.Players[1].ID)
	}
	if merged.Players[2].ID != "p3" || merged.Players[3].ID != "p4" {
		t.Errorf("native players should follow: got %s, %s", merged.Players[2].ID, merged.Players[3].ID)
	}

	// Should be heterogeneous.
	if merged.Homogeneous {
		t.Error("merged bracket should be heterogeneous")
	}

	// OriginalScore from native bracket.
	if merged.OriginalScore != 1.0 {
		t.Errorf("expected OriginalScore 1.0, got %.1f", merged.OriginalScore)
	}

	// Downfloaters should be set.
	if len(merged.Downfloaters) != 2 {
		t.Fatalf("expected 2 downfloaters, got %d", len(merged.Downfloaters))
	}
	if merged.Downfloaters[0].ID != "p1" || merged.Downfloaters[1].ID != "p2" {
		t.Errorf("downfloaters should match input: got %s, %s", merged.Downfloaters[0].ID, merged.Downfloaters[1].ID)
	}

	// Downfloaters should be a defensive copy (not same slice).
	floaters[0] = &PlayerState{ID: "changed"}
	if merged.Downfloaters[0].ID == "changed" {
		t.Error("Downfloaters should be a defensive copy of input floaters")
	}
}

func TestCollapseBrackets(t *testing.T) {
	upper := Bracket{
		Players:       makePlayers("p1", "p2"),
		Homogeneous:   true,
		OriginalScore: 2.0,
	}
	lower := Bracket{
		Players:       makePlayers("p3", "p4"),
		Homogeneous:   true,
		OriginalScore: 1.0,
	}

	collapsed := CollapseBrackets(upper, lower)

	// Upper players first.
	if len(collapsed.Players) != 4 {
		t.Fatalf("expected 4 players, got %d", len(collapsed.Players))
	}
	if collapsed.Players[0].ID != "p1" || collapsed.Players[1].ID != "p2" {
		t.Errorf("upper players should be first: got %s, %s", collapsed.Players[0].ID, collapsed.Players[1].ID)
	}
	if collapsed.Players[2].ID != "p3" || collapsed.Players[3].ID != "p4" {
		t.Errorf("lower players should follow: got %s, %s", collapsed.Players[2].ID, collapsed.Players[3].ID)
	}

	// Uses lower's OriginalScore.
	if collapsed.OriginalScore != 1.0 {
		t.Errorf("expected OriginalScore 1.0 (from lower), got %.1f", collapsed.OriginalScore)
	}

	// Should be heterogeneous.
	if collapsed.Homogeneous {
		t.Error("collapsed bracket should be heterogeneous")
	}
}
