package team

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestAllocateColor_BothNewTeams_OddTPN(t *testing.T) {
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1}
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2}
	initialColor := "white"

	wID, bID := AllocateColor(a, b, ColorPrefTypeA, false, &initialColor, nil)
	// First-team = t1 (smaller TPN, equal scores). TPN=1 is odd → initial-colour (White).
	if wID != "t1" || bID != "t2" {
		t.Errorf("odd TPN: expected t1=W, t2=B; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_BothNewTeams_EvenTPN(t *testing.T) {
	a := &lexswiss.ParticipantState{ID: "t2", TPN: 2}
	b := &lexswiss.ParticipantState{ID: "t3", TPN: 3}
	initialColor := "white"

	wID, bID := AllocateColor(a, b, ColorPrefTypeA, false, &initialColor, nil)
	// First-team = t2 (smaller TPN). TPN=2 is even → opposite of initial (Black).
	if wID != "t3" || bID != "t2" {
		t.Errorf("even TPN: expected t3=W, t2=B; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_BothNewTeams_InitialBlack(t *testing.T) {
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1}
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2}
	initialColor := "black"

	wID, bID := AllocateColor(a, b, ColorPrefTypeA, false, &initialColor, nil)
	// First-team = t1 (smaller TPN). TPN=1 is odd → initial-colour (Black).
	if wID != "t2" || bID != "t1" {
		t.Errorf("initial=black, odd TPN: expected t2=W, t1=B; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_OnlyOnePref(t *testing.T) {
	// Art. 4.3.2: only one team has preference → grant it.
	W := lexswiss.ColorWhite
	B := lexswiss.ColorBlack
	// t1: pref White (CD=-1, last two Black)
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1, ColorHistory: []lexswiss.Color{W, B, B}}
	// t2: no pref (CD=0, mixed last)
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2, ColorHistory: []lexswiss.Color{W, B}}

	wID, bID := AllocateColor(a, b, ColorPrefTypeA, false, nil, nil)
	if wID != "t1" || bID != "t2" {
		t.Errorf("only one pref: expected t1=W; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_OppositePrefs(t *testing.T) {
	// Art. 4.3.3: opposite preferences → grant both.
	W := lexswiss.ColorWhite
	B := lexswiss.ColorBlack
	// t1: pref White
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1, ColorHistory: []lexswiss.Color{W, B, B}}
	// t2: pref Black
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2, ColorHistory: []lexswiss.Color{B, W, W}}

	wID, bID := AllocateColor(a, b, ColorPrefTypeA, false, nil, nil)
	if wID != "t1" || bID != "t2" {
		t.Errorf("opposite prefs: expected t1=W, t2=B; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_TypeB_OnlyOneStrong(t *testing.T) {
	// Art. 4.3.4: Type B, only one has strong preference → grant it.
	W := lexswiss.ColorWhite
	B := lexswiss.ColorBlack
	// t1: strong White
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1, ColorHistory: []lexswiss.Color{W, B, B}}
	// t2: mild Black (CD=+1, single game)
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2, ColorHistory: []lexswiss.Color{W}}

	wID, bID := AllocateColor(a, b, ColorPrefTypeB, false, nil, nil)
	// t1 has strong White → grant.
	if wID != "t1" || bID != "t2" {
		t.Errorf("one strong: expected t1=W; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_LowerColorDifference(t *testing.T) {
	// Art. 4.3.5: Give White to the team with the lower colour difference.
	W := lexswiss.ColorWhite
	B := lexswiss.ColorBlack
	// t1: CD = 0 (W, B)
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1, Score: 2, ColorHistory: []lexswiss.Color{W, B}}
	// t2: CD = +2 (W, W, B, W) → CD = +2
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2, Score: 2, ColorHistory: []lexswiss.Color{W, W, B, W}}

	wID, bID := AllocateColor(a, b, ColorPrefTypeNone, false, nil, nil)
	// No colour prefs. Lower CD → White. t1 CD=0 < t2 CD=+2 → t1 gets White.
	if wID != "t1" || bID != "t2" {
		t.Errorf("lower CD: expected t1=W; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_Alternate(t *testing.T) {
	// Art. 4.3.6: Alternate from most recent time one had W and other B.
	W := lexswiss.ColorWhite
	B := lexswiss.ColorBlack
	// Same CD. Both have same prefs. Look at history:
	// t1: W, B → last = B
	// t2: B, W → last = W
	// Most recent time one had W and other B: round 2 (t1=B, t2=W). Alternate: t1→W, t2→B.
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1, Score: 1, ColorHistory: []lexswiss.Color{W, B}}
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2, Score: 1, ColorHistory: []lexswiss.Color{B, W}}

	wID, bID := AllocateColor(a, b, ColorPrefTypeNone, false, nil, nil)
	// No prefs. Equal CD (0). Alternate: t1 had B last when t2 had W → t1 gets W.
	if wID != "t1" || bID != "t2" {
		t.Errorf("alternate: expected t1=W; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_FirstTeamPref(t *testing.T) {
	// Art. 4.3.7: Grant the colour preference of the first-team.
	W := lexswiss.ColorWhite
	B := lexswiss.ColorBlack
	// Both have same preference (e.g., both want White) — C8 failed, pairing accepted anyway.
	// First-team's preference is granted.
	// t1 (higher score, smaller TPN) is first-team, wants White.
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1, Score: 3, ColorHistory: []lexswiss.Color{W, B, B}}
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2, Score: 2, ColorHistory: []lexswiss.Color{W, B, B}}

	wID, bID := AllocateColor(a, b, ColorPrefTypeA, false, nil, nil)
	// Both want White. First-team = t1 (higher score). t1 gets White.
	if wID != "t1" || bID != "t2" {
		t.Errorf("first-team pref: expected t1=W; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_FirstTeamAlternate(t *testing.T) {
	// Art. 4.3.8: Alternate the first-team's colour from its last played round.
	W := lexswiss.ColorWhite
	// Both no pref, same CD, no alternation point.
	// First-team = t1 (smaller TPN). Last played = W → alternate to B.
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1, Score: 1, ColorHistory: []lexswiss.Color{W}}
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2, Score: 1, ColorHistory: []lexswiss.Color{W}}

	wID, bID := AllocateColor(a, b, ColorPrefTypeNone, false, nil, nil)
	// No prefs. CD both = +1, same. No alternation point (both had W in round 1).
	// Grant first-team pref: first-team = t1, no pref → skip.
	// Art. 4.3.8: alternate first-team from last. t1 last = W → t1 gets B.
	if wID != "t2" || bID != "t1" {
		t.Errorf("first-team alternate: expected t2=W, t1=B; got w=%s, b=%s", wID, bID)
	}
}

func TestAllocateColor_SecondaryScore(t *testing.T) {
	// Art. 4.2.2: first-team uses secondary score as tiebreaker.
	// Both have same primary score, but t2 has higher secondary score.
	a := &lexswiss.ParticipantState{ID: "t1", TPN: 1, Score: 2, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite}}
	b := &lexswiss.ParticipantState{ID: "t2", TPN: 2, Score: 2, ColorHistory: []lexswiss.Color{lexswiss.ColorWhite}}
	secondaryScores := map[string]float64{"t1": 4.0, "t2": 6.0}

	wID, bID := AllocateColor(a, b, ColorPrefTypeNone, false, nil, secondaryScores)
	// First-team = t2 (higher secondary score). Both CD=+1.
	// Art. 4.3.8: alternate first-team (t2) from last (W) → t2 gets B.
	if wID != "t1" || bID != "t2" {
		t.Errorf("secondary score: expected t1=W, t2=B; got w=%s, b=%s", wID, bID)
	}
}
