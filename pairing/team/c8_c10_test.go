// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"strconv"
	"testing"

	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestPairBracket_C8_LaterIdentifierFulfilsAll(t *testing.T) {
	// A bracket whose first legal identifier leaves two teams with the same
	// Type A preference paired, while a later identifier fulfils all → the later one.
	// t1, t3 want White; t2, t4 want Black.
	// Identifier 1: (t1-t3, t2-t4) -> C8 = 2.
	// Identifier 2: (t1-t4, t2-t3) -> C8 = 0.
	participants := []*lexswiss.ParticipantState{
		{ID: "t1", TPN: 1, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t2", TPN: 2, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite, lexswiss.ColorWhite}},
		{ID: "t3", TPN: 3, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t4", TPN: 4, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite, lexswiss.ColorWhite}},
	}
	cost, err := BracketPairCost(len(participants), nil, ColorPrefTypeA, false, false)
	if err != nil {
		t.Fatalf("BracketPairCost() error: %v", err)
	}
	pairs, err := lexswiss.PairBracketMinimizing(context.Background(), participants, nil, nil, cost, nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("got %d pairs, want 2", len(pairs))
	}
	if pairs[0][0].ID != "t1" || pairs[0][1].ID != "t4" || pairs[1][0].ID != "t2" || pairs[1][1].ID != "t3" {
		t.Errorf("got pairs %v-%v, %v-%v; want t1-t4, t2-t3", pairs[0][0].ID, pairs[0][1].ID, pairs[1][0].ID, pairs[1][1].ID)
	}
}

func TestPairBracket_C8_CannotReachZero_FirstIdentifierAtMinimum(t *testing.T) {
	// A case where C8 cannot reach 0 → first identifier at the minimum.
	// All 4 teams want White. Every pairing has C8 = 2.
	// First legal identifier is (t1-t3, t2-t4).
	participants := []*lexswiss.ParticipantState{
		{ID: "t1", TPN: 1, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t2", TPN: 2, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t3", TPN: 3, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t4", TPN: 4, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
	}
	cost, err := BracketPairCost(len(participants), nil, ColorPrefTypeA, false, false)
	if err != nil {
		t.Fatalf("BracketPairCost() error: %v", err)
	}
	pairs, err := lexswiss.PairBracketMinimizing(context.Background(), participants, nil, nil, cost, nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("got %d pairs, want 2", len(pairs))
	}
	if pairs[0][0].ID != "t1" || pairs[0][1].ID != "t3" || pairs[1][0].ID != "t2" || pairs[1][1].ID != "t4" {
		t.Errorf("got pairs %v-%v, %v-%v; want first identifier t1-t3, t2-t4", pairs[0][0].ID, pairs[0][1].ID, pairs[1][0].ID, pairs[1][1].ID)
	}
}

func TestPairBracket_C9_OnlyWithTypeB(t *testing.T) {
	// Under Type B, strong preferences are distinguished from mild ones by C9.
	// Under Type A, C9 does not apply (Art. 2.3.6: Type B only).
	// t1, t2: strong White (CD = -2).
	// t3, t4: mild White (CD = -1, last match Black).
	// In Type B, all 4 want White, so C8 = 2 for all pairings.
	// (t1-t2, t3-t4) pairs strong vs strong (C9 = 1).
	// (t1-t3, t2-t4) pairs strong vs mild (C9 = 0).
	// Type B must choose (t1-t3, t2-t4).
	participants := func() []*lexswiss.ParticipantState {
		return []*lexswiss.ParticipantState{
			{ID: "t1", TPN: 1, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack}},
			{ID: "t2", TPN: 2, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack}},
			{ID: "t3", TPN: 3, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack}},
			{ID: "t4", TPN: 4, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack}},
		}
	}

	costB, err := BracketPairCost(4, nil, ColorPrefTypeB, false, false)
	if err != nil {
		t.Fatalf("BracketPairCost(TypeB) error: %v", err)
	}
	pairsB, err := lexswiss.PairBracketMinimizing(context.Background(), participants(), nil, nil, costB, nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing(TypeB) error: %v", err)
	}
	if pairsB[0][0].ID != "t1" || pairsB[0][1].ID != "t3" || pairsB[1][0].ID != "t2" || pairsB[1][1].ID != "t4" {
		t.Errorf("Type B got %v-%v, %v-%v; want t1-t3, t2-t4 to minimise C9", pairsB[0][0].ID, pairsB[0][1].ID, pairsB[1][0].ID, pairsB[1][1].ID)
	}

	// Verify C9 is 0 under Type A for the pair of strong preferences.
	p := participants()
	if got := C9Cost(p[0], p[1], ColorPrefTypeA, false); got != 0 {
		t.Errorf("C9Cost under Type A = %d, want 0", got)
	}
	if got := C9Cost(p[0], p[1], ColorPrefTypeB, false); got != 1 {
		t.Errorf("C9Cost under Type B = %d, want 1", got)
	}
}

