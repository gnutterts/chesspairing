package lim

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestIsCompatible_NeverPlayed(t *testing.T) {
	a := &swisslib.PlayerState{ID: "a", ColorHistory: nil}
	b := &swisslib.PlayerState{ID: "b", ColorHistory: nil}
	if !IsCompatible(a, b, nil) {
		t.Error("two players with no history should be compatible")
	}
}

func TestIsCompatible_AlreadyPlayed(t *testing.T) {
	a := &swisslib.PlayerState{ID: "a", Opponents: []string{"b"}}
	b := &swisslib.PlayerState{ID: "b", Opponents: []string{"a"}}
	if IsCompatible(a, b, nil) {
		t.Error("players who already played should NOT be compatible")
	}
}

func TestIsCompatible_ThreeConsecutiveSameColor(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Player a has WW history, b has BB history.
	// Assignment a=B, b=W is valid, so they ARE compatible.
	a := &swisslib.PlayerState{ID: "a", ColorHistory: []swisslib.Color{W, W}}
	b := &swisslib.PlayerState{ID: "b", ColorHistory: []swisslib.Color{B, B}}
	if !IsCompatible(a, b, nil) {
		t.Error("a=WW, b=BB: should be compatible (a gets B, b gets W)")
	}
}

func TestIsCompatible_BothNeedSameColorToAvoidThreeConsecutive(t *testing.T) {
	W := swisslib.ColorWhite
	// Both players have WW — both MUST get B to avoid 3 consecutive.
	// Only one can get B, so they're incompatible.
	a := &swisslib.PlayerState{ID: "a", ColorHistory: []swisslib.Color{W, W}}
	b := &swisslib.PlayerState{ID: "b", ColorHistory: []swisslib.Color{W, W}}
	if IsCompatible(a, b, nil) {
		t.Error("a=WW, b=WW: both need Black, should be incompatible")
	}
}

func TestIsCompatible_ThreeMoreOfOneColor(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Player a has WWWB (3W, 1B, diff=2). If a gets W, diff becomes 3.
	// Player b has BBBW (3B, 1W, diff=2). If b gets B, diff becomes 3.
	// Assignment a=B,b=W: both OK.
	a := &swisslib.PlayerState{ID: "a", ColorHistory: []swisslib.Color{W, W, W, B}}
	b := &swisslib.PlayerState{ID: "b", ColorHistory: []swisslib.Color{B, B, B, W}}
	if !IsCompatible(a, b, nil) {
		t.Error("should be compatible via a=B, b=W")
	}
}

func TestIsCompatible_BothExceedImbalance(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack
	// Both a and b have WWWB (diff=2). Both MUST get B (can't get W -> diff=3).
	// Only one can get B -> incompatible.
	a := &swisslib.PlayerState{ID: "a", ColorHistory: []swisslib.Color{W, W, W, B}}
	b := &swisslib.PlayerState{ID: "b", ColorHistory: []swisslib.Color{W, W, W, B}}
	if IsCompatible(a, b, nil) {
		t.Error("both need B to avoid 3+ imbalance, should be incompatible")
	}
}

func TestIsCompatible_ForbiddenPair(t *testing.T) {
	a := &swisslib.PlayerState{ID: "a"}
	b := &swisslib.PlayerState{ID: "b"}
	forbidden := map[[2]string]bool{{"a", "b"}: true}
	if IsCompatible(a, b, forbidden) {
		t.Error("forbidden pair should be incompatible")
	}
}

func TestCanReceiveColor(t *testing.T) {
	W := swisslib.ColorWhite
	B := swisslib.ColorBlack

	tests := []struct {
		name    string
		history []swisslib.Color
		color   swisslib.Color
		want    bool
	}{
		{"empty history, W", nil, W, true},
		{"empty history, B", nil, B, true},
		{"WW, W -> 3 consecutive", []swisslib.Color{W, W}, W, false},
		{"WW, B -> OK", []swisslib.Color{W, W}, B, true},
		{"WWWB, W -> 3+ imbalance", []swisslib.Color{W, W, W, B}, W, false},
		{"WWWB, B -> OK", []swisslib.Color{W, W, W, B}, B, true},
		{"WBW, W -> OK (imbalance=2, no 3 consec)", []swisslib.Color{W, B, W}, W, true},
		{"WBWW, W -> 3 consec", []swisslib.Color{W, B, W, W}, W, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &swisslib.PlayerState{ColorHistory: tt.history}
			got := CanReceiveColor(p, tt.color)
			if got != tt.want {
				t.Errorf("CanReceiveColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
