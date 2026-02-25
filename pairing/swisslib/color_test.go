package swisslib

import "testing"

func TestComputeColorPreference_NoHistory(t *testing.T) {
	pref := ComputeColorPreference(nil)
	if pref.AbsolutePreference {
		t.Error("no history should give no absolute preference")
	}
	if pref.StrongPreference {
		t.Error("no history should give no strong preference")
	}
	if pref.Color != nil {
		t.Error("no history should give no color preference")
	}
}

func TestComputeColorPreference_OneGame(t *testing.T) {
	pref := ComputeColorPreference([]Color{ColorWhite})
	if pref.AbsolutePreference {
		t.Error("one game should not give absolute preference")
	}
	// bbpPairings: colorImbalance=1, !absoluteColorPreference() → strongColorPreference=true
	if !pref.StrongPreference {
		t.Error("one game with imbalance=1 should give strong preference per bbpPairings")
	}
	if pref.Color == nil || *pref.Color != ColorBlack {
		t.Error("after White, preference should be Black")
	}
}

func TestComputeColorPreference_TwoSameColor(t *testing.T) {
	pref := ComputeColorPreference([]Color{ColorWhite, ColorWhite})
	// Absolute: 2 consecutive same color → absolute (bbpPairings: consecutiveCount > 1)
	if !pref.AbsolutePreference {
		t.Error("two consecutive White should give absolute preference")
	}
	if pref.Color == nil || *pref.Color != ColorBlack {
		t.Error("two consecutive White should prefer Black")
	}
}

func TestComputeColorPreference_AbsolutePreference(t *testing.T) {
	// 3 whites, 0 blacks → colorImbalance = 3 (>1) → absolute Black
	pref := ComputeColorPreference([]Color{ColorWhite, ColorWhite, ColorWhite})
	if !pref.AbsolutePreference {
		t.Error("3 whites should give absolute preference")
	}
	if pref.Color == nil || *pref.Color != ColorBlack {
		t.Error("3 whites should prefer Black")
	}
}

func TestComputeColorPreference_AbsoluteWithMixedHistory(t *testing.T) {
	// W,W,B,W,W → whites=4, blacks=1, imbalance=3 → absolute Black
	// Also 2 consecutive White at end → would be absolute anyway
	pref := ComputeColorPreference([]Color{ColorWhite, ColorWhite, ColorBlack, ColorWhite, ColorWhite})
	if !pref.AbsolutePreference {
		t.Error("whites=4,blacks=1 should give absolute preference")
	}
	if pref.Color == nil || *pref.Color != ColorBlack {
		t.Error("should prefer Black")
	}
}

func TestComputeColorPreference_ByeDoesNotCount(t *testing.T) {
	// W, None (bye), W → whites=2, blacks=0 in played games
	// ColorNone should not count toward color diff or consecutive tracking
	pref := ComputeColorPreference([]Color{ColorWhite, ColorNone, ColorWhite})
	// imbalance=2 (>1) → absolute Black
	if !pref.AbsolutePreference {
		t.Error("bye should not count, imbalance=2 gives absolute preference")
	}
	if pref.Color == nil || *pref.Color != ColorBlack {
		t.Error("should prefer Black")
	}
}

func TestComputeColorPreference_AlternatingColors(t *testing.T) {
	// W,B,W,B → balanced, last=Black → mild White
	pref := ComputeColorPreference([]Color{ColorWhite, ColorBlack, ColorWhite, ColorBlack})
	if pref.AbsolutePreference {
		t.Error("alternating should not give absolute")
	}
	if pref.StrongPreference {
		t.Error("alternating should not give strong")
	}
	if pref.Color == nil || *pref.Color != ColorWhite {
		t.Error("last was Black, mild should be White")
	}
}

func TestComputeColorPreference_StrongPreference(t *testing.T) {
	// B,W → imbalance=0, consecutive=0 → mild
	// W,B,W → whites=2, blacks=1, imbalance=1, consecutive=0 → strong Black
	pref := ComputeColorPreference([]Color{ColorWhite, ColorBlack, ColorWhite})
	if pref.AbsolutePreference {
		t.Error("imbalance=1 with no consecutive should not give absolute")
	}
	if !pref.StrongPreference {
		t.Error("imbalance=1 with no consecutive should give strong preference")
	}
	if pref.Color == nil || *pref.Color != ColorBlack {
		t.Error("more whites than blacks should prefer Black")
	}
}

func TestAllocateColor_BothNoPreference(t *testing.T) {
	// Round 1, no history. Board 1 (odd): higher-ranked (lower TPN) gets white.
	white := &PlayerState{ID: "p1", TPN: 1}
	black := &PlayerState{ID: "p2", TPN: 2}
	wID, bID := AllocateColor(white, black, false, 1)
	if wID != "p1" || bID != "p2" {
		t.Errorf("higher-ranked should get white on odd board: got white=%s, black=%s", wID, bID)
	}
}