func TestPairBracket_C8_Dominates_C9(t *testing.T) {
	// C8 strictly dominates C9: a pairing with smaller C8 must beat one with smaller C9.
	// 6 teams:
	// t1, t2: strong White
	// t3, t4: mild Black
	// t5, t6: mild White
	// By forbidding cross-pairs, only two pairings are possible:
	// Pairing A: (t1-t2, t3-t5, t4-t6) -> (1,2) strong vs strong (C8=1, C9=1); others opposite (C8=0, C9=0).
	//            Total: C8 = 1, C9 = 1.
	// Pairing B: (t1-t5, t2-t6, t3-t4) -> (1,5) strong vs mild (C8=1, C9=0); (2,6) strong vs mild (C8=1, C9=0);
	//            (3,4) mild Black vs mild Black (C8=1, C9=0).
	//            Total: C8 = 3, C9 = 0.
	// C8 dominance requires Pairing A (C8=1) over Pairing B (C8=3, C9=0).
	participants := []*lexswiss.ParticipantState{
		{ID: "t1", TPN: 1, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t2", TPN: 2, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t3", TPN: 3, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite}},
		{ID: "t4", TPN: 4, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite}},
		{ID: "t5", TPN: 5, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack}},
		{ID: "t6", TPN: 6, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack}},
	}
	forbidden := make(map[[2]string]bool)
	allowed := map[[2]string]bool{
		{"t1", "t2"}: true, {"t2", "t1"}: true,
		{"t3", "t5"}: true, {"t5", "t3"}: true,
		{"t4", "t6"}: true, {"t6", "t4"}: true,
		{"t1", "t5"}: true, {"t5", "t1"}: true,
		{"t2", "t6"}: true, {"t6", "t2"}: true,
		{"t3", "t4"}: true, {"t4", "t3"}: true,
	}
	for i := range participants {
		for j := i + 1; j < len(participants); j++ {
			p1, p2 := participants[i].ID, participants[j].ID
			if !allowed[[2]string{p1, p2}] {
				forbidden[[2]string{p1, p2}] = true
				forbidden[[2]string{p2, p1}] = true
			}
		}
	}
	cost, err := BracketPairCost(len(participants), nil, ColorPrefTypeB, false, false)
	if err != nil {
		t.Fatalf("BracketPairCost() error: %v", err)
	}
	pairs, err := lexswiss.PairBracketMinimizing(context.Background(), participants, forbidden, nil, cost, nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v", err)
	}
	// Pairing A must be selected: {t1, t2}, {t3, t5}, {t4, t6}
	if pairs[0][0].ID != "t1" || pairs[0][1].ID != "t2" {
		t.Errorf("got pair 0 %v-%v, want t1-t2 (Pairing A favoured by C8 over Pairing B)", pairs[0][0].ID, pairs[0][1].ID)
	}
}

