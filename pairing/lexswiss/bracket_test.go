// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"
)

func makeParticipant(id string, tpn int) *ParticipantState {
	return &ParticipantState{ID: id, TPN: tpn, Active: true}
}

func TestPairBracketMinimizing_NoZeroCostTiming(t *testing.T) {
	participants := make([]*ParticipantState, 14)
	for i := range participants {
		participants[i] = makeParticipant("p"+strconv.Itoa(i+1), i+1)
	}
	budget := topDownSearchLimit
	start := time.Now()
	pairs, err := PairBracketMinimizing(context.Background(), participants, nil, nil, func(*ParticipantState, *ParticipantState) int { return 1 }, &budget)
	if elapsed := time.Since(start); !raceEnabled && elapsed > 2*time.Second {
		t.Errorf("PairBracketMinimizing() returned after %v, want at most 2s", elapsed)
	}
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v", err)
	}
	if len(pairs) != 7 {
		t.Errorf("PairBracketMinimizing() returned %d pairs, want 7", len(pairs))
	}
}

func TestPairBracketMinimizing_EquivalentToBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for n := 2; n <= 10; n += 2 {
		for trial := 0; trial < 50; trial++ {
			participants := make([]*ParticipantState, n)
			upfloaters := make(map[string]bool)
			for i := range participants {
				participants[i] = makeParticipant("p"+strconv.Itoa(i+1), i+1)
				participants[i].WasFloater = rng.Intn(2) == 0
				upfloaters[participants[i].ID] = rng.Intn(2) == 0
			}
			for i := range participants {
				j := rng.Intn(len(participants))
				participants[i], participants[j] = participants[j], participants[i]
			}
			forbidden := make(map[[2]string]bool)
			for i := 0; i < n; i++ {
				for j := i + 1; j < n; j++ {
					if rng.Intn(8) == 0 {
						forbidden[[2]string{participants[i].ID, participants[j].ID}] = true
					}
					if rng.Intn(8) == 0 {
						participants[i].Opponents = append(participants[i].Opponents, participants[j].ID)
						participants[j].Opponents = append(participants[j].Opponents, participants[i].ID)
					}
				}
			}
			cost := func(a, b *ParticipantState) int {
				result := 0
				if upfloaters[a.ID] && b.WasFloater {
					result++
				}
				if upfloaters[b.ID] && a.WasFloater {
					result++
				}
				return result
			}
			want, wantErr := bruteForceMinimumPairing(participants, forbidden, cost)
			budget := topDownSearchLimit
			got, gotErr := PairBracketMinimizing(context.Background(), participants, forbidden, nil, cost, &budget)
			if !errors.Is(gotErr, wantErr) {
				t.Fatalf("n=%d trial=%d: error = %v, want %v", n, trial, gotErr, wantErr)
			}
			if wantErr != nil {
				continue
			}
			if !samePairs(got, want) {
				t.Errorf("n=%d trial=%d: pairs = %v, want %v", n, trial, pairIDs(got), pairIDs(want))
			}
		}
	}
}

func TestPairBracketMinimizing_C8Timing(t *testing.T) {
	participants := make([]*ParticipantState, 30)
	upfloaters := make(map[string]bool)
	for i := range participants {
		participants[i] = makeParticipant("p"+strconv.Itoa(i+1), i+1)
		if i < 26 {
			participants[i].WasFloater = true
		} else {
			upfloaters[participants[i].ID] = true
		}
	}
	forbidden := make(map[[2]string]bool)
	for i := 26; i < len(participants); i++ {
		for j := i + 1; j < len(participants); j++ {
			forbidden[[2]string{participants[i].ID, participants[j].ID}] = true
		}
	}
	cost := func(a, b *ParticipantState) int {
		if upfloaters[a.ID] && b.WasFloater || upfloaters[b.ID] && a.WasFloater {
			return 1
		}
		return 0
	}
	budget := topDownSearchLimit
	start := time.Now()
	pairs, err := PairBracketMinimizing(context.Background(), participants, forbidden, nil, cost, &budget)
	if elapsed := time.Since(start); !raceEnabled && elapsed > time.Second {
		t.Errorf("PairBracketMinimizing() returned after %v, want at most 1s", elapsed)
	}
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v", err)
	}
	if len(pairs) != len(participants)/2 {
		t.Errorf("PairBracketMinimizing() returned %d pairs, want %d", len(pairs), len(participants)/2)
	}
}

