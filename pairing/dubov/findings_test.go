// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"context"
	"errors"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func playerIDs(players []*swisslib.PlayerState) []string {
	ids := make([]string, len(players))
	for i, p := range players {
		ids[i] = p.ID
	}
	return ids
}

func inGroup(players []*swisslib.PlayerState, id string) bool {
	for _, p := range players {
		if p.ID == id {
			return true
		}
	}
	return false
}

// TestDUBF1_ByeRequiresCompletableRest verifies Article 3.1.2: the bye goes
// to the first eligible candidate whose removal leaves a complete pairing.
func TestDUBF1_ByeRequiresCompletableRest(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", Rating: 2000},
		{ID: "p2", Rating: 1900},
		{ID: "p3", Rating: 1800},
		{ID: "p4", Rating: 1700},
		{ID: "p5", Rating: 1600},
	}
	// Only (p1,p5) and (p2,p3) may be legal; everything else is forbidden.
	forbidden := [][]string{
		{"p1", "p2"},
		{"p1", "p3"},
		{"p1", "p4"},
		{"p2", "p4"},
		{"p2", "p5"},
		{"p3", "p4"},
		{"p3", "p5"},
		{"p4", "p5"},
	}
	state := &chesspairing.TournamentState{Players: players, CurrentRound: 1}
	result, err := New(Options{TotalRounds: dubovTestRounds(), ForbiddenPairs: forbidden}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error = %v", err)
	}
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p4" {
		t.Fatalf("bye = %v, want p4 (only p4 leaves p1-p5, p2-p3 pairable)", result.Byes)
	}
	assertPaired(t, result, "p1", "p5")
	assertPaired(t, result, "p2", "p3")
}

// TestDUBF2_UpfloatersPulledUp verifies Articles 3.2.1-3.2.2/4.2: the lower
// player is pulled up into the higher bracket instead of a resident being
// downfloated.
func TestDUBF2_UpfloatersPulledUp(t *testing.T) {
	p1 := &swisslib.PlayerState{ID: "p1", TPN: 1, Score: 2.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}}
	p2 := &swisslib.PlayerState{ID: "p2", TPN: 2, Score: 2.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}}
	p3 := &swisslib.PlayerState{ID: "p3", TPN: 3, Score: 2.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}}
	p4 := &swisslib.PlayerState{ID: "p4", TPN: 4, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}}
	mctx := &matchContext{ratings: map[string]int{}, totalRounds: 9}
	pairs, err := processBrackets(context.Background(), []*swisslib.PlayerState{p1, p2, p3, p4}, mctx)
	if err != nil {
		t.Fatalf("processBrackets() error = %v", err)
	}
	for _, pair := range pairs {
		if pair.white.ID == "p4" || pair.black.ID == "p4" {
			if pair.bracketScore != 2.0 {
				t.Fatalf("p4 pair bracketScore = %v, want 2.0 (p4 upfloated into the 2.0 bracket)", pair.bracketScore)
			}
			return
		}
	}
	t.Fatalf("p4 (the 1.0 player) was not paired")
}

// TestDUBF3_ShiftersNumberedFromMiddle verifies Article 4.3.3 with White
// seekers: the middle-ARO player is shifted, not the highest-ARO player.
func TestDUBF3_ShiftersNumberedFromMiddle(t *testing.T) {
	ratings := map[string]int{"x": 1600, "y": 1200, "z": 1000, "w": 1400}
	// ARO order differs from TPN order: C(1000), B(1200), D(1400), A(1600).
	g1 := []*swisslib.PlayerState{
		{ID: "A", TPN: 1, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"x"}},
		{ID: "B", TPN: 2, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"y"}},
		{ID: "C", TPN: 3, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"z"}},
		{ID: "D", TPN: 4, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"w"}},
	}
	g2 := []*swisslib.PlayerState{
		{ID: "E", TPN: 5, ColorHistory: []swisslib.Color{swisslib.ColorWhite}},
		{ID: "F", TPN: 6, ColorHistory: []swisslib.Color{swisslib.ColorWhite}},
	}
	bg1, bg2 := BalanceG1G2(g1, g2, ratings)
	if len(bg1) != 3 || len(bg2) != 3 {
		t.Fatalf("balanced groups = %d+%d, want 3+3", len(bg1), len(bg2))
	}
	if !inGroup(bg2, "B") {
		t.Fatalf("shifted player missing B (middle ARO); G2 = %v", playerIDs(bg2))
	}
	if inGroup(bg2, "A") {
		t.Fatalf("A (highest ARO) must not be shifted; G2 = %v", playerIDs(bg2))
	}
}