func TestPairBracket_C9_Dominates_C10(t *testing.T) {
	// C9 strictly dominates C10: when C8 is equal, smaller C9 beats smaller C10.
	// 4 teams:
	// Residents: t1, t2 (WasFloater = true)
	// Upfloaters: t3, t4 (WasFloater = false)
	// t1, t2: strong White
	// t3, t4: mild White
	// Pairing A: (t1-t3, t2-t4) -> C8 = 2, C9 = 0, C10 = 2 (each resident is a floater paired with an upfloater).
	// Pairing B: (t1-t2, t3-t4) -> C8 = 2, C9 = 1, C10 = 0 (no resident paired with an upfloater).
	// Because C9 dominates C10, Pairing A (C9=0, C10=2) must beat Pairing B (C9=1, C10=0).
	participants := []*lexswiss.ParticipantState{
		{ID: "t1", TPN: 1, WasFloater: true, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t2", TPN: 2, WasFloater: true, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t3", TPN: 3, WasFloater: false, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack}},
		{ID: "t4", TPN: 4, WasFloater: false, ColorHistory: []lexswiss.Color{lexswiss.ColorBlack}},
	}
	upfloaters := participants[2:]
	cost, err := BracketPairCost(len(participants), upfloaters, ColorPrefTypeB, false, false)
	if err != nil {
		t.Fatalf("BracketPairCost() error: %v", err)
	}
	pairs, err := lexswiss.PairBracketMinimizing(context.Background(), participants, nil, nil, cost, nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v", err)
	}
	// Must choose Pairing A: t1-t3, t2-t4
	if pairs[0][0].ID != "t1" || pairs[0][1].ID != "t3" || pairs[1][0].ID != "t2" || pairs[1][1].ID != "t4" {
		t.Errorf("got %v-%v, %v-%v; want t1-t3, t2-t4 (C9=0 beats C10=0)", pairs[0][0].ID, pairs[0][1].ID, pairs[1][0].ID, pairs[1][1].ID)
	}
}

func TestPairBracket_C10_OffInLastTwoRounds(t *testing.T) {
	// In the last two rounds, C10 is not applied (Art. 2.3.7).
	// Residents: t1 (WasFloater = true), t2 (WasFloater = false).
	// Upfloaters: t3, t4 (WasFloater = false).
	// No colour preferences.
	// Pairing A: (t1-t2, t3-t4) -> C10 = 0.
	// Pairing B: (t1-t3, t2-t4) -> C10 = 1 (t3 upfloater opponent t1 was floater).
	// In non-last two rounds: C10 active → chooses (t1-t2, t3-t4) to minimise C10.
	// In last two rounds: C10 off → all pairings cost 0 → chooses first identifier (t1-t3, t2-t4).
	participants := []*lexswiss.ParticipantState{
		{ID: "t1", TPN: 1, WasFloater: true},
		{ID: "t2", TPN: 2, WasFloater: false},
		{ID: "t3", TPN: 3, WasFloater: false},
		{ID: "t4", TPN: 4, WasFloater: false},
	}
	upfloaters := participants[2:]

	// Non-last two rounds: C10 active
	costActive, err := BracketPairCost(len(participants), upfloaters, ColorPrefTypeNone, false, false)
	if err != nil {
		t.Fatalf("BracketPairCost(active) error: %v", err)
	}
	pairsActive, err := lexswiss.PairBracketMinimizing(context.Background(), participants, nil, nil, costActive, nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing(active) error: %v", err)
	}
	if pairsActive[0][0].ID != "t1" || pairsActive[0][1].ID != "t2" {
		t.Errorf("C10 active: got pair 0 %v-%v, want t1-t2 to minimise C10", pairsActive[0][0].ID, pairsActive[0][1].ID)
	}

	// Last two rounds: C10 off
	costOff, err := BracketPairCost(len(participants), upfloaters, ColorPrefTypeNone, true, false)
	if err != nil {
		t.Fatalf("BracketPairCost(off) error: %v", err)
	}
	pairsOff, err := lexswiss.PairBracketMinimizing(context.Background(), participants, nil, nil, costOff, nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing(off) error: %v", err)
	}
	if pairsOff[0][0].ID != "t1" || pairsOff[0][1].ID != "t3" {
		t.Errorf("C10 off: got pair 0 %v-%v, want t1-t3 (first identifier when C10 is inactive)", pairsOff[0][0].ID, pairsOff[0][1].ID)
	}
}

