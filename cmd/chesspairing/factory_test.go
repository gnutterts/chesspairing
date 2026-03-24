// cmd/chesspairing/factory_test.go
package main

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestNewPairer_AllSystems(t *testing.T) {
	systems := []cp.PairingSystem{
		cp.PairingDutch,
		cp.PairingBurstein,
		cp.PairingDubov,
		cp.PairingLim,
		cp.PairingDoubleSwiss,
		cp.PairingTeam,
		cp.PairingKeizer,
		cp.PairingRoundRobin,
	}
	for _, sys := range systems {
		t.Run(string(sys), func(t *testing.T) {
			p, err := newPairer(sys, nil)
			if err != nil {
				t.Fatalf("newPairer(%q): %v", sys, err)
			}
			if p == nil {
				t.Fatalf("newPairer(%q) returned nil", sys)
			}
		})
	}
}

func TestNewPairer_UnknownSystem(t *testing.T) {
	_, err := newPairer("bogus", nil)
	if err == nil {
		t.Fatal("expected error for unknown system")
	}
}

func TestNewScorer_AllSystems(t *testing.T) {
	systems := []cp.ScoringSystem{
		cp.ScoringStandard,
		cp.ScoringKeizer,
		cp.ScoringFootball,
	}
	for _, sys := range systems {
		t.Run(string(sys), func(t *testing.T) {
			s, err := newScorer(sys, nil)
			if err != nil {
				t.Fatalf("newScorer(%q): %v", sys, err)
			}
			if s == nil {
				t.Fatalf("newScorer(%q) returned nil", sys)
			}
		})
	}
}

func TestNewScorer_UnknownSystem(t *testing.T) {
	_, err := newScorer("bogus", nil)
	if err == nil {
		t.Fatal("expected error for unknown system")
	}
}

func TestNewScorer_CustomPoints(t *testing.T) {
	opts := map[string]any{
		"pointWin":  3.0,
		"pointDraw": 1.0,
		"pointLoss": 0.0,
	}
	s, err := newScorer(cp.ScoringStandard, opts)
	if err != nil {
		t.Fatalf("newScorer with custom points: %v", err)
	}
	if s == nil {
		t.Fatal("newScorer returned nil")
	}
}

func TestParseSystemFlag_Valid(t *testing.T) {
	tests := []struct {
		flag string
		want cp.PairingSystem
	}{
		{"--dutch", cp.PairingDutch},
		{"--burstein", cp.PairingBurstein},
		{"--dubov", cp.PairingDubov},
		{"--lim", cp.PairingLim},
		{"--double-swiss", cp.PairingDoubleSwiss},
		{"--team", cp.PairingTeam},
		{"--keizer", cp.PairingKeizer},
		{"--roundrobin", cp.PairingRoundRobin},
	}
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			got, ok := parseSystemFlag(tt.flag)
			if !ok {
				t.Fatalf("parseSystemFlag(%q) returned false", tt.flag)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSystemFlag_Invalid(t *testing.T) {
	_, ok := parseSystemFlag("--foobar")
	if ok {
		t.Fatal("expected false for unknown flag")
	}
	_, ok = parseSystemFlag("dutch")
	if ok {
		t.Fatal("expected false for flag without --")
	}
}
