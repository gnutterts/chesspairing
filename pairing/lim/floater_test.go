// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestFloaterType_Order(t *testing.T) {
	if FloaterTypeA >= FloaterTypeB {
		t.Error("TypeA should be worse (lower value) than TypeB")
	}
	if FloaterTypeB >= FloaterTypeC {
		t.Error("TypeB should be worse than TypeC")
	}
	if FloaterTypeC >= FloaterTypeD {
		t.Error("TypeC should be worse than TypeD")
	}
}

func TestClassifyFloater(t *testing.T) {
	// Player who already floated + no compatible opponent in adjacent group.
	p := &swisslib.PlayerState{ID: "p1", Opponents: []string{"adj1"}}
	adjacent := []*swisslib.PlayerState{
		{ID: "adj1", Opponents: []string{"p1"}}, // already played p1
	}

	ft := ClassifyFloater(p, true, adjacent, nil)
	if ft != FloaterTypeA {
		t.Errorf("already floated + no compatible opponent = TypeA, got %v", ft)
	}
}

func TestClassifyFloater_AlreadyFloatedWithCompatible(t *testing.T) {
	p := &swisslib.PlayerState{ID: "p1"}
	adjacent := []*swisslib.PlayerState{
		{ID: "adj1"}, // hasn't played p1
	}

	ft := ClassifyFloater(p, true, adjacent, nil)
	if ft != FloaterTypeB {
		t.Errorf("already floated + has compatible opponent = TypeB, got %v", ft)
	}
}

func TestClassifyFloater_NotFloatedNoCompatible(t *testing.T) {
	p := &swisslib.PlayerState{ID: "p1", Opponents: []string{"adj1"}}
	adjacent := []*swisslib.PlayerState{
		{ID: "adj1", Opponents: []string{"p1"}},
	}

	ft := ClassifyFloater(p, false, adjacent, nil)
	if ft != FloaterTypeC {
		t.Errorf("not floated + no compatible = TypeC, got %v", ft)
	}
}

func TestClassifyFloater_NotFloatedWithCompatible(t *testing.T) {
	p := &swisslib.PlayerState{ID: "p1"}
	adjacent := []*swisslib.PlayerState{
		{ID: "adj1"},
	}

	ft := ClassifyFloater(p, false, adjacent, nil)
	if ft != FloaterTypeD {
		t.Errorf("not floated + has compatible = TypeD, got %v", ft)
	}
}

func TestSelectDownFloater_LowestRanked(t *testing.T) {
	// Art. 3.2.4: if due colours are equal, the lowest-ranked player floats down.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 2.0, ColorHistory: nil},
		{ID: "b", TPN: 2, Score: 2.0, ColorHistory: nil},
		{ID: "c", TPN: 3, Score: 2.0, ColorHistory: nil},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	floater := SelectDownFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "c" {
		t.Errorf("expected 'c' (largest TPN), got %v", floater)
	}
}

func TestSelectUpFloater_HighestRanked(t *testing.T) {
	// Art. 3.2.4: the highest-ranked player floats up.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 0.0, ColorHistory: nil},
		{ID: "b", TPN: 2, Score: 0.0, ColorHistory: nil},
		{ID: "c", TPN: 3, Score: 0.0, ColorHistory: nil},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	floater := SelectUpFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "a" {
		t.Errorf("expected 'a' (smallest TPN), got %v", floater)
	}
}

func TestSelectDownFloater_ColorBalance(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Art. 3.2.2: select floater to equalise due colours.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 2.0, ColorHistory: []swisslib.Color{W}},       // due B
		{ID: "b", TPN: 2, Score: 2.0, ColorHistory: []swisslib.Color{W}},       // due B
		{ID: "c", TPN: 3, Score: 2.0, ColorHistory: []swisslib.Color{B}},       // due W
		{ID: "d", TPN: 4, Score: 2.0, ColorHistory: []swisslib.Color{B}},       // due W
		{ID: "e", TPN: 5, Score: 2.0, ColorHistory: []swisslib.Color{W, B, W}}, // due B
	}
	// 3 due B, 2 due W — float someone due B to balance.
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	floater := SelectDownFloater(players, adjacent, nil, nil, false)
	if floater == nil {
		t.Fatal("expected a floater")
	}
	// Floater should be due Black (to equalise)
	pref := swisslib.ComputeColorPreference(floater.ColorHistory)
	if pref.Color == nil || *pref.Color != B {
		t.Errorf("expected floater due Black for colour balance, got %v", floater.ID)
	}
}

