package lexswiss

import (
	"testing"
)

func makeParticipant(id string, tpn int) *ParticipantState {
	return &ParticipantState{ID: id, TPN: tpn, Active: true}
}

func TestPairBracket_BasicFourPlayers(t *testing.T) {
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
		makeParticipant("p3", 3),
		makeParticipant("p4", 4),
	}

	// No extra criteria.
	pairs := PairBracket(participants, nil, nil)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// Lexicographic order: p1 pairs with lowest available TPN (p2),
	// then p3 pairs with p4.
	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p2" {
		t.Errorf("pair 0: expected p1 vs p2, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
	if pairs[1][0].ID != "p3" || pairs[1][1].ID != "p4" {
		t.Errorf("pair 1: expected p3 vs p4, got %s vs %s", pairs[1][0].ID, pairs[1][1].ID)
	}
}

func TestPairBracket_AvoidRepeatPairing(t *testing.T) {
	// p1 already played p2 → C1 violated. Next try: p1 vs p3.
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

	// p1 can't play p2, so p1 plays p3. Then p2 plays p4.
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

	pairs := PairBracket(participants, forbidden, nil)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// p1 can't play p2 (forbidden), so p1 plays p3. Then p2 plays p4.
	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p3" {
		t.Errorf("pair 0: expected p1 vs p3, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
}

func TestPairBracket_TwoPlayers(t *testing.T) {
	participants := []*ParticipantState{
		makeParticipant("p1", 1),
		makeParticipant("p2", 2),
	}

	pairs := PairBracket(participants, nil, nil)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p2" {
		t.Errorf("expected p1 vs p2, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
}

func TestPairBracket_SixPlayersWithConstraints(t *testing.T) {
	// p1 played p2 and p3. Next lexicographic: p1 vs p4.
	// Then p2 hasn't played p3, so p2 vs p3. Then p5 vs p6.
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Opponents: []string{"p2", "p3"}, Active: true},
		{ID: "p2", TPN: 2, Opponents: []string{"p1"}, Active: true},
		{ID: "p3", TPN: 3, Opponents: []string{"p1"}, Active: true},
		makeParticipant("p4", 4),
		makeParticipant("p5", 5),
		makeParticipant("p6", 6),
	}

	pairs := PairBracket(participants, nil, nil)
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}

	if pairs[0][0].ID != "p1" || pairs[0][1].ID != "p4" {
		t.Errorf("pair 0: expected p1 vs p4, got %s vs %s", pairs[0][0].ID, pairs[0][1].ID)
	}
	if pairs[1][0].ID != "p2" || pairs[1][1].ID != "p3" {
		t.Errorf("pair 1: expected p2 vs p3, got %s vs %s", pairs[1][0].ID, pairs[1][1].ID)
	}
	if pairs[2][0].ID != "p5" || pairs[2][1].ID != "p6" {
		t.Errorf("pair 2: expected p5 vs p6, got %s vs %s", pairs[2][0].ID, pairs[2][1].ID)
	}
}

func TestPairBracket_ImpossiblePairing(t *testing.T) {
	// p1 and p2 have both played each other — can't pair.
	participants := []*ParticipantState{
		{ID: "p1", TPN: 1, Opponents: []string{"p2"}, Active: true},
		{ID: "p2", TPN: 2, Opponents: []string{"p1"}, Active: true},
	}

	pairs := PairBracket(participants, nil, nil)
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs (impossible), got %d", len(pairs))
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

	pairs := PairBracket(participants, nil, nil)
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

	pairs := PairBracket(participants, nil, criteria)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// p1 vs p2 rejected by criteria → p1 vs p3, p2 vs p4.
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

	pairs := PairBracket(participants, nil, nil)
	// Should pair p1 vs p2, leave p3 unpaired.
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair (odd players), got %d", len(pairs))
	}
}