// TestDUBF4_FirstLegalTranspositionBeyond120 verifies Article 3.2.6: the
// first legal transposition is used even when it is number 121.
func TestDUBF4_FirstLegalTranspositionBeyond120(t *testing.T) {
	var g1, g2 []*swisslib.PlayerState
	for i := 1; i <= 6; i++ {
		g1 = append(g1, &swisslib.PlayerState{ID: "a" + string(rune('0'+i)), TPN: i, ColorHistory: []swisslib.Color{swisslib.ColorBlack}})
	}
	for i := 1; i <= 6; i++ {
		g2 = append(g2, &swisslib.PlayerState{ID: "b" + string(rune('0'+i)), TPN: i, ColorHistory: []swisslib.Color{swisslib.ColorWhite}})
	}
	// Legal cross-pairs: a1-b2, a2-b1, and ai-bi for i >= 3.
	forbidden := map[[2]string]bool{}
	for i := 1; i <= 6; i++ {
		for j := 1; j <= 6; j++ {
			legal := (i == 1 && j == 2) || (i == 2 && j == 1) || (i >= 3 && i == j)
			if !legal {
				forbidden[swisslib.CanonicalPairKey(g1[i-1].ID, g2[j-1].ID)] = true
			}
		}
	}
	mctx := &matchContext{ratings: map[string]int{}, forbidden: forbidden, totalRounds: 9}
	players := append(append([]*swisslib.PlayerState{}, g1...), g2...)
	result, err := MatchBracketDubov(swisslib.Bracket{Players: players, OriginalScore: 0}, mctx)
	if err != nil {
		t.Fatalf("MatchBracketDubov() error = %v", err)
	}
	if result == nil {
		t.Fatalf("MatchBracketDubov() = nil, want first legal transposition (#121)")
	}
	for _, pair := range result.pairs {
		if pair.white.ID == "a1" {
			if pair.black.ID != "b2" {
				t.Fatalf("a1 paired with %s, want b2", pair.black.ID)
			}
			return
		}
	}
	t.Fatalf("a1 not paired")
}

// TestDUBF5_ARORoundedHalfUp verifies Article 1.7.1: ARO is rounded half up
// before comparison, with TPN as the tie-break.
func TestDUBF5_ARORoundedHalfUp(t *testing.T) {
	ratings := map[string]int{"o1": 1000, "o2": 1001, "o3": 1001}
	a := &swisslib.PlayerState{ID: "A", TPN: 2, PairingNumber: 2, Opponents: []string{"o1", "o2"}}       // raw 1000.5 -> 1001
	b := &swisslib.PlayerState{ID: "B", TPN: 1, PairingNumber: 1, Opponents: []string{"o1", "o2", "o3"}} // raw 1000.667 -> 1001
	players := []*swisslib.PlayerState{a, b}
	SortByAROAscending(players, ratings)
	if players[0].ID != "B" {
		t.Fatalf("first player = %s, want B (TPN 1 breaks the rounded-ARO tie)", players[0].ID)
	}
}

// TestDUBF6_MaxTUsesTotalRounds verifies Article 1.8.2: MaxT uses the planned
// number of rounds, not the number of completed rounds.
func TestDUBF6_MaxTUsesTotalRounds(t *testing.T) {
	totalRounds := 9
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a", Rating: 2000}, {ID: "b", Rating: 1900}},
		Rounds: []chesspairing.RoundData{
			{Number: 1}, {Number: 2}, {Number: 3}, {Number: 4},
		},
		CurrentRound: 5,
	}
	mctx := newMatchContext(state, Options{TotalRounds: &totalRounds})
	if MaxT(mctx.totalRounds) != 3 {
		t.Fatalf("MaxT(totalRounds) = %d, want 3 (2 + floor(9/5))", MaxT(mctx.totalRounds))
	}
	if MaxT(len(state.Rounds)) != 2 {
		t.Fatalf("sanity: MaxT(completed rounds) = %d, want 2", MaxT(len(state.Rounds)))
	}
}

