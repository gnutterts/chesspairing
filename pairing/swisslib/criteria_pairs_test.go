package swisslib

import "testing"

// --- C10: topscorer color diff > 2 ---

func TestPairCriterionC10_NoTopScorers(t *testing.T) {
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorWhite, ColorWhite, ColorWhite}}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorBlack, ColorBlack, ColorBlack}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{TopScorers: map[string]bool{}}
	got := PairCriterionC10(pair, ctx)
	if got != 0 {
		t.Fatalf("no topscorers → 0 violations, got %d", got)
	}
}

func TestPairCriterionC10_TopScorerBadDiff(t *testing.T) {
	// White is topscorer, already has W,W,W (diff=3). Will get black (mild pref),
	// diff becomes |3-1|=2, which is NOT > 2. So no violation from diff perspective.
	// Make it worse: W,W,W,W → diff=4, gets B → diff=|4-1|=3 > 2 → violation.
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorWhite, ColorWhite, ColorWhite, ColorWhite}}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorBlack}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{
		TopScorers:  map[string]bool{"w": true},
		IsLastRound: true,
	}
	got := PairCriterionC10(pair, ctx)
	// After AllocateColor: white has absolute W (diff=4), black has mild W.
	// White gets W → diff 5, Black gets B → diff 2.
	// White is topscorer with diff 5 > 2 → violation. Black not topscorer.
	if got < 1 {
		t.Fatalf("expected ≥1 violation for topscorer with extreme color diff, got %d", got)
	}
}

// --- C11: topscorer 3+ consecutive ---

func TestPairCriterionC11_NoTopScorers(t *testing.T) {
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorWhite, ColorWhite}}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorBlack, ColorBlack}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{TopScorers: map[string]bool{}}
	got := PairCriterionC11(pair, ctx)
	if got != 0 {
		t.Fatalf("no topscorers → 0 violations, got %d", got)
	}
}

func TestPairCriterionC11_TopScorerStreak(t *testing.T) {
	// White is topscorer, last 2 games were white. If gets white again → streak 3.
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorBlack, ColorWhite, ColorWhite}}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorWhite, ColorBlack, ColorBlack}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{
		TopScorers:  map[string]bool{"w": true},
		IsLastRound: true,
	}
	// white has strong pref for B (last 2 = W,W), black has strong pref for W (last 2 = B,B).
	// AllocateColor: white gets B, black gets W → no streak violation.
	got := PairCriterionC11(pair, ctx)
	if got != 0 {
		t.Fatalf("color allocation respects prefs → 0 violations, got %d", got)
	}
}

// --- C12: color preference ---

func TestPairCriterionC12_NoViolation(t *testing.T) {
	// White wants white (last=B), black wants black (last=W) → both satisfied.
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorBlack}}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorWhite}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{IsLastRound: false}
	got := PairCriterionC12(pair, ctx)
	if got != 0 {
		t.Fatalf("expected 0 violations, got %d", got)
	}
}

func TestPairCriterionC12_OneViolation(t *testing.T) {
	// Both want white (last=B for both) → one must get black → 1 violation.
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorBlack}}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorBlack}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{IsLastRound: false}
	got := PairCriterionC12(pair, ctx)
	if got != 1 {
		t.Fatalf("expected 1 violation, got %d", got)
	}
}

func TestPairCriterionC12_NoHistory(t *testing.T) {
	// No color history → no preference → 0 violations.
	white := &PlayerState{ID: "w"}
	black := &PlayerState{ID: "b"}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{}
	got := PairCriterionC12(pair, ctx)
	if got != 0 {
		t.Fatalf("no history → no preference → 0 violations, got %d", got)
	}
}

// --- C13: strong/absolute color preference ---

func TestPairCriterionC13_NoStrongPreference(t *testing.T) {
	// Only mild preferences → 0 violations for C13.
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorBlack}}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorWhite}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{}
	got := PairCriterionC13(pair, ctx)
	if got != 0 {
		t.Fatalf("mild prefs only → 0 C13 violations, got %d", got)
	}
}