func bruteForceMinimumPairing(participants []*ParticipantState, forbidden map[[2]string]bool, cost PairCostFunc) ([][2]*ParticipantState, error) {
	participants = append([]*ParticipantState(nil), participants...)
	sort.Slice(participants, func(i, j int) bool { return participants[i].TPN < participants[j].TPN })
	minimum := 0
	var result [][2]*ParticipantState
	found := false
	_, err := enumerateBracketPairings(context.Background(), participants, forbidden, nil, func(pairs [][2]*ParticipantState) bool {
		pairingCost := 0
		for _, pair := range pairs {
			pairingCost += cost(pair[0], pair[1])
		}
		if !found || pairingCost < minimum {
			minimum = pairingCost
			result = append([][2]*ParticipantState(nil), pairs...)
			found = true
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNoCompletePairing
	}
	return result, nil
}

func samePairs(a, b [][2]*ParticipantState) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i][0].ID != b[i][0].ID || a[i][1].ID != b[i][1].ID {
			return false
		}
	}
	return true
}

func pairIDs(pairs [][2]*ParticipantState) [][2]string {
	ids := make([][2]string, len(pairs))
	for i, pair := range pairs {
		ids[i] = [2]string{pair[0].ID, pair[1].ID}
	}
	return ids
}

func TestPairBracket_BasicFourPlayers(t *testing.T) {
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
		makeParticipant("p3", 3),
		makeParticipant("p4", 4),
	}

	// No extra criteria.
	pairs, err := PairBracket(context.Background(), participants, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// The first identifier is (1,2,3,4): p1-p3, p2-p4.
	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p3" {
		t.Errorf("pair 0: expected p1 vs p3, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
	if pairs[1][0].ID != "p2" || pairs[1][1].ID != "p4" {
		t.Errorf("pair 1: expected p2 vs p4, got %s vs %s", pairs[1][0].ID, pairs[1][1].ID)
	}
}

func TestEnumerateBracketPairings_IdentifierOrder(t *testing.T) {
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
		makeParticipant("p3", 3),
		makeParticipant("p4", 4),
		makeParticipant("p5", 5),
		makeParticipant("p6", 6),
	}
	var identifiers [][]int
	_, err := enumerateBracketPairings(context.Background(), participants, nil, nil, func(pairs [][2]*ParticipantState) bool {
		identifier := make([]int, 0, len(pairs)*2)
		for _, pair := range pairs {
			identifier = append(identifier, pair[0].TPN)
		}
		for _, pair := range pairs {
			identifier = append(identifier, pair[1].TPN)
		}
		identifiers = append(identifiers, identifier)
		return len(identifiers) < 4
	})
	if err != nil {
		t.Fatalf("enumerateBracketPairings() error: %v", err)
	}
	want := [][]int{
		{1, 2, 3, 4, 5, 6},
		{1, 2, 3, 4, 6, 5},
		{1, 2, 3, 5, 4, 6},
		{1, 2, 3, 5, 6, 4},
	}
	if len(identifiers) != len(want) {
		t.Fatalf("got %d identifiers, want %d", len(identifiers), len(want))
	}
	for i := range want {
		for j := range want[i] {
			if identifiers[i][j] != want[i][j] {
				t.Errorf("identifier %d = %v, want %v", i, identifiers[i], want[i])
				break
			}
		}
	}
}

func TestPairBracket_IdentifierOrderAfterForbiddenPair(t *testing.T) {
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
		makeParticipant("p3", 3),
		makeParticipant("p4", 4),
		makeParticipant("p5", 5),
		makeParticipant("p6", 6),
	}
	pairs, err := PairBracket(context.Background(), participants, map[[2]string]bool{{"p1", "p4"}: true}, nil)
	if err != nil {
		t.Fatalf("PairBracket() error: %v", err)
	}
	want := [][2]string{{"p1", "p5"}, {"p2", "p4"}, {"p3", "p6"}}
	for i, pair := range pairs {
		if got := [2]string{pair[0].ID, pair[1].ID}; got != want[i] {
			t.Errorf("pair %d = %v, want %v", i, got, want[i])
		}
	}
}