// TestDUBF7_C6VectorAndLastRoundRelaxation verifies Article 2.3: C6 is a
// descending vector (not a sum) and C8-C10 are dropped in the last round.
func TestDUBF7_C6VectorAndLastRoundRelaxation(t *testing.T) {
	better := candidateQuality{upfloaters: 1, differences: []float64{2, 2}}
	worse := candidateQuality{upfloaters: 1, differences: []float64{3, 1}}
	if !betterQuality(better, worse) {
		t.Fatalf("differences {2,2} should beat {3,1}")
	}

	upfloater := &swisslib.PlayerState{ID: "u", Score: 1.0, FloatHistory: []swisslib.Float{swisslib.FloatUp, swisslib.FloatUp}}
	resident := &swisslib.PlayerState{ID: "r", Score: 4.0}
	pairs := []proposedPairing{{white: resident, black: upfloater}}
	upfloaterIDs := map[string]bool{"u": true}

	last := qualityFor(pairs, upfloaterIDs, true, 2)
	if last.maximumUpfloaters != 0 || last.maximumTimes != 0 || last.previousUpfloaters != 0 {
		t.Fatalf("last round must drop C8-C10, got %+v", last)
	}
	notLast := qualityFor(pairs, upfloaterIDs, false, 2)
	if notLast.maximumUpfloaters == 0 || notLast.maximumTimes == 0 || notLast.previousUpfloaters == 0 {
		t.Fatalf("non-last round must count C8-C10, got %+v", notLast)
	}
}

// TestDUBV1_G1G2SplitWhenAllUnplayed verifies Article 3.2.3: an all-unplayed
// bracket splits by TPN halves even after round 1.
func TestDUBV1_G1G2SplitWhenAllUnplayed(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "p1", TPN: 1}, {ID: "p2", TPN: 2}, {ID: "p3", TPN: 3}, {ID: "p4", TPN: 4},
	}
	g1, g2 := SplitG1G2(players, false)
	if len(g1) != 2 || len(g2) != 2 {
		t.Fatalf("split = %d+%d, want 2+2", len(g1), len(g2))
	}
	if g1[0].ID != "p1" || g1[1].ID != "p2" {
		t.Fatalf("G1 = %v, want first half by TPN", playerIDs(g1))
	}
	if g2[0].ID != "p3" || g2[1].ID != "p4" {
		t.Fatalf("G2 = %v, want second half by TPN", playerIDs(g2))
	}
}

// TestNEW1_UnplayedPlayerPrefersBlack verifies Article 1.6.4: a player who
// has not played has a mild Black preference.
func TestNEW1_UnplayedPlayerPrefersBlack(t *testing.T) {
	a := &swisslib.PlayerState{ID: "a", TPN: 1, ColorHistory: nil}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, ColorHistory: []swisslib.Color{swisslib.ColorBlack, swisslib.ColorWhite}} // mild Black
	whiteID, _ := AllocateColor(a, b, 1, nil)
	if whiteID != "b" {
		t.Fatalf("white = %s, want b (unplayed a has a mild Black preference)", whiteID)
	}
}

// TestNEW3_SortsByFixedTPN verifies 4.3.2/3.2.5: ARO ties sort by fixed TPN
// (PairingNumber), not by the score-based live TPN.
func TestNEW3_SortsByFixedTPN(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, PairingNumber: 3, Opponents: []string{"x"}},
		{ID: "b", TPN: 2, PairingNumber: 1, Opponents: []string{"y"}},
	}
	SortByAROAscending(players, map[string]int{})
	if players[0].ID != "b" {
		t.Fatalf("first = %s, want b (fixed TPN 1 breaks the ARO tie)", players[0].ID)
	}
}

// TestErrBracketTooLarge verifies the enumeration guard: a bracket whose
// transposition search exceeds the safety limit returns ErrBracketTooLarge
// from Pair instead of hanging.
func TestErrBracketTooLarge(t *testing.T) {
	old := maxSearchNodes
	// This bracket has legal prefixes but no final assignment, so it needs real
	// DFS enumeration. A normally legal four-pair bracket takes only seven
	// shared nodes (two recomposition nodes plus five DFS nodes).
	maxSearchNodes = 10
	t.Cleanup(func() { maxSearchNodes = old })

	players := []chesspairing.PlayerEntry{
		{ID: "p1", Rating: 2000},
		{ID: "p2", Rating: 1900},
		{ID: "p3", Rating: 1800},
		{ID: "p4", Rating: 1700},
		{ID: "p5", Rating: 1600},
		{ID: "p6", Rating: 1500},
		{ID: "p7", Rating: 1400},
		{ID: "p8", Rating: 1300},
	}
	// p4 can be paired with none of G2, after p1-p3 have explored legal
	// prefixes. This exceeds the cap only through actual enumeration.
	forbidden := [][]string{{"p4", "p5"}, {"p4", "p6"}, {"p4", "p7"}, {"p4", "p8"}}
	state := &chesspairing.TournamentState{Players: players, CurrentRound: 1}
	_, err := New(Options{TotalRounds: dubovTestRounds(), ForbiddenPairs: forbidden}).Pair(context.Background(), state)
	if !errors.Is(err, ErrBracketTooLarge) {
		t.Fatalf("Pair() error = %v, want ErrBracketTooLarge", err)
	}
}