func TestPairCriterionC13_StrongSatisfied(t *testing.T) {
	// White has strong W (last 2 = B,B), black has strong B (W,W) → both satisfied.
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorBlack, ColorBlack}}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorWhite, ColorWhite}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{}
	got := PairCriterionC13(pair, ctx)
	if got != 0 {
		t.Fatalf("strong prefs satisfied → 0 violations, got %d", got)
	}
}

func TestPairCriterionC13_SameStrongViolation(t *testing.T) {
	// Both have strong W (last 2 = B,B) → one won't get it → 1 violation.
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorBlack, ColorBlack}}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorBlack, ColorBlack}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{}
	got := PairCriterionC13(pair, ctx)
	if got != 1 {
		t.Fatalf("one strong pref violated → 1 violation, got %d", got)
	}
}

// --- C14: downfloater also downfloated previous round ---

func TestPairCriterionC14_Downfloater(t *testing.T) {
	// White downfloated in round 1, now in round 2 downfloating again.
	white := &PlayerState{ID: "w", FloatHistory: []Float{FloatDown}}
	black := &PlayerState{ID: "b"}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 2}
	isDown := map[string]bool{"w": true}
	got := PairCriterionC14(pair, ctx, isDown)
	if got != 1 {
		t.Fatalf("expected 1 violation, got %d", got)
	}
}

func TestPairCriterionC14_NotDownfloater(t *testing.T) {
	white := &PlayerState{ID: "w", FloatHistory: []Float{FloatDown}}
	black := &PlayerState{ID: "b"}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 2}
	isDown := map[string]bool{} // white is NOT a downfloater this round
	got := PairCriterionC14(pair, ctx, isDown)
	if got != 0 {
		t.Fatalf("not a downfloater → 0 violations, got %d", got)
	}
}

func TestPairCriterionC14_Round1(t *testing.T) {
	// Round 1: no previous round → prevRoundIdx = -1 → 0 violations.
	white := &PlayerState{ID: "w"}
	black := &PlayerState{ID: "b"}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 1}
	isDown := map[string]bool{"w": true}
	got := PairCriterionC14(pair, ctx, isDown)
	if got != 0 {
		t.Fatalf("round 1, no history → 0 violations, got %d", got)
	}
}

// --- C15: MDP opponent upfloated previous round ---

func TestPairCriterionC15_OpponentUpfloated(t *testing.T) {
	// White is downfloater, black upfloated last round → violation.
	white := &PlayerState{ID: "w"}
	black := &PlayerState{ID: "b", FloatHistory: []Float{FloatUp}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 2}
	isDown := map[string]bool{"w": true}
	got := PairCriterionC15(pair, ctx, isDown)
	if got != 1 {
		t.Fatalf("expected 1 violation, got %d", got)
	}
}

func TestPairCriterionC15_NeitherDownfloater(t *testing.T) {
	white := &PlayerState{ID: "w"}
	black := &PlayerState{ID: "b", FloatHistory: []Float{FloatUp}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 2}
	isDown := map[string]bool{}
	got := PairCriterionC15(pair, ctx, isDown)
	if got != 0 {
		t.Fatalf("neither is downfloater → 0 violations, got %d", got)
	}
}

// --- C16: downfloater also downfloated 2 rounds ago ---

func TestPairCriterionC16_TwoRoundsAgo(t *testing.T) {
	// Round 3: white downfloated in round 1 (idx 0) and is downfloating now.
	white := &PlayerState{ID: "w", FloatHistory: []Float{FloatDown, FloatNone}}
	black := &PlayerState{ID: "b"}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 3}
	isDown := map[string]bool{"w": true}
	got := PairCriterionC16(pair, ctx, isDown)
	if got != 1 {
		t.Fatalf("expected 1 violation, got %d", got)
	}
}

// --- C17: MDP opponent upfloated 2 rounds ago ---

func TestPairCriterionC17_OpponentUpfloated2Ago(t *testing.T) {
	// Round 3: white is downfloater, black upfloated in round 1 (idx 0).
	white := &PlayerState{ID: "w"}
	black := &PlayerState{ID: "b", FloatHistory: []Float{FloatUp, FloatNone}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 3}
	isDown := map[string]bool{"w": true}
	got := PairCriterionC17(pair, ctx, isDown)
	if got != 1 {
		t.Fatalf("expected 1 violation, got %d", got)
	}
}