func TestSelectDownFloater_AlreadyFloatedPreferred(t *testing.T) {
	// Art. 3.9: a player that has already floated (multi-hop) should be classified
	// as Type A/B (more disadvantaged) and thus preferred for further floating over
	// a player that has not floated yet, to avoid accumulating disadvantage.
	//
	// Setup: 3 players in a group, all with equal colour balance and no history.
	// Player "a" was already floated from a higher group (tracked in floatedIDs).
	// Player "b" and "c" are native to this group.
	// Adjacent group has compatible opponents for all.
	// Without floatedIDs tracking, all would be TypeD and "a" would win on lowest TPN.
	// With floatedIDs tracking, "a" is TypeB (already floated + compatible),
	// while "b" and "c" are TypeD (not floated + compatible).
	// TypeD > TypeB, so "b" (lowest TPN among TypeD) should be selected, not "a".
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 2.0},
		{ID: "b", TPN: 2, Score: 2.0},
		{ID: "c", TPN: 3, Score: 2.0},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}
	floatedIDs := map[string]bool{"a": true}

	floater := SelectDownFloater(players, adjacent, nil, floatedIDs, false)
	if floater == nil {
		t.Fatal("expected a floater")
	}
	// "a" is TypeB (already floated), "b" and "c" are TypeD (not floated).
	// TypeD is preferred (least disadvantage), so "c" should be selected (largest TPN among TypeD).
	if floater.ID != "c" {
		t.Errorf("expected 'c' (TypeD, largest TPN), got %v — already-floated player should not be preferred", floater.ID)
	}
}

func TestSelectUpFloater_AlreadyFloatedPreferred(t *testing.T) {
	// Same principle as above but for upward floating.
	// "c" was already floated, so it gets TypeB. "a" and "b" are TypeD.
	// Among TypeD, highest TPN wins for upward floating → "b" selected.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 0.0},
		{ID: "b", TPN: 2, Score: 0.0},
		{ID: "c", TPN: 3, Score: 0.0},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}
	floatedIDs := map[string]bool{"c": true}

	floater := SelectUpFloater(players, adjacent, nil, floatedIDs, false)
	if floater == nil {
		t.Fatal("expected a floater")
	}
	if floater.ID != "a" {
		t.Errorf("expected 'a' (TypeD, smallest TPN among non-floated), got %v", floater.ID)
	}
}

func TestSelectDownFloater_MaxiRatingCap(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2100, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "b", TPN: 2, Rating: 2400, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "c", TPN: 3, Rating: 2400, Score: 2.0, ColorHistory: []swisslib.Color{W}}, // due B
	}
	// 2 due W, 1 due B → majority is W. Candidates: a, b. Largest TPN is b.
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	// Without maxi: "b" selected.
	floater := SelectDownFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "b" {
		t.Fatalf("without maxi: expected 'b', got %v", floater)
	}

	// Art. 3.2.3's "lowest numbered player" is the player with the largest
	// TPN: c (2100). Its rating differs from b by more than 100 points, so c
	// is selected.
	players[2].Rating = 2100
	floater = SelectDownFloater(players, adjacent, nil, nil, true)
	if floater == nil || floater.ID != "c" {
		t.Errorf("with maxi: expected 'c', got %v", floater)
	}
}