func TestAllocateColor_OneAbsolute(t *testing.T) {
	// Player A has absolute White preference → A gets white
	white := &PlayerState{
		ID:           "p1",
		TPN:          2,
		ColorHistory: []Color{ColorBlack, ColorBlack, ColorBlack}, // absolute White
	}
	black := &PlayerState{
		ID:           "p2",
		TPN:          1,
		ColorHistory: []Color{ColorWhite},
	}
	wID, bID := AllocateColor(white, black, false, 1)
	if wID != "p1" || bID != "p2" {
		t.Errorf("p1 has absolute White, should get white: got white=%s, black=%s", wID, bID)
	}
}

func TestAllocateColor_BothAbsoluteOpposite(t *testing.T) {
	// A has absolute White, B has absolute Black → both granted
	a := &PlayerState{
		ID:           "p1",
		TPN:          1,
		ColorHistory: []Color{ColorBlack, ColorBlack, ColorBlack},
	}
	b := &PlayerState{
		ID:           "p2",
		TPN:          2,
		ColorHistory: []Color{ColorWhite, ColorWhite, ColorWhite},
	}
	wID, bID := AllocateColor(a, b, false, 1)
	if wID != "p1" || bID != "p2" {
		t.Errorf("p1 absolute White, p2 absolute Black: got white=%s, black=%s", wID, bID)
	}
}

func TestAllocateColor_StrongerPreferenceWins(t *testing.T) {
	// A has absolute White (2 consecutive Black),
	// B has mild White (last was Black).
	// Absolute > mild → A gets White
	a := &PlayerState{
		ID:           "p1",
		TPN:          2,
		ColorHistory: []Color{ColorBlack, ColorBlack},
	}
	b := &PlayerState{
		ID:           "p2",
		TPN:          1,
		ColorHistory: []Color{ColorBlack},
	}
	wID, bID := AllocateColor(a, b, false, 1)
	if wID != "p1" || bID != "p2" {
		t.Errorf("p1 has absolute White > p2 mild White: got white=%s, black=%s", wID, bID)
	}
}

func TestAllocateColor_EqualPreferenceSameColor_RankBreaks(t *testing.T) {
	// Both have mild White preference, equal strength → higher-ranked gets White
	a := &PlayerState{
		ID:           "p1",
		TPN:          1,
		ColorHistory: []Color{ColorBlack},
	}
	b := &PlayerState{
		ID:           "p2",
		TPN:          2,
		ColorHistory: []Color{ColorBlack},
	}
	wID, bID := AllocateColor(a, b, false, 1)
	if wID != "p1" || bID != "p2" {
		t.Errorf("equal mild, higher-ranked p1 should get white: got white=%s, black=%s", wID, bID)
	}
}

func TestAllocateColor_DifferentStrengthOpposingPreferences(t *testing.T) {
	// A has strong White (2 consecutive Black), B has strong Black (2 consecutive White).
	// Equal strength, different colors → both satisfied.
	a := &PlayerState{
		ID:           "p1",
		TPN:          2,
		ColorHistory: []Color{ColorWhite, ColorBlack, ColorBlack},
	}
	b := &PlayerState{
		ID:           "p2",
		TPN:          1,
		ColorHistory: []Color{ColorBlack, ColorWhite, ColorWhite},
	}
	wID, bID := AllocateColor(a, b, false, 1)
	if wID != "p1" || bID != "p2" {
		t.Errorf("A wants White, B wants Black, both satisfied: got white=%s, black=%s", wID, bID)
	}
}

func TestAllocateColor_NoPreference_EvenBoard(t *testing.T) {
	// Round 1, no history. Board 2 (even): lower-ranked (higher TPN) gets white.
	a := &PlayerState{ID: "p1", TPN: 1}
	b := &PlayerState{ID: "p2", TPN: 2}
	wID, bID := AllocateColor(a, b, false, 2)
	if wID != "p2" || bID != "p1" {
		t.Errorf("on even board, lower-ranked should get white: got white=%s, black=%s", wID, bID)
	}
}

func TestAllocateColor_NoPreference_AlternatesByBoard(t *testing.T) {
	// Round 1: verify boards 1-4 alternate correctly.
	a := &PlayerState{ID: "p1", TPN: 1}
	b := &PlayerState{ID: "p2", TPN: 2}

	// Board 1 (odd): higher-ranked white
	w, bl := AllocateColor(a, b, false, 1)
	if w != "p1" || bl != "p2" {
		t.Errorf("board 1: expected p1-p2, got %s-%s", w, bl)
	}
	// Board 2 (even): lower-ranked white
	w, bl = AllocateColor(a, b, false, 2)
	if w != "p2" || bl != "p1" {
		t.Errorf("board 2: expected p2-p1, got %s-%s", w, bl)
	}
	// Board 3 (odd): higher-ranked white
	w, bl = AllocateColor(a, b, false, 3)
	if w != "p1" || bl != "p2" {
		t.Errorf("board 3: expected p1-p2, got %s-%s", w, bl)
	}
	// Board 4 (even): lower-ranked white
	w, bl = AllocateColor(a, b, false, 4)
	if w != "p2" || bl != "p1" {
		t.Errorf("board 4: expected p2-p1, got %s-%s", w, bl)
	}
}