// --- C18: downfloater with score diff > 0 ---

func TestPairCriterionC18_ScoreDiffPositive(t *testing.T) {
	white := &PlayerState{ID: "w", Score: 3.0}
	black := &PlayerState{ID: "b", Score: 2.0}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{}
	isDown := map[string]bool{"w": true}
	bracketScore := 2.0
	got := PairCriterionC18(pair, ctx, isDown, bracketScore)
	if got != 1 {
		t.Fatalf("downfloater w/ score 3 in bracket 2 → diff > 0 → 1 violation, got %d", got)
	}
}

func TestPairCriterionC18_ScoreDiffZero(t *testing.T) {
	white := &PlayerState{ID: "w", Score: 2.0}
	black := &PlayerState{ID: "b", Score: 2.0}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{}
	isDown := map[string]bool{"w": true}
	bracketScore := 2.0
	got := PairCriterionC18(pair, ctx, isDown, bracketScore)
	if got != 0 {
		t.Fatalf("score diff = 0 → 0 violations, got %d", got)
	}
}

// --- C19: MDP opponent upfloated previous round with score diff > 0 ---

func TestPairCriterionC19_UpfloatedOpponentWithDiff(t *testing.T) {
	// White is downfloater, black upfloated last round and has score > bracket.
	white := &PlayerState{ID: "w", Score: 2.0}
	black := &PlayerState{ID: "b", Score: 3.0, FloatHistory: []Float{FloatUp}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 2}
	isDown := map[string]bool{"w": true}
	bracketScore := 2.0
	got := PairCriterionC19(pair, ctx, isDown, bracketScore)
	if got != 1 {
		t.Fatalf("expected 1 violation, got %d", got)
	}
}

// --- C20: downfloater also downfloated 2 rounds ago with score diff > 0 ---

func TestPairCriterionC20_TwoRoundsAgoWithDiff(t *testing.T) {
	white := &PlayerState{ID: "w", Score: 3.0, FloatHistory: []Float{FloatDown, FloatNone}}
	black := &PlayerState{ID: "b", Score: 2.0}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 3}
	isDown := map[string]bool{"w": true}
	bracketScore := 2.0
	got := PairCriterionC20(pair, ctx, isDown, bracketScore)
	if got != 1 {
		t.Fatalf("expected 1 violation, got %d", got)
	}
}

// --- C21: MDP opponent upfloated 2 rounds ago with score diff > 0 ---

func TestPairCriterionC21_UpfloatedOpponent2AgoWithDiff(t *testing.T) {
	white := &PlayerState{ID: "w", Score: 2.0}
	black := &PlayerState{ID: "b", Score: 3.0, FloatHistory: []Float{FloatUp, FloatNone}}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{CurrentRound: 3}
	isDown := map[string]bool{"w": true}
	bracketScore := 2.0
	got := PairCriterionC21(pair, ctx, isDown, bracketScore)
	if got != 1 {
		t.Fatalf("expected 1 violation, got %d", got)
	}
}

// --- ComputeEdgeWeight ---

// expectedWeight computes the expected edge weight from violations using
// the 3-bit-per-criterion layout (matching ComputeEdgeWeight internals).
func expectedWeight(violations [12]int) int64 {
	const bitsPerCriterion = 3
	const maxVal = 7
	var w int64
	for i := 0; i < 12; i++ {
		v := violations[i]
		if v > maxVal {
			v = maxVal
		}
		inverted := int64(maxVal - v)
		shift := EdgeWeightTieBreakBits + (11-i)*bitsPerCriterion
		w |= inverted << shift
	}
	return w
}

func TestEdgeWeight_AllGood(t *testing.T) {
	violations := [12]int{} // all zeros = no violations
	w := ComputeEdgeWeight(violations)
	expected := expectedWeight(violations)
	if w != expected {
		t.Fatalf("expected %d, got %d", expected, w)
	}
	// Should be the maximum possible weight (all fields = 7).
	if w == 0 {
		t.Fatal("all-good weight should be non-zero")
	}
}

