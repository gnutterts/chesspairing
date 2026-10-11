// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestUpfloaterSets_Article354Order(t *testing.T) {
	potential := []*ParticipantState{
		{ID: "1", TPN: 1, Score: 2.5},
		{ID: "2", TPN: 2, Score: 3},
		{ID: "3", TPN: 3, Score: 2.5},
		{ID: "5", TPN: 5, Score: 2.5},
		{ID: "6", TPN: 6, Score: 3},
		{ID: "8", TPN: 8, Score: 3},
	}
	var got [][]int
	err := upfloaterSets(context.Background(), potential, 3, func(set []*ParticipantState) error {
		if sameProfile(set, []*ParticipantState{{Score: 3}, {Score: 3}, {Score: 2.5}}) {
			got = append(got, []int{set[0].TPN, set[1].TPN, set[2].TPN})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{2, 6, 1}, {2, 6, 3}, {2, 6, 5}, {2, 8, 1}, {2, 8, 3}, {2, 8, 5}, {6, 8, 1}, {6, 8, 3}, {6, 8, 5}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Article 3.5.4 sets = %v, want %v", got, want)
	}
}

func TestPairTopDown_C7AvoidsPreviousFloater(t *testing.T) {
	residents := []*ParticipantState{{ID: "a", TPN: 1, Score: 1}}
	potential := []*ParticipantState{
		{ID: "floater", TPN: 2, Score: 0.5, WasFloater: true},
		{ID: "resident", TPN: 3, Score: 0.5},
		{ID: "d", TPN: 4, Score: 0},
		{ID: "e", TPN: 5, Score: 0},
		{ID: "f", TPN: 6, Score: 0},
	}
	budget := topDownSearchLimit
	set, _, err := selectUpfloaterSet(context.Background(), residents, potential, nil, TopDownOptions{ApplyC7: true}, &budget)
	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 1 || set[0].ID != "resident" {
		t.Fatalf("C7 selected %v, want the non-floater resident", set)
	}
}

func TestC6UpfloaterCount(t *testing.T) {
	residents := []*ParticipantState{{ID: "a", TPN: 1, Score: 2}}
	potential := []*ParticipantState{
		{ID: "b", TPN: 2, Score: 1},
		{ID: "c", TPN: 3, Score: 1},
		{ID: "d", TPN: 4, Score: 1},
		{ID: "e", TPN: 5, Score: 0},
		{ID: "f", TPN: 6, Score: 0},
	}
	forbidden := map[[2]string]bool{{"c", "d"}: true, {"d", "c"}: true}
	budget := topDownSearchLimit
	count, err := c6UpfloaterCount(context.Background(), residents, potential, []*ParticipantState{potential[1]}, forbidden, &budget)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("C6 upfloater count after c = %d, want 0", count)
	}
	count, err = c6UpfloaterCount(context.Background(), residents, potential, []*ParticipantState{potential[0]}, forbidden, &budget)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("C6 upfloater count after b = %d, want 2", count)
	}
}

func TestC6UpfloaterCount_EmptyFollowingScoregroup(t *testing.T) {
	residents := []*ParticipantState{{ID: "a", TPN: 1, Score: 2}}
	potential := []*ParticipantState{{ID: "b", TPN: 2, Score: 1}, {ID: "c", TPN: 3, Score: 1}}
	budget := topDownSearchLimit
	count, err := c6UpfloaterCount(context.Background(), residents, potential, potential, nil, &budget)
	if err != nil {
		t.Fatal(err)
	}
	if count != -1 {
		t.Fatalf("C6 upfloater count = %d, want empty scoregroup marker", count)
	}
}

func TestPairTopDown_C6ChoosesBestFollowingBracket(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "a", TPN: 1, Score: 2},
		{ID: "b", TPN: 2, Score: 1},
		{ID: "c", TPN: 3, Score: 1},
		{ID: "d", TPN: 4, Score: 1},
		{ID: "e", TPN: 5, Score: 0},
		{ID: "f", TPN: 6, Score: 0},
	}
	groups := []ScoreGroup{{Score: 2, Participants: participants[:1]}, {Score: 1, Participants: participants[1:4]}, {Score: 0, Participants: participants[4:]}}
	forbidden := map[[2]string]bool{{"c", "d"}: true, {"d", "c"}: true}
	pairs, err := PairTopDown(context.Background(), groups, forbidden, TopDownOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasPair(pairs, "a", "c") || !hasPair(pairs, "b", "d") || !hasPair(pairs, "e", "f") {
		t.Fatalf("C6 pairs = %v, want a-c, b-d, e-f", pairs)
	}
}