func TestSelectUpFloater_MaxiRatingCap(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2400, Score: 0.0, ColorHistory: []swisslib.Color{W}}, // due B
		{ID: "b", TPN: 2, Rating: 2400, Score: 0.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "c", TPN: 3, Rating: 2100, Score: 0.0, ColorHistory: []swisslib.Color{B}}, // due W
	}
	// 2 due W, 1 due B → majority is W. Candidates: b, c. Smallest TPN is b.
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	// Without maxi: "b" selected.
	floater := SelectUpFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "b" {
		t.Fatalf("without maxi: expected 'b', got %v", floater)
	}

	// Art. 3.2.3's "highest numbered player" is the player with the lowest
	// TPN: a. Its rating differs from b by more than 100 points, so a is
	// selected.
	players[0].Rating = 2100
	floater = SelectUpFloater(players, adjacent, nil, nil, true)
	if floater == nil || floater.ID != "a" {
		t.Errorf("with maxi: expected 'a', got %v", floater)
	}
}

func TestSelectDownFloater_MaxiWithinCap(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2400, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "b", TPN: 2, Rating: 2350, Score: 2.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "c", TPN: 3, Rating: 2300, Score: 2.0, ColorHistory: []swisslib.Color{W}}, // due B
	}
	// 2 due W, 1 due B → majority is W. Candidates: a, b. Largest TPN is b.
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	// With maxi: "b" (2350) vs reference "c" (TPN 3, 2300). diff = 50 <= 100 → no override.
	floater := SelectDownFloater(players, adjacent, nil, nil, true)
	if floater == nil || floater.ID != "b" {
		t.Errorf("with maxi within cap: expected 'b' (diff 50 <= 100, no override), got %v", floater)
	}
}

func TestSelectFloater_Art310ProtectsPreviousFloater(t *testing.T) {
	players := []*swisslib.PlayerState{{ID: "a", TPN: 1}, {ID: "b", TPN: 2}, {ID: "c", TPN: 3}}
	adjacent := []*swisslib.PlayerState{{ID: "x"}}
	floater := selectFloater(players, adjacent, nil, nil, false, true, map[string]bool{"c": true}, true, nil)
	if floater == nil || floater.ID != "b" {
		t.Fatalf("Art. 3.10 should protect c with type-D replacement b, got %v", floater)
	}
}

func TestSelectFloater_Art310DoesNotProduceTypeB(t *testing.T) {
	players := []*swisslib.PlayerState{{ID: "a", TPN: 1, Opponents: []string{"x"}}, {ID: "b", TPN: 2}, {ID: "c", TPN: 3}}
	adjacent := []*swisslib.PlayerState{{ID: "x"}}
	floater := selectFloater(players, adjacent, nil, map[string]bool{"b": true}, false, true, map[string]bool{"c": true}, true, nil)
	if floater == nil || floater.ID != "c" {
		t.Fatalf("Art. 3.10.1 must not replace c with type-B b, got %v", floater)
	}
}

func TestSelectFloater_Art310TriesFurtherTypeDReplacement(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1}, {ID: "b", TPN: 4}, {ID: "c", TPN: 5}, {ID: "d", TPN: 2}, {ID: "e", TPN: 3},
	}
	adjacent := []*swisslib.PlayerState{{ID: "x"}}
	forbidden := make(map[[2]string]bool)
	for _, left := range players {
		for _, right := range players {
			if left.ID != right.ID {
				forbidden[[2]string{left.ID, right.ID}] = true
			}
		}
	}
	for _, pair := range [][2]string{{"a", "c"}, {"a", "d"}, {"b", "e"}} {
		delete(forbidden, pair)
		delete(forbidden, [2]string{pair[1], pair[0]})
	}
	floater := selectFloater(players, adjacent, forbidden, nil, false, true, map[string]bool{"c": true}, true, nil)
	if floater == nil || floater.ID != "d" {
		t.Fatalf("Art. 3.10.2 must skip b, which loses a pair, and choose d, got %v", floater)
	}
}

