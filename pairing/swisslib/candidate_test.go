package swisslib

import "testing"

func TestCandidateScore_Compare_AllZero(t *testing.T) {
	a := CandidateScore{}
	b := CandidateScore{}
	if a.Compare(&b) != 0 {
		t.Error("two zero scores should be equal")
	}
}

func TestCandidateScore_Compare_FloaterScores(t *testing.T) {
	// C7: floater scores sorted descending, lexicographic comparison.
	// Fewer/lower floater scores = better.
	a := CandidateScore{FloaterScores: []float64{3.0, 2.0}}
	b := CandidateScore{FloaterScores: []float64{3.0, 1.0}}
	// a has higher second floater score → a is worse → Compare returns +1
	if a.Compare(&b) != 1 {
		t.Errorf("expected a > b (worse), got %d", a.Compare(&b))
	}
	if b.Compare(&a) != -1 {
		t.Errorf("expected b < a (better), got %d", b.Compare(&a))
	}
}

func TestCandidateScore_Compare_FloaterScoresLength(t *testing.T) {
	// More floaters = worse (even if individual scores are lower).
	a := CandidateScore{FloaterScores: []float64{3.0, 2.0, 1.0}}
	b := CandidateScore{FloaterScores: []float64{3.0, 2.0}}
	if a.Compare(&b) != 1 {
		t.Errorf("more floaters should be worse, got %d", a.Compare(&b))
	}
}

func TestCandidateScore_Compare_ViolationsLexicographic(t *testing.T) {
	// C8 (index 0) has higher priority than C10 (index 1).
	a := CandidateScore{Violations: [NumViolations]int{0, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
	b := CandidateScore{Violations: [NumViolations]int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
	// b has C8=1, a has C8=0 → b is worse.
	if b.Compare(&a) != 1 {
		t.Errorf("C8 violation should make b worse, got %d", b.Compare(&a))
	}
}

func TestCandidateScore_Compare_FloatersThenViolations(t *testing.T) {
	// C7 (floater scores) is checked before C8-C21 (violations).
	a := CandidateScore{
		FloaterScores: []float64{3.0},
		Violations:    [NumViolations]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	b := CandidateScore{
		FloaterScores: []float64{2.0},
		Violations:    [NumViolations]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}
	// b has lower floater scores → b is better despite all violations.
	if b.Compare(&a) != -1 {
		t.Errorf("lower floater score should win over violations, got %d", b.Compare(&a))
	}
}

func TestCandidateScore_IsPerfect(t *testing.T) {
	perfect := CandidateScore{}
	if !perfect.IsPerfect() {
		t.Error("zero score should be perfect")
	}

	imperfect := CandidateScore{Violations: [NumViolations]int{0, 1}}
	if imperfect.IsPerfect() {
		t.Error("non-zero violation should not be perfect")
	}

	withFloaters := CandidateScore{FloaterScores: []float64{1.0}}
	if withFloaters.IsPerfect() {
		t.Error("score with floaters should not be perfect")
	}
}

func TestDutchByeSelector_C9_FewestUnplayedGames(t *testing.T) {
	// Two players with same score, different game counts.
	// C9: minimize unplayed games of PAB receiver → prefer the one with
	// MORE games played (fewer unplayed games).
	players := []*PlayerState{
		{ID: "p1", TPN: 3, Score: 1.0, ColorHistory: []Color{ColorWhite, ColorWhite, ColorWhite}}, // 3 games
		{ID: "p2", TPN: 4, Score: 1.0, ColorHistory: []Color{ColorWhite}},                         // 1 game
	}

	selector := DutchByeSelector{}
	selected := selector.SelectBye(players)

	// p1 has 3 games (fewer unplayed), so should be preferred for bye per C9.
	if selected.ID != "p1" {
		t.Errorf("expected p1 (more games played, fewer unplayed), got %s", selected.ID)
	}
}

func TestDutchByeSelector_C9_SameGamesFallsBackToTPN(t *testing.T) {
	// Two players with same score and same game count → TPN tiebreak.
	players := []*PlayerState{
		{ID: "p1", TPN: 3, Score: 1.0, ColorHistory: []Color{ColorWhite}},
		{ID: "p2", TPN: 5, Score: 1.0, ColorHistory: []Color{ColorBlack}},
	}

	selector := DutchByeSelector{}
	selected := selector.SelectBye(players)

	// Same games played, so highest TPN wins.
	if selected.ID != "p2" {
		t.Errorf("expected p2 (higher TPN), got %s", selected.ID)
	}
}