func TestC6UpfloaterCount_EmptyFollowingScoregroupUsesSetOrder(t *testing.T) {
	residents := []*ParticipantState{{ID: "a", TPN: 1, Score: 2}}
	potential := []*ParticipantState{
		{ID: "b", TPN: 2, Score: 1},
		{ID: "c", TPN: 3, Score: 1},
		{ID: "e", TPN: 5, Score: 0},
		{ID: "f", TPN: 6, Score: 0},
	}
	first := []*ParticipantState{potential[0], potential[1], potential[2]}
	second := []*ParticipantState{potential[0], potential[1], potential[3]}
	budget := topDownSearchLimit
	count, err := c6UpfloaterCount(context.Background(), residents, potential, first, nil, &budget)
	if err != nil {
		t.Fatal(err)
	}
	if count != -1 || compareSetOrder(first, second) >= 0 {
		t.Fatalf("empty C6 sets = %v and %v, want first set by Article 3.5.4", first, second)
	}
}

func hasPair(pairs [][2]*ParticipantState, first, second string) bool {
	for _, pair := range pairs {
		if (pair[0].ID == first && pair[1].ID == second) || (pair[0].ID == second && pair[1].ID == first) {
			return true
		}
	}
	return false
}

func TestPairTopDown_C3ChoosesLaterSet(t *testing.T) {
	residents := []*ParticipantState{{ID: "top", TPN: 1, Score: 2}}
	potential := []*ParticipantState{
		{ID: "first", TPN: 2, Score: 1},
		{ID: "later", TPN: 3, Score: 1},
		{ID: "blocked", TPN: 4, Score: 1},
	}
	forbidden := map[[2]string]bool{{"later", "blocked"}: true, {"blocked", "later"}: true}
	budget := topDownSearchLimit
	set, _, err := selectUpfloaterSet(context.Background(), residents, potential, forbidden, TopDownOptions{}, &budget)
	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 1 || set[0].ID != "later" {
		t.Fatalf("C3 selected %v, want the later set that leaves a completion", set)
	}
}

func TestPairTopDown_ManyPotentialUpfloaters(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "r1", TPN: 1, Score: 2},
		{ID: "r2", TPN: 2, Score: 2},
		{ID: "r3", TPN: 3, Score: 2},
		{ID: "r4", TPN: 4, Score: 2},
	}
	for i := 0; i < 50; i++ {
		participants = append(participants, &ParticipantState{ID: string(rune('a' + i)), TPN: i + 5, Score: 1 - float64(i)/100})
	}
	forbidden := make(map[[2]string]bool)
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			forbidden[[2]string{participants[i].ID, participants[j].ID}] = true
			forbidden[[2]string{participants[j].ID, participants[i].ID}] = true
		}
	}
	// With no resident-resident edge, two upfloaters cannot complete this
	// four-player group; k=4 is the first possible bracket.
	groups := []ScoreGroup{{Score: 2, Participants: participants[:4]}, {Score: 1, Participants: participants[4:]}}
	_, err := PairTopDown(context.Background(), groups, forbidden, TopDownOptions{})
	if err != nil {
		t.Fatalf("PairTopDown() error = %v, want no search limit", err)
	}
}