func TestSelectFloater_Art310SkipsPreviousReplacement(t *testing.T) {
	players := []*swisslib.PlayerState{{ID: "a", TPN: 1}, {ID: "previous-1", TPN: 2}, {ID: "previous-2", TPN: 3}}
	floater := selectFloater(players, []*swisslib.PlayerState{{ID: "x"}}, nil, nil, false, true, map[string]bool{"previous-1": true, "previous-2": true}, true, nil)
	if floater == nil || floater.ID != "a" {
		t.Fatalf("Art. 3.10 must skip previous-round floaters as replacements, got %v", floater)
	}
}

func TestSelectFloater_Art310UsesLastRoundCompatibilityForPairCount(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1}, {ID: "b", TPN: 4}, {ID: "previous", TPN: 5}, {ID: "d", TPN: 2}, {ID: "e", TPN: 3},
	}
	forbidden := make(map[[2]string]bool)
	for _, left := range players {
		for _, right := range players {
			if left.ID != right.ID {
				forbidden[[2]string{left.ID, right.ID}] = true
			}
		}
	}
	for _, pair := range [][2]string{{"a", "previous"}, {"a", "d"}, {"b", "e"}} {
		delete(forbidden, pair)
		delete(forbidden, [2]string{pair[1], pair[0]})
	}
	lastRoundCompatible := func(_, _ *swisslib.PlayerState) bool { return true }
	floater := selectFloater(players, []*swisslib.PlayerState{{ID: "x"}}, forbidden, nil, false, true, map[string]bool{"previous": true}, true, lastRoundCompatible)
	if floater == nil || floater.ID != "b" {
		t.Fatalf("Art. 3.10.2 must count pairs with the last-round rule, got %v", floater)
	}
}

func TestSelectFloater_Art323MaxiOverridesPreviousFloaterProtection(t *testing.T) {
	// All candidates are type d: 3.10 first replaces previous with b, then
	// the unconditional Maxi rule in 3.2.3 chooses previous again.
	players := []*swisslib.PlayerState{{ID: "a", TPN: 1, Rating: 2000}, {ID: "b", TPN: 2, Rating: 2200}, {ID: "previous", TPN: 3, Rating: 2400}}
	floater := selectFloater(players, []*swisslib.PlayerState{{ID: "x"}}, nil, nil, true, true, map[string]bool{"previous": true}, true, nil)
	if floater == nil || floater.ID != "previous" {
		t.Fatalf("Art. 3.2.3 must override Art. 3.10, got %v", floater)
	}
}

func TestSelectFloater_Article33ExcludesEarlierFloaterOpponent(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "candidate", TPN: 3, Score: 2, Opponents: []string{"y"}},
		{ID: "replacement", TPN: 2, Score: 2, Opponents: []string{"x"}},
	}
	adjacent := []*swisslib.PlayerState{{ID: "x", TPN: 1}, {ID: "y", TPN: 2}}
	earlier := []FloaterEntry{{Player: &swisslib.PlayerState{ID: "earlier", TPN: 1}, Direction: FloatDown, SourceScore: 2}}
	floater := selectFloaterWithEarlier(players, adjacent, nil, map[string]bool{"replacement": true}, false, true, nil, false, earlier, nil)
	if floater == nil || floater.ID != "replacement" {
		t.Fatalf("Art. 3.3 must exclude x taken by earlier floater, got %v", floater)
	}
}

func TestSelectFloater_Article39CompatibleBeforeType(t *testing.T) {
	players := []*swisslib.PlayerState{{ID: "a", TPN: 1, Opponents: []string{"x"}}, {ID: "b", TPN: 2}, {ID: "c", TPN: 3, Opponents: []string{"x"}}}
	adjacent := []*swisslib.PlayerState{{ID: "x"}}
	floater := SelectDownFloater(players, adjacent, nil, map[string]bool{"b": true}, false)
	if floater == nil || floater.ID != "b" {
		t.Fatalf("Arts. 3.3 and 3.5 require compatible type-B b before type-C c, got %v", floater)
	}
}

