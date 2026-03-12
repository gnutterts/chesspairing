package chesspairing_test

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPairingSystem_IsValid(t *testing.T) {
	valid := []chesspairing.PairingSystem{
		chesspairing.PairingDutch,
		chesspairing.PairingBurstein,
		chesspairing.PairingKeizer,
		chesspairing.PairingRoundRobin,
	}
	for _, ps := range valid {
		if !ps.IsValid() {
			t.Errorf("%q.IsValid() = false, want true", ps)
		}
	}

	invalid := []chesspairing.PairingSystem{"swiss", "unknown", ""}
	for _, ps := range invalid {
		if ps.IsValid() {
			t.Errorf("%q.IsValid() = true, want false", ps)
		}
	}
}

func TestDefaultTiebreakers_Dutch(t *testing.T) {
	tb := chesspairing.DefaultTiebreakers(chesspairing.PairingDutch)
	if len(tb) != 4 || tb[0] != "buchholz-cut1" {
		t.Errorf("DefaultTiebreakers(PairingDutch) = %v, want buchholz-cut1 first", tb)
	}
}

func TestDefaultTiebreakers_Burstein(t *testing.T) {
	tb := chesspairing.DefaultTiebreakers(chesspairing.PairingBurstein)
	if len(tb) != 4 || tb[0] != "buchholz-cut1" {
		t.Errorf("DefaultTiebreakers(PairingBurstein) = %v, want buchholz-cut1 first", tb)
	}
}

func TestScoringSystem_IsValid(t *testing.T) {
	valid := []chesspairing.ScoringSystem{
		chesspairing.ScoringStandard,
		chesspairing.ScoringKeizer,
		chesspairing.ScoringFootball,
	}
	for _, ss := range valid {
		if !ss.IsValid() {
			t.Errorf("%q.IsValid() = false, want true", ss)
		}
	}

	if chesspairing.ScoringSystem("unknown").IsValid() {
		t.Error("unknown.IsValid() = true, want false")
	}
}