func TestPairBracket_OldFiltersImpossible_C1Allows(t *testing.T) {
	// A bracket that old filters made impossible but C1 allows → paired (no error).
	// 4 teams, all wanting White. Under old criteria filters, every pair was rejected.
	// With exact minimisation, all pairings are legal and evaluated at cost.
	participants := []*lexswiss.ParticipantState{
		{ID: "t1", TPN: 1, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t2", TPN: 2, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t3", TPN: 3, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
		{ID: "t4", TPN: 4, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}},
	}
	cost, err := BracketPairCost(len(participants), nil, ColorPrefTypeA, false, false)
	if err != nil {
		t.Fatalf("BracketPairCost() error: %v", err)
	}
	pairs, err := lexswiss.PairBracketMinimizing(context.Background(), participants, nil, nil, cost, nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v, want successful pairing", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("got %d pairs, want 2", len(pairs))
	}
}

func TestPairBracket_BruteForceEquivalence(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for n := 2; n <= 8; n += 2 {
		for trial := 0; trial < 50; trial++ {
			participants := make([]*lexswiss.ParticipantState, n)
			upfloaters := make([]*lexswiss.ParticipantState, 0)
			for i := range participants {
				participants[i] = &lexswiss.ParticipantState{
					ID:         "t" + strconv.Itoa(i+1),
					TPN:        i + 1,
					WasFloater: rng.Intn(2) == 0,
				}
				// Generate random color history of length 0 to 4
				hLen := rng.Intn(5)
				for h := 0; h < hLen; h++ {
					if rng.Intn(2) == 0 {
						participants[i].ColorHistory = append(participants[i].ColorHistory, lexswiss.ColorWhite)
					} else {
						participants[i].ColorHistory = append(participants[i].ColorHistory, lexswiss.ColorBlack)
					}
				}
				if rng.Intn(2) == 0 {
					upfloaters = append(upfloaters, participants[i])
				}
			}
			forbidden := make(map[[2]string]bool)
			for i := 0; i < n; i++ {
				for j := i + 1; j < n; j++ {
					if rng.Intn(10) == 0 {
						forbidden[[2]string{participants[i].ID, participants[j].ID}] = true
						forbidden[[2]string{participants[j].ID, participants[i].ID}] = true
					}
					if rng.Intn(10) == 0 {
						participants[i].Opponents = append(participants[i].Opponents, participants[j].ID)
						participants[j].Opponents = append(participants[j].Opponents, participants[i].ID)
					}
				}
			}

			prefType := ColorPrefTypeA
			if rng.Intn(2) == 0 {
				prefType = ColorPrefTypeB
			}
			isLastTwoRounds := rng.Intn(3) == 0
			isLastRound := isLastTwoRounds && rng.Intn(2) == 0

			cost, err := BracketPairCost(n, upfloaters, prefType, isLastTwoRounds, isLastRound)
			if err != nil {
				t.Fatalf("BracketPairCost() error: %v", err)
			}

			want, wantErr := bruteForceMinimum(participants, forbidden, cost)
			got, gotErr := lexswiss.PairBracketMinimizing(context.Background(), participants, forbidden, nil, cost, nil)

			if !errors.Is(gotErr, wantErr) {
				t.Fatalf("n=%d trial=%d: gotErr=%v, wantErr=%v", n, trial, gotErr, wantErr)
			}
			if wantErr != nil {
				continue
			}
			if len(got) != len(want) {
				t.Fatalf("n=%d trial=%d: got %d pairs, want %d", n, trial, len(got), len(want))
			}
			for i := range got {
				if got[i][0].ID != want[i][0].ID || got[i][1].ID != want[i][1].ID {
					t.Fatalf("n=%d trial=%d: pair %d = %v-%v, want %v-%v", n, trial, i, got[i][0].ID, got[i][1].ID, want[i][0].ID, want[i][1].ID)
				}
			}
		}
	}
}

type candidatePairing struct {
	pairs      [][2]*lexswiss.ParticipantState
	identifier []int
	cost       int
}