func TestPreviousRoundFloaters_ReconstructsLimTransfers(t *testing.T) {
	state := &chesspairing.TournamentState{
		CurrentRound: 3,
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				{WhiteID: "high", BlackID: "h0", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "low", BlackID: "l0", Result: chesspairing.ResultBlackWins},
				{WhiteID: "equal1", BlackID: "e0", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "equal2", BlackID: "e1", Result: chesspairing.ResultWhiteWins},
			}},
			{Number: 2, Games: []chesspairing.GameData{
				{WhiteID: "high", BlackID: "low", Result: chesspairing.ResultDraw},
				{WhiteID: "equal1", BlackID: "equal2", Result: chesspairing.ResultDraw},
			}, Byes: []chesspairing.ByeEntry{{PlayerID: "bye", Type: chesspairing.ByePAB}}},
		},
	}
	floaters := previousRoundFloaters(state)
	if !floaters["high"] || !floaters["low"] || floaters["equal1"] || floaters["equal2"] || floaters["bye"] {
		t.Fatalf("Lim previous-round floaters = %#v, want high and low", floaters)
	}

	residentState := &chesspairing.TournamentState{
		CurrentRound: 4,
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "floater", BlackID: "a", Result: chesspairing.ResultWhiteWins}, {WhiteID: "resident", BlackID: "b", Result: chesspairing.ResultWhiteWins}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "floater", BlackID: "c", Result: chesspairing.ResultDraw}, {WhiteID: "resident", BlackID: "d", Result: chesspairing.ResultBlackWins}}},
			{Number: 3, Games: []chesspairing.GameData{{WhiteID: "floater", BlackID: "resident", Result: chesspairing.ResultDraw}}},
		},
	}
	floaters = previousRoundFloaters(residentState)
	if !floaters["floater"] || floaters["resident"] {
		t.Fatalf("Lim previous-round floaters = %#v, want floater but not resident opponent", floaters)
	}

	// Both players are above the median before round 3. The reconstruction
	// cannot tell which one floated, so its documented limitation marks only
	// the player farther from the median.
	sameSide := &chesspairing.TournamentState{
		CurrentRound: 4,
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "far", BlackID: "x", Result: chesspairing.ResultWhiteWins}, {WhiteID: "near", BlackID: "y", Result: chesspairing.ResultWhiteWins}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "far", BlackID: "x2", Result: chesspairing.ResultWhiteWins}, {WhiteID: "near", BlackID: "y2", Result: chesspairing.ResultDraw}}},
			{Number: 3, Games: []chesspairing.GameData{{WhiteID: "far", BlackID: "near", Result: chesspairing.ResultDraw}}},
		},
	}
	floaters = previousRoundFloaters(sameSide)
	if !floaters["far"] || floaters["near"] {
		t.Fatalf("same-side Lim floaters = %#v, want only farther player", floaters)
	}
}

func TestSelectUpFloater_MaxiRatingCapOverride(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Rating: 2100, Score: 0.0, ColorHistory: []swisslib.Color{B}}, // due W
		{ID: "b", TPN: 2, Rating: 2350, Score: 0.0, ColorHistory: []swisslib.Color{W}}, // due B
		{ID: "c", TPN: 3, Rating: 2400, Score: 0.0, ColorHistory: []swisslib.Color{W}}, // due B
	}
	// 2 due B, 1 due W → majority is B. Candidates: b, c. Smallest TPN is b.
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	// Without maxi: b selected (due B, TPN 2).
	floater := SelectUpFloater(players, adjacent, nil, nil, false)
	if floater == nil || floater.ID != "b" {
		t.Fatalf("without maxi: expected 'b', got %v", floater)
	}

	// With maxi: b (2350) differs by more than 100 points from the reference
	// player a (TPN 1, 2100), so Art. 3.2.3 selects a.
	floater = SelectUpFloater(players, adjacent, nil, nil, true)
	if floater == nil || floater.ID != "a" {
		t.Errorf("with maxi: expected 'a', got %v", floater)
	}
}