// TestIncompleteCarriesTopGroupPairs verifies Article 3.2.4's first stage:
// when same-subgroup pairs are unavoidable, players are shifted out of the
// smaller subgroup (G1 when the sizes are equal) so the whole bracket pairs.
func TestIncompleteCarriesTopGroupPairs(t *testing.T) {
	t1 := &swisslib.PlayerState{ID: "t1", TPN: 1, PairingNumber: 1, Score: 3.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}}
	t2 := &swisslib.PlayerState{ID: "t2", TPN: 2, PairingNumber: 2, Score: 3.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}}
	a := &swisslib.PlayerState{ID: "a", TPN: 3, PairingNumber: 3, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"c", "d"}}
	b := &swisslib.PlayerState{ID: "b", TPN: 4, PairingNumber: 4, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"c", "d"}}
	c := &swisslib.PlayerState{ID: "c", TPN: 5, PairingNumber: 5, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"a", "b"}}
	d := &swisslib.PlayerState{ID: "d", TPN: 6, PairingNumber: 6, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"a", "b"}}
	mctx := &matchContext{ratings: map[string]int{}, totalRounds: 9}
	pairs, err := processBrackets(context.Background(), []*swisslib.PlayerState{t1, t2, a, b, c, d}, mctx)
	if err != nil {
		t.Fatalf("processBrackets() error = %v", err)
	}
	if len(pairs) != 3 {
		t.Fatalf("pairs = %d, want 3 (t1-t2, a-b, c-d)", len(pairs))
	}
	assertPairedPlayers(t, pairs, "t1", "t2")
	assertPairedPlayers(t, pairs, "a", "b")
	assertPairedPlayers(t, pairs, "c", "d")
}

// TestFirstStage324ShiftsUnavoidablePair verifies the first sentence of
// Article 3.2.4 directly: a G1-G1 pair (here a-b) is unavoidable because every
// G1-G2 pairing is a rematch, so one G1 player is shifted and the bracket
// pairs completely.
func TestFirstStage324ShiftsUnavoidablePair(t *testing.T) {
	a := &swisslib.PlayerState{ID: "a", TPN: 1, PairingNumber: 1, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"c", "d"}}
	b := &swisslib.PlayerState{ID: "b", TPN: 2, PairingNumber: 2, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"c", "d"}}
	c := &swisslib.PlayerState{ID: "c", TPN: 3, PairingNumber: 3, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"a", "b"}}
	d := &swisslib.PlayerState{ID: "d", TPN: 4, PairingNumber: 4, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"a", "b"}}
	mctx := &matchContext{ratings: map[string]int{}, totalRounds: 9}
	result, err := MatchBracketDubov(swisslib.Bracket{Players: []*swisslib.PlayerState{a, b, c, d}, OriginalScore: 1}, mctx)
	if err != nil {
		t.Fatalf("MatchBracketDubov() error = %v", err)
	}
	if result == nil {
		t.Fatalf("MatchBracketDubov() = nil, want a legal pairing after shifting one G1 player")
	}
	assertPairedPlayers(t, result.pairs, "a", "b")
	assertPairedPlayers(t, result.pairs, "c", "d")
}

// TestEqualisation324DoesNotShiftFirstStagePlayer verifies the second stage
// draws only from the colour seekers in its source subgroup (4.3.1-4.3.2).
func TestEqualisation324DoesNotShiftFirstStagePlayer(t *testing.T) {
	white := func(id string, tpn int) *swisslib.PlayerState {
		return &swisslib.PlayerState{ID: id, TPN: tpn, ColorHistory: []swisslib.Color{swisslib.ColorBlack}}
	}
	black := func(id string, tpn int) *swisslib.PlayerState {
		return &swisslib.PlayerState{ID: id, TPN: tpn, ColorHistory: []swisslib.Color{swisslib.ColorWhite}}
	}
	g1 := []*swisslib.PlayerState{white("w1", 1), white("w2", 2), white("w3", 3), white("w4", 4)}
	g2 := []*swisslib.PlayerState{black("b1", 5), black("b2", 6)}
	var got *recomposition
	err := forEachRecomposition(g1, g2, map[string]int{}, nil, func(r recomposition) (bool, error) {
		if inGroup(r.g1, "b1") { // b1 was the first G2 shifter at s=1.
			got = &r
			return true, nil
		}
		return false, nil
	})
	if err != nil || got == nil {
		t.Fatalf("recomposition = %v, %v; want the s=1 recomposition", got, err)
	}
	if !inGroup(got.g1, "b1") {
		t.Fatalf("first-stage shifter b1 was shifted back: G1 = %v", playerIDs(got.g1))
	}
	// The 4.3 order of w1..w4 is w2,w3,w1,w4, so the equalisation set is w2,w3.
	if !inGroup(got.g2, "w2") || !inGroup(got.g2, "w3") {
		t.Fatalf("equalisation set = G2 %v, want w2,w3", playerIDs(got.g2))
	}
}

