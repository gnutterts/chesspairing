package lim

import (
	"testing"

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

func TestSelectDownFloater_LowestPairingNumber(t *testing.T) {
	// Art. 3.2.4: if equal due colours, lowest TPN floats down.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 2.0, ColorHistory: nil},
		{ID: "b", TPN: 2, Score: 2.0, ColorHistory: nil},
		{ID: "c", TPN: 3, Score: 2.0, ColorHistory: nil},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	floater := SelectDownFloater(players, adjacent, nil)
	if floater == nil || floater.ID != "a" {
		t.Errorf("expected 'a' (lowest TPN), got %v", floater)
	}
}

func TestSelectUpFloater_HighestPairingNumber(t *testing.T) {
	// Art. 3.2.4: highest TPN floats up.
	players := []*swisslib.PlayerState{
		{ID: "a", TPN: 1, Score: 0.0, ColorHistory: nil},
		{ID: "b", TPN: 2, Score: 0.0, ColorHistory: nil},
		{ID: "c", TPN: 3, Score: 0.0, ColorHistory: nil},
	}
	adjacent := []*swisslib.PlayerState{
		{ID: "x", TPN: 10, Score: 1.0},
	}

	floater := SelectUpFloater(players, adjacent, nil)
	if floater == nil || floater.ID != "c" {
		t.Errorf("expected 'c' (highest TPN), got %v", floater)
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

	floater := SelectDownFloater(players, adjacent, nil)
	if floater == nil {
		t.Fatal("expected a floater")
	}
	// Floater should be due Black (to equalise)
	pref := swisslib.ComputeColorPreference(floater.ColorHistory)
	if pref.Color == nil || *pref.Color != B {
		t.Errorf("expected floater due Black for colour balance, got %v", floater.ID)
	}
}
