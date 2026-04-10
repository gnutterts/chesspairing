// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import "testing"

// TestInvariant_PairBracket_Completeness verifies that PairBracket produces
// a complete pairing when there are no constraints: 6 unconstrained
// participants must yield 3 pairs covering all 6 participants exactly once.
func TestInvariant_PairBracket_Completeness(t *testing.T) {
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
		makeParticipant("p3", 3),
		makeParticipant("p4", 4),
		makeParticipant("p5", 5),
		makeParticipant("p6", 6),
	}

	pairs := PairBracket(participants, nil, nil)

	// Invariant: exactly 3 pairs for 6 participants.
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs for 6 participants, got %d", len(pairs))
	}

	// Invariant: every participant appears exactly once across all pairs.
	seen := make(map[string]int)
	for _, pair := range pairs {
		seen[pair[0].ID]++
		seen[pair[1].ID]++
	}

	if len(seen) != 6 {
		t.Fatalf("expected 6 distinct participants in pairs, got %d", len(seen))
	}

	for id, count := range seen {
		if count != 1 {
			t.Errorf("participant %s appears %d times, want exactly 1", id, count)
		}
	}
}

// TestInvariant_PairBracket_NoRematches verifies that PairBracket respects
// criterion C1: two participants who have already played each other are
// never paired again.
func TestInvariant_PairBracket_NoRematches(t *testing.T) {
	// p1 and p2 have already played each other.
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Opponents: []string{"p2"}, Active: true},
		{ID: "p2", TPN: 2, Opponents: []string{"p1"}, Active: true},
		{ID: "p3", TPN: 3, Active: true},
		{ID: "p4", TPN: 4, Active: true},
	}

	pairs := PairBracket(participants, nil, nil)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// Invariant: p1 and p2 must NOT be in the same pair.
	for _, pair := range pairs {
		ids := [2]string{pair[0].ID, pair[1].ID}
		if (ids[0] == "p1" && ids[1] == "p2") || (ids[0] == "p2" && ids[1] == "p1") {
			t.Errorf("p1 and p2 were paired despite having already played each other")
		}
	}
}

// TestInvariant_PairBracket_DoubleForfeitExcluded verifies the caller
// contract for double forfeits: a double-forfeited game is NOT recorded
// in the Opponents history (the game "never happened"), so PairBracket
// treats the two participants as eligible to be paired.
//
// This is a contract test: the test sets up the state the way a correct
// caller would (no opponent entry for the double-forfeited game) and
// verifies that PairBracket allows the pairing.
func TestInvariant_PairBracket_DoubleForfeitExcluded(t *testing.T) {
	// Scenario: p1 and p3 had a double forfeit in round 1.
	// Per the forfeit semantics, double forfeits are excluded from opponent
	// history — so Opponents slices do NOT contain each other.
	// PairBracket should therefore allow p1 vs p3.
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Active: true}, // no opponent history for the double forfeit
		{ID: "p2", TPN: 2, Active: true},
		{ID: "p3", TPN: 3, Active: true},
		{ID: "p4", TPN: 4, Active: true},
	}

	pairs := PairBracket(participants, nil, nil)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// With no constraints at all, lexicographic order gives p1-p2, p3-p4.
	// The key invariant: p1 and p3 ARE allowed to pair (no C1 block).
	// We verify this indirectly: the default lexicographic pairing (p1-p2,
	// p3-p4) succeeds, which means p1 and p3 were not blocked.
	//
	// For a stronger check, forbid p1-p2 so the algorithm must pair p1-p3.
	forbidden := map[[2]string]bool{
		{"p1", "p2"}: true,
	}

	pairs = PairBracket(participants, forbidden, nil)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs with forbidden p1-p2, got %d", len(pairs))
	}

	// Now p1 must pair with p3 (next lexicographic choice after p2 is blocked).
	// This confirms p1-p3 is allowed despite the "double forfeit" having occurred.
	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p3" {
		t.Errorf("expected p1 vs p3 (double forfeit not blocking), got %s vs %s",
			pairs[0][0].ID, pairs[0][1].ID)
	}
	if pairs[1][0].ID != "p2" || pairs[1][1].ID != "p4" {
		t.Errorf("expected p2 vs p4 as remaining pair, got %s vs %s",
			pairs[1][0].ID, pairs[1][1].ID)
	}
}