// TestUpfloaterSetsFollowC6NotArrivalOrder verifies that the upfloaters for
// a bracket are chosen by C6 over the candidate sets with equal score
// multisets, not by the order in which the sets were generated: B and C
// (1.0) upfloat to R1 and R2, not A (2.0) with D (0.0).
func TestUpfloaterSetsFollowC6NotArrivalOrder(t *testing.T) {
	r1 := &swisslib.PlayerState{ID: "R1", TPN: 1, PairingNumber: 1, Score: 3.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"R2"}}
	r2 := &swisslib.PlayerState{ID: "R2", TPN: 2, PairingNumber: 2, Score: 3.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"R1"}}
	a := &swisslib.PlayerState{ID: "A", TPN: 3, PairingNumber: 3, Score: 2.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}}
	b := &swisslib.PlayerState{ID: "B", TPN: 4, PairingNumber: 4, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"D"}}
	c := &swisslib.PlayerState{ID: "C", TPN: 5, PairingNumber: 5, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"D"}}
	d := &swisslib.PlayerState{ID: "D", TPN: 6, PairingNumber: 6, Score: 0.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"B", "C"}}
	mctx := &matchContext{ratings: map[string]int{}, totalRounds: 9}
	pairs, err := processBrackets(context.Background(), []*swisslib.PlayerState{r1, r2, a, b, c, d}, mctx)
	if err != nil {
		t.Fatalf("processBrackets() error = %v", err)
	}
	for _, pair := range pairs {
		for _, id := range []string{"R1", "R2"} {
			if pair.white.ID != id && pair.black.ID != id {
				continue
			}
			other := pair.white.ID
			if other == id {
				other = pair.black.ID
			}
			if other != "B" && other != "C" {
				t.Errorf("%s paired with %s, want an upfloater from {B, C}", id, other)
			}
		}
	}
}

// TestUpfloaterSetsWithEqualMultisetAreNotContiguous verifies that all
// upfloater sets with the same score multiset are compared on C6, also when
// they are not generated next to each other: R1-D and C-R2.
func TestUpfloaterSetsWithEqualMultisetAreNotContiguous(t *testing.T) {
	r1 := &swisslib.PlayerState{ID: "R1", TPN: 1, PairingNumber: 1, Score: 3.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"R2", "A", "B", "C"}}
	r2 := &swisslib.PlayerState{ID: "R2", TPN: 2, PairingNumber: 2, Score: 3.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"R1", "A"}}
	a := &swisslib.PlayerState{ID: "A", TPN: 3, PairingNumber: 3, Score: 2.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"R1", "R2", "C", "E", "F"}}
	b := &swisslib.PlayerState{ID: "B", TPN: 4, PairingNumber: 4, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"R1"}}
	c := &swisslib.PlayerState{ID: "C", TPN: 5, PairingNumber: 5, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"R1", "A"}}
	d := &swisslib.PlayerState{ID: "D", TPN: 6, PairingNumber: 6, Score: 1.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}}
	e := &swisslib.PlayerState{ID: "E", TPN: 7, PairingNumber: 7, Score: 0.0, ColorHistory: []swisslib.Color{swisslib.ColorWhite}, Opponents: []string{"A"}}
	f := &swisslib.PlayerState{ID: "F", TPN: 8, PairingNumber: 8, Score: 0.0, ColorHistory: []swisslib.Color{swisslib.ColorBlack}, Opponents: []string{"A"}}
	mctx := &matchContext{ratings: map[string]int{}, totalRounds: 9}
	pairs, err := processBrackets(context.Background(), []*swisslib.PlayerState{r1, r2, a, b, c, d, e, f}, mctx)
	if err != nil {
		t.Fatalf("processBrackets() error = %v", err)
	}
	assertPairedPlayers(t, pairs, "R1", "D")
	assertPairedPlayers(t, pairs, "C", "R2")
}

func assertPairedPlayers(t *testing.T, pairs []proposedPairing, id1, id2 string) {
	t.Helper()
	for _, pair := range pairs {
		if (pair.white.ID == id1 && pair.black.ID == id2) || (pair.white.ID == id2 && pair.black.ID == id1) {
			return
		}
	}
	t.Errorf("expected pairing %s vs %s not found", id1, id2)
}