func TestPairTopDown_EvenGroupMayNeedUpfloaters(t *testing.T) {
	participants := []*ParticipantState{
		{ID: "a", TPN: 1, Score: 1},
		{ID: "b", TPN: 2, Score: 1},
		{ID: "c", TPN: 3, Score: 0.5, Opponents: []string{"d"}},
		{ID: "d", TPN: 4, Score: 0.5, Opponents: []string{"c"}},
		{ID: "e", TPN: 5, Score: 0},
		{ID: "f", TPN: 6, Score: 0},
	}
	groups := []ScoreGroup{{Score: 1, Participants: participants[:2]}, {Score: 0.5, Participants: participants[2:4]}, {Score: 0, Participants: participants[4:]}}
	pairs, err := PairTopDown(context.Background(), groups, nil, TopDownOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 3 {
		t.Fatalf("pair count = %d, want 3", len(pairs))
	}
	for _, pair := range pairs {
		if (pair[0].ID == "c" && pair[1].ID == "d") || (pair[0].ID == "d" && pair[1].ID == "c") {
			t.Fatal("the already-played even scoregroup was paired without upfloaters")
		}
	}
}

// TestUpfloaterSets_C5ProfileOrder verifies that profiles arrive in C5 order
// (Article 2.3.2: maximise the upfloaters' scores taken in ascending order):
// with scores 3 > 2 > 1 and two upfloaters, 3+3, 3+2, 2+2, 3+1, 2+1, 1+1.
func TestUpfloaterSets_C5ProfileOrder(t *testing.T) {
	potential := []*ParticipantState{
		{ID: "a1", TPN: 1, Score: 3},
		{ID: "a2", TPN: 2, Score: 3},
		{ID: "b1", TPN: 3, Score: 2},
		{ID: "b2", TPN: 4, Score: 2},
		{ID: "c1", TPN: 5, Score: 1},
		{ID: "c2", TPN: 6, Score: 1},
	}
	var got [][2]float64
	err := upfloaterSets(context.Background(), potential, 2, func(set []*ParticipantState) error {
		profile := [2]float64{set[0].Score, set[1].Score}
		if len(got) == 0 || got[len(got)-1] != profile {
			got = append(got, profile)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][2]float64{{3, 3}, {3, 2}, {2, 2}, {3, 1}, {2, 1}, {1, 1}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("C5 profile order = %v, want %v", got, want)
	}
}

// TestPairTopDown_C5PrefersHigherLowestUpfloater: {x, z} (scores 4 and 2)
// and {y1, y2} (3 and 3) both work, and C5 prefers the set whose lowest
// upfloater scores more, so y1 and y2 float.
func TestPairTopDown_C5PrefersHigherLowestUpfloater(t *testing.T) {
	residents := []*ParticipantState{{ID: "r1", TPN: 1, Score: 5}, {ID: "r2", TPN: 2, Score: 5}}
	potential := []*ParticipantState{
		{ID: "x", TPN: 3, Score: 4},
		{ID: "y1", TPN: 4, Score: 3},
		{ID: "y2", TPN: 5, Score: 3},
		{ID: "z", TPN: 6, Score: 2},
	}
	forbidden := map[[2]string]bool{}
	for _, pair := range [][2]string{{"r1", "r2"}, {"y1", "z"}, {"y2", "z"}} {
		forbidden[pair] = true
		forbidden[[2]string{pair[1], pair[0]}] = true
	}
	budget := topDownSearchLimit
	set, _, err := selectUpfloaterSet(context.Background(), residents, potential, forbidden, TopDownOptions{}, &budget)
	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 2 || set[0].ID != "y1" || set[1].ID != "y2" {
		t.Fatalf("C5 selected %v, want y1 and y2", set)
	}
}

// TestUpfloaterSets_VisitedSetsAreIndependent verifies that a set kept by
// the caller is not overwritten by later sets of the same profile.
func TestUpfloaterSets_VisitedSetsAreIndependent(t *testing.T) {
	potential := []*ParticipantState{
		{ID: "a", TPN: 1, Score: 4},
		{ID: "b1", TPN: 2, Score: 3},
		{ID: "b2", TPN: 3, Score: 3},
		{ID: "c1", TPN: 4, Score: 2},
		{ID: "c2", TPN: 5, Score: 2},
		{ID: "d1", TPN: 6, Score: 1},
		{ID: "d2", TPN: 7, Score: 1},
		{ID: "d3", TPN: 8, Score: 1},
	}
	var kept [][]*ParticipantState
	var want [][]string
	err := upfloaterSets(context.Background(), potential, 4, func(set []*ParticipantState) error {
		kept = append(kept, set)
		ids := make([]string, len(set))
		for i, player := range set {
			ids[i] = player.ID
		}
		want = append(want, ids)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool, len(kept))
	for i, set := range kept {
		ids := make([]string, len(set))
		for j, player := range set {
			ids[j] = player.ID
		}
		if !reflect.DeepEqual(ids, want[i]) {
			t.Fatalf("set %d changed after it was visited: %v, want %v", i, ids, want[i])
		}
		key := fmt.Sprint(ids)
		if seen[key] {
			t.Fatalf("set %v visited twice", ids)
		}
		seen[key] = true
	}
}

// TestSelectUpfloaterSet_C8OnlyForChosenSetTiming: with C8 costs and 190
// candidate sets of one profile, only the chosen set is minimised, so the
// search stays within its limit.
func TestSelectUpfloaterSet_C8OnlyForChosenSetTiming(t *testing.T) {
	residents := []*ParticipantState{{ID: "r1", TPN: 1, Score: 3, WasFloater: true}, {ID: "r2", TPN: 2, Score: 3, WasFloater: true}}
	forbidden := map[[2]string]bool{{"r1", "r2"}: true, {"r2", "r1"}: true}
	potential := make([]*ParticipantState, 20)
	for i := range potential {
		potential[i] = &ParticipantState{ID: fmt.Sprintf("u%02d", i+1), TPN: i + 3, Score: 2}
	}
	cost := func(upfloaters []*ParticipantState) PairCostFunc {
		up := make(map[string]bool, len(upfloaters))
		for _, player := range upfloaters {
			up[player.ID] = true
		}
		return func(a, b *ParticipantState) int {
			n := 0
			if up[a.ID] && b.WasFloater {
				n++
			}
			if up[b.ID] && a.WasFloater {
				n++
			}
			return n
		}
	}
	for _, p := range potential[:2] {
		forbidden[[2]string{"r1", p.ID}] = true
		forbidden[[2]string{p.ID, "r1"}] = true
	}
	budget := topDownSearchLimit
	start := time.Now()
	set, pairs, err := selectUpfloaterSet(context.Background(), residents, potential, forbidden, TopDownOptions{CostForBracket: cost}, &budget)
	if err != nil {
		t.Fatalf("selectUpfloaterSet() error: %v", err)
	}
	if len(set) == 0 || len(pairs) == 0 {
		t.Fatalf("selectUpfloaterSet() = %v, %v; want a set and its pairs", set, pairs)
	}
	if elapsed := time.Since(start); !raceEnabled && elapsed > 2*time.Second {
		t.Fatalf("selectUpfloaterSet() took %v, want at most 2s", elapsed)
	}
}