func TestEdgeWeight_AllBad(t *testing.T) {
	violations := [12]int{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}
	w := ComputeEdgeWeight(violations)
	if w != 0 {
		t.Fatalf("all max violations → weight 0, got %d", w)
	}
}

func TestEdgeWeight_C10Violated(t *testing.T) {
	violations := [12]int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	w := ComputeEdgeWeight(violations)
	expected := expectedWeight(violations)
	if w != expected {
		t.Fatalf("expected %d, got %d", expected, w)
	}
	// Should be less than all-good.
	allGood := ComputeEdgeWeight([12]int{})
	if w >= allGood {
		t.Fatalf("C10 violated weight (%d) should be less than all-good (%d)", w, allGood)
	}
}

func TestEdgeWeight_Ordering(t *testing.T) {
	// Violating only C21 (lowest priority) should yield higher weight
	// than violating only C10 (highest priority).
	vC10 := [12]int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	vC21 := [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	wC10 := ComputeEdgeWeight(vC10)
	wC21 := ComputeEdgeWeight(vC21)
	if wC21 <= wC10 {
		t.Fatalf("C21 violation (%d) should produce higher weight than C10 violation (%d)", wC21, wC10)
	}
}

func TestEdgeWeight_PriorityOrder(t *testing.T) {
	// Verify each criterion violation produces a weight less than the previous.
	for i := 0; i < 11; i++ {
		var vHigh, vLow [12]int
		vHigh[i] = 1  // violate higher priority criterion
		vLow[i+1] = 1 // violate lower priority criterion
		wHigh := ComputeEdgeWeight(vHigh)
		wLow := ComputeEdgeWeight(vLow)
		if wLow <= wHigh {
			t.Fatalf("C%d violation (w=%d) should produce higher weight than C%d violation (w=%d)",
				10+i+1, wLow, 10+i, wHigh)
		}
	}
}

func TestEdgeWeight_ViolationCountGranularity(t *testing.T) {
	// 1 violation should give higher weight than 2 violations of same criterion.
	v1 := [12]int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	v2 := [12]int{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	w1 := ComputeEdgeWeight(v1)
	w2 := ComputeEdgeWeight(v2)
	if w1 <= w2 {
		t.Fatalf("1 violation (w=%d) should produce higher weight than 2 violations (w=%d)", w1, w2)
	}
}

// --- PairEdgeWeight integration ---

func TestPairEdgeWeight_PerfectPair(t *testing.T) {
	// Perfect pair: compatible preferences, no floats, no topscorers.
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorBlack}, Score: 1.0}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorWhite}, Score: 1.0}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{
		TopScorers:   map[string]bool{},
		CurrentRound: 2,
	}
	isDown := map[string]bool{}
	w := PairEdgeWeight(pair, ctx, isDown, 1.0)
	// Should be max weight (all criteria satisfied).
	maxWeight := ComputeEdgeWeight([12]int{})
	if w != maxWeight {
		t.Fatalf("perfect pair should have max weight %d, got %d", maxWeight, w)
	}
}

func TestPairEdgeWeight_ColorConflict(t *testing.T) {
	// Both want white → C12 violated (1 violation).
	// With bbpPairings-correct color preferences, [ColorBlack] gives strong
	// preference (imbalance=1), so C13 is also violated (1 violation).
	white := &PlayerState{ID: "w", ColorHistory: []Color{ColorBlack}, Score: 1.0}
	black := &PlayerState{ID: "b", ColorHistory: []Color{ColorBlack}, Score: 1.0}
	pair := &ProposedPairing{White: white, Black: black}
	ctx := &CriteriaContext{
		TopScorers:   map[string]bool{},
		CurrentRound: 2,
	}
	isDown := map[string]bool{}
	w := PairEdgeWeight(pair, ctx, isDown, 1.0)
	// C12 (index 2) has 1 violation, C13 (index 3) has 1 violation.
	var expected [12]int
	expected[2] = 1 // C12 violated (mild/strong color preference denied)
	expected[3] = 1 // C13 violated (strong color preference denied)
	expectedW := ComputeEdgeWeight(expected)
	if w != expectedW {
		t.Fatalf("expected weight %d (C12+C13 violated), got %d", expectedW, w)
	}
}
