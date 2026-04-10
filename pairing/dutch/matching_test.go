// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestSplitS1S2(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "p1", TPN: 1},
		{ID: "p2", TPN: 2},
		{ID: "p3", TPN: 3},
		{ID: "p4", TPN: 4},
	}
	s1, s2 := SplitS1S2(players)
	if len(s1) != 2 || len(s2) != 2 {
		t.Errorf("expected s1=2, s2=2, got s1=%d, s2=%d", len(s1), len(s2))
	}
	if s1[0].ID != "p1" || s1[1].ID != "p2" {
		t.Errorf("s1 should be [p1, p2], got [%s, %s]", s1[0].ID, s1[1].ID)
	}
}

func TestSplitS1S2Heterogeneous(t *testing.T) {
	bracket := swisslib.Bracket{
		Players: []*swisslib.PlayerState{
			{ID: "f1", TPN: 1},
			{ID: "n1", TPN: 2},
			{ID: "n2", TPN: 3},
		},
		Homogeneous:  false,
		Downfloaters: []*swisslib.PlayerState{{ID: "f1", TPN: 1}},
	}
	s1, s2 := SplitS1S2Heterogeneous(bracket)
	if len(s1) != 1 || s1[0].ID != "f1" {
		t.Errorf("s1 should be [f1], got %v", s1)
	}
	if len(s2) != 2 {
		t.Errorf("s2 should have 2 players, got %d", len(s2))
	}
}

func TestBuildCandidate(t *testing.T) {
	s1 := []*swisslib.PlayerState{
		{ID: "p1", TPN: 1, Score: 3.0},
		{ID: "p2", TPN: 2, Score: 3.0},
	}
	s2 := []*swisslib.PlayerState{
		{ID: "p3", TPN: 3, Score: 2.0},
		{ID: "p4", TPN: 4, Score: 2.0},
		{ID: "p5", TPN: 5, Score: 2.0},
	}

	downfloaterIDs := map[string]bool{"p1": true, "p2": true}
	cand := buildCandidate(s1, s2, downfloaterIDs, 2.0)

	if len(cand.Pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(cand.Pairs))
	}
	if len(cand.Residuals) != 1 {
		t.Errorf("expected 1 residual, got %d", len(cand.Residuals))
	}
	if cand.Residuals[0].ID != "p5" {
		t.Errorf("expected residual p5, got %s", cand.Residuals[0].ID)
	}
	if cand.BracketScore != 2.0 {
		t.Errorf("expected bracket score 2.0, got %f", cand.BracketScore)
	}
}

func TestGenerateTranspositions(t *testing.T) {
	s2 := []*swisslib.PlayerState{
		{ID: "a", TPN: 1},
		{ID: "b", TPN: 2},
		{ID: "c", TPN: 3},
	}
	perms := GenerateTranspositions(s2, 100)
	// 3! = 6 permutations.
	if len(perms) != 6 {
		t.Errorf("expected 6 permutations, got %d", len(perms))
	}
}

func TestRecordFloats(t *testing.T) {
	players := map[string]*swisslib.PlayerState{
		"p1": {ID: "p1", Score: 3.0},
		"p2": {ID: "p2", Score: 3.0},
		"p3": {ID: "p3", Score: 2.0},
		"p4": {ID: "p4", Score: 2.0},
	}

	// Simulate: p1 and p2 float down from bracket 3.0 to bracket 2.0.
	floaters := []*swisslib.PlayerState{players["p1"], players["p2"]}
	paired := []*swisslib.PlayerState{players["p3"], players["p4"]}

	recordFloats(floaters, paired, players)

	// Floaters should have FloatDown appended.
	if len(players["p1"].FloatHistory) != 1 || players["p1"].FloatHistory[0] != swisslib.FloatDown {
		t.Errorf("p1 should have FloatDown, got %v", players["p1"].FloatHistory)
	}
	if len(players["p2"].FloatHistory) != 1 || players["p2"].FloatHistory[0] != swisslib.FloatDown {
		t.Errorf("p2 should have FloatDown, got %v", players["p2"].FloatHistory)
	}

	// Paired native players should have FloatNone appended.
	if len(players["p3"].FloatHistory) != 1 || players["p3"].FloatHistory[0] != swisslib.FloatNone {
		t.Errorf("p3 should have FloatNone, got %v", players["p3"].FloatHistory)
	}
	if len(players["p4"].FloatHistory) != 1 || players["p4"].FloatHistory[0] != swisslib.FloatNone {
		t.Errorf("p4 should have FloatNone, got %v", players["p4"].FloatHistory)
	}
}

func TestRecordFloats_EmptyFloaters(t *testing.T) {
	players := map[string]*swisslib.PlayerState{
		"p1": {ID: "p1", Score: 2.0},
		"p2": {ID: "p2", Score: 2.0},
	}

	paired := []*swisslib.PlayerState{players["p1"], players["p2"]}

	recordFloats(nil, paired, players)

	// All paired players should have FloatNone.
	if len(players["p1"].FloatHistory) != 1 || players["p1"].FloatHistory[0] != swisslib.FloatNone {
		t.Errorf("p1 should have FloatNone, got %v", players["p1"].FloatHistory)
	}
}

func TestRecordFloats_NoDoubleCounting(t *testing.T) {
	players := map[string]*swisslib.PlayerState{
		"p1": {ID: "p1", Score: 3.0, FloatHistory: []swisslib.Float{swisslib.FloatDown}},
	}

	// p1 floats down again in a second bracket processing.
	floaters := []*swisslib.PlayerState{players["p1"]}

	recordFloats(floaters, nil, players)

	if len(players["p1"].FloatHistory) != 2 {
		t.Errorf("p1 should have 2 float entries, got %d", len(players["p1"].FloatHistory))
	}
	if players["p1"].FloatHistory[1] != swisslib.FloatDown {
		t.Errorf("p1 second entry should be FloatDown, got %v", players["p1"].FloatHistory[1])
	}
}

func TestGenerateExchanges(t *testing.T) {
	s1 := []*swisslib.PlayerState{
		{ID: "a", TPN: 1},
		{ID: "b", TPN: 2},
	}
	s2 := []*swisslib.PlayerState{
		{ID: "c", TPN: 3},
		{ID: "d", TPN: 4},
	}
	exchanges := GenerateExchanges(s1, s2)
	// 1-swap: C(2,1)*C(2,1) = 4; 2-swap: C(2,2)*C(2,2) = 1. Total = 5.
	if len(exchanges) != 5 {
		t.Errorf("expected 5 exchanges, got %d", len(exchanges))
	}
}