func TestPairBracket_AvoidRepeatPairing(t *testing.T) {
	// p1 already played p2, which is not in the first identifier pairing.
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Opponents: []string{"p2"}, Active: true},
		{ID: "p2", TPN: 2, Opponents: []string{"p1"}, Active: true},
		{ID: "p3", TPN: 3, Active: true},
		{ID: "p4", TPN: 4, Active: true},
	}

	pairs, err := PairBracket(context.Background(), participants, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// The first identifier pairing remains p1-p3, p2-p4.
	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p3" {
		t.Errorf("pair 0: expected p1 vs p3, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
	if pairs[1][0].ID != "p2" || pairs[1][1].ID != "p4" {
		t.Errorf("pair 1: expected p2 vs p4, got %s vs %s", pairs[1][0].ID, pairs[1][1].ID)
	}
}

func TestPairBracket_ForbiddenPair(t *testing.T) {
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
		makeParticipant("p3", 3),
		makeParticipant("p4", 4),
	}

	// p1 vs p2 is forbidden.
	forbidden := map[[2]string]bool{
		{"p1", "p2"}: true,
	}

	pairs, err := PairBracket(context.Background(), participants, forbidden, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// The first identifier pairing remains p1-p3, p2-p4.
	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p3" {
		t.Errorf("pair 0: expected p1 vs p3, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
}

func TestPairBracket_TwoPlayers(t *testing.T) {
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
	}

	pairs, err := PairBracket(context.Background(), participants, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p2" {
		t.Errorf("expected p1 vs p2, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
}

func TestPairBracket_SixPlayersWithConstraints(t *testing.T) {
	// p1 played p2 and p3. The first identifier pairs p1-p4, p2-p5, p3-p6.
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Opponents: []string{"p2", "p3"}, Active: true},
		{ID: "p2", TPN: 2, Opponents: []string{"p1"}, Active: true},
		{ID: "p3", TPN: 3, Opponents: []string{"p1"}, Active: true},
		makeParticipant("p4", 4),
		makeParticipant("p5", 5),
		makeParticipant("p6", 6),
	}

	pairs, err := PairBracket(context.Background(), participants, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}

	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p4" {
		t.Errorf("pair 0: expected p1 vs p4, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
	if pairs[1][0].ID != "p2" || pairs[1][1].ID != "p5" {
		t.Errorf("pair 1: expected p2 vs p5, got %s vs %s", pairs[1][0].ID, pairs[1][1].ID)
	}
	if pairs[2][0].ID != "p3" || pairs[2][1].ID != "p6" {
		t.Errorf("pair 2: expected p3 vs p6, got %s vs %s", pairs[2][0].ID, pairs[2][1].ID)
	}
}

func TestPairBracket_ConstrainedFortyParticipants(t *testing.T) {
	participants := make([]*ParticipantState, 40)
	p1Opponents := make([]string, 0, 38)
	for i := range participants {
		participants[i] = makeParticipant("p"+strconv.Itoa(i+1), i+1)
		if i > 1 {
			p1Opponents = append(p1Opponents, participants[i].ID)
		}
	}
	participants[0].Opponents = p1Opponents // p1's only legal opponent is p2.

	started := time.Now()
	pairs, err := PairBracket(context.Background(), participants, nil, nil)
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("PairBracket() error: %v", err)
	}
	if elapsed >= 2*time.Second {
		t.Fatalf("PairBracket() took %v, want less than 2s", elapsed)
	}
	t.Logf("PairBracket() completed in %v", elapsed)
	if got := [2]string{pairs[0][0].ID, pairs[0][1].ID}; got != [2]string{"p1", "p2"} {
		t.Errorf("first pair = %v, want [p1 p2]", got)
	}
}

func TestPairBracket_ImpossiblePairing(t *testing.T) {
	// p1 and p2 have both played each other — can't pair.
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Opponents: []string{"p2"}, Active: true},
		{ID: "p2", TPN: 2, Opponents: []string{"p1"}, Active: true},
	}

	pairs, err := PairBracket(context.Background(), participants, nil, nil)
	if !errors.Is(err, ErrNoCompletePairing) {
		t.Fatalf("error = %v, want ErrNoCompletePairing", err)
	}
	if pairs != nil {
		t.Errorf("pairs = %v, want nil", pairs)
	}
}