func bruteForceMinimum(participants []*lexswiss.ParticipantState, forbidden map[[2]string]bool, cost lexswiss.PairCostFunc) ([][2]*lexswiss.ParticipantState, error) {
	sorted := append([]*lexswiss.ParticipantState(nil), participants...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].TPN < sorted[j].TPN })

	oppMap := make(map[[2]string]bool)
	for _, p := range sorted {
		for _, opp := range p.Opponents {
			oppMap[[2]string{p.ID, opp}] = true
		}
	}

	canPair := func(a, b *lexswiss.ParticipantState) bool {
		if forbidden != nil && (forbidden[[2]string{a.ID, b.ID}] || forbidden[[2]string{b.ID, a.ID}]) {
			return false
		}
		if oppMap[[2]string{a.ID, b.ID}] {
			return false
		}
		return true
	}

	var candidates []candidatePairing
	used := make([]bool, len(sorted))

	var build func([][2]*lexswiss.ParticipantState)
	build = func(current [][2]*lexswiss.ParticipantState) {
		if len(current) == len(sorted)/2 {
			// Compute Article 3.6 identifier
			id := make([]int, 0, len(sorted))
			totalCost := 0
			for _, pair := range current {
				id = append(id, pair[0].TPN)
			}
			for _, pair := range current {
				id = append(id, pair[1].TPN)
				totalCost += cost(pair[0], pair[1])
			}
			candidates = append(candidates, candidatePairing{
				pairs:      append([][2]*lexswiss.ParticipantState(nil), current...),
				identifier: id,
				cost:       totalCost,
			})
			return
		}

		// Pick the first unused participant as top member
		firstUnused := -1
		for i := range used {
			if !used[i] {
				firstUnused = i
				break
			}
		}
		if firstUnused == -1 {
			return
		}

		used[firstUnused] = true
		top := sorted[firstUnused]
		for j := firstUnused + 1; j < len(sorted); j++ {
			if !used[j] && canPair(top, sorted[j]) {
				used[j] = true
				build(append(current, [2]*lexswiss.ParticipantState{top, sorted[j]}))
				used[j] = false
			}
		}
		used[firstUnused] = false
	}

	build(nil)

	if len(candidates) == 0 {
		return nil, lexswiss.ErrNoCompletePairing
	}

	// Sort candidates by cost ascending, then by identifier lexicographically
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].cost != candidates[j].cost {
			return candidates[i].cost < candidates[j].cost
		}
		for k := range candidates[i].identifier {
			if candidates[i].identifier[k] != candidates[j].identifier[k] {
				return candidates[i].identifier[k] < candidates[j].identifier[k]
			}
		}
		return false
	})

	return candidates[0].pairs, nil
}

func TestPairBracket_C8_MinimumBelowFirstIdentifier(t *testing.T) {
	// t1, t2, t4, t5 want White; t3, t6 want Black. No pairing fulfils all
	// (four teams want White, two Black), so a filter with an unfiltered
	// retry returns the first identifier 1 2 3 4 5 6 (t1-t4, t2-t5, t3-t6),
	// which leaves three teams unfulfilled. C.04.6 2.3.5 and 3.6.4 ask for
	// the first identifier at the minimum, C8 = 1: 1 2 3 4 6 5
	// (t1-t4, t2-t6, t3-t5). Identifiers follow Article 3.6.2.
	wantsWhite := []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorBlack, lexswiss.ColorBlack}
	wantsBlack := []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorWhite, lexswiss.ColorWhite}
	participants := []*lexswiss.ParticipantState{
		{ID: "t1", TPN: 1, ColorHistory: wantsWhite},
		{ID: "t2", TPN: 2, ColorHistory: wantsWhite},
		{ID: "t3", TPN: 3, ColorHistory: wantsBlack},
		{ID: "t4", TPN: 4, ColorHistory: wantsWhite},
		{ID: "t5", TPN: 5, ColorHistory: wantsWhite},
		{ID: "t6", TPN: 6, ColorHistory: wantsBlack},
	}
	cost, err := BracketPairCost(len(participants), nil, ColorPrefTypeA, false, false)
	if err != nil {
		t.Fatalf("BracketPairCost() error: %v", err)
	}
	pairs, err := lexswiss.PairBracketMinimizing(context.Background(), participants, nil, nil, cost, nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v", err)
	}
	want := [][2]string{{"t1", "t4"}, {"t2", "t6"}, {"t3", "t5"}}
	if len(pairs) != len(want) {
		t.Fatalf("got %d pairs, want %d", len(pairs), len(want))
	}
	for i, pair := range pairs {
		if pair[0].ID != want[i][0] || pair[1].ID != want[i][1] {
			t.Errorf("pair %d = %s-%s, want %s-%s", i, pair[0].ID, pair[1].ID, want[i][0], want[i][1])
		}
	}
}