func TestPairBracketUsesFixedNonContiguousTPNs(t *testing.T) {
	states := SortParticipants([]ParticipantState{
		{ID: "p2", TPN: 2, Score: 0, Active: true},
		{ID: "p3", TPN: 3, Score: 0, Opponents: []string{"p10"}, Active: true},
		{ID: "p6", TPN: 6, Score: 2, Active: true},
		{ID: "p10", TPN: 10, Score: 1, Opponents: []string{"p3"}, Active: true},
	})
	participants := make([]*ParticipantState, len(states))
	for i := range states {
		participants[i] = &states[i]
	}

	pairs, err := PairBracket(context.Background(), participants, nil, nil)
	if err != nil {
		t.Fatalf("PairBracket() error: %v", err)
	}
	if len(pairs) != 2 || pairs[0][0].ID != "p2" || pairs[0][1].ID != "p10" || pairs[1][0].ID != "p3" || pairs[1][1].ID != "p6" {
		t.Fatalf("pairs = %v; want 2-10, 3-6 under Article 3.6", pairs)
	}
}

func TestPairBracket_Backtracking(t *testing.T) {
	// 4 players: p1 played p2, p3 played p4.
	// First try: p1 vs p3 → leaves p2 vs p4 (OK).
	// Without backtracking a naive greedy would try p1 vs p3,
	// but both p1-p2 and p3-p4 are blocked so we need:
	// p1 vs p3 (OK), p2 vs p4 (OK) — or p1 vs p4, p2 vs p3.
	// Lexicographic: p1 vs p3 first (lower TPN), then p2 vs p4.
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Opponents: []string{"p2"}, Active: true},
		{ID: "p2", TPN: 2, Opponents: []string{"p1"}, Active: true},
		{ID: "p3", TPN: 3, Opponents: []string{"p4"}, Active: true},
		{ID: "p4", TPN: 4, Opponents: []string{"p3"}, Active: true},
	}

	pairs, err := PairBracket(context.Background(), participants, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p3" {
		t.Errorf("pair 0: expected p1 vs p3, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
	if pairs[1][0].ID != "p2" || pairs[1][1].ID != "p4" {
		t.Errorf("pair 1: expected p2 vs p4, got %s vs %s", pairs[1][0].ID, pairs[1][1].ID)
	}
}

func TestPairBracket_WithCriteriaFunc(t *testing.T) {
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
		makeParticipant("p3", 3),
		makeParticipant("p4", 4),
	}

	// Criteria: reject p1 vs p2 pairing.
	criteria := func(a, b *ParticipantState) bool {
		if (a.ID == "p1" && b.ID == "p2") || (a.ID == "p2" && b.ID == "p1") {
			return false
		}
		return true
	}

	pairs, err := PairBracket(context.Background(), participants, nil, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// The rejected pair is not part of the first identifier pairing.
	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p3" {
		t.Errorf("pair 0: expected p1 vs p3, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
}

func TestPairBracket_OddPlayers(t *testing.T) {
	// Odd number — caller should have removed bye player.
	// If called with odd, pair as many as possible (one unpaired).
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
		makeParticipant("p3", 3),
	}

	pairs, err := PairBracket(context.Background(), participants, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Odd brackets retain their existing handling: p1 pairs with p2, leaving p3 unpaired.
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair (odd players), got %d", len(pairs))
	}
}

func TestPairBracket_BlockedBottomBacktracks(t *testing.T) {
	participants := make([]*ParticipantState, 40)
	for i := range participants {
		participants[i] = makeParticipant("p"+strconv.Itoa(i+1), i+1)
	}
	// p20's only legal bottom is p21, so p1 must not take p21.
	for j := 21; j < 40; j++ {
		participants[19].Opponents = append(participants[19].Opponents, participants[j].ID)
		participants[j].Opponents = append(participants[j].Opponents, participants[19].ID)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pairs, err := PairBracket(ctx, participants, nil, nil)
	if err != nil {
		t.Fatalf("PairBracket() error: %v", err)
	}
	if got := [2]string{pairs[0][0].ID, pairs[0][1].ID}; got != [2]string{"p1", "p22"} {
		t.Errorf("first pair = %v, want [p1 p22]", got)
	}
	if got := [2]string{pairs[19][0].ID, pairs[19][1].ID}; got != [2]string{"p20", "p21"} {
		t.Errorf("last pair = %v, want [p20 p21]", got)
	}
}
