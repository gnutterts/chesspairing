package lim

import (
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

func TestPairerImplementsInterface(t *testing.T) {
	var _ chesspairing.Pairer = (*Pairer)(nil)
}

func TestNewWithDefaults(t *testing.T) {
	p := New(Options{})
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if *p.opts.TopSeedColor != "auto" {
		t.Errorf("expected TopSeedColor default 'auto', got %q", *p.opts.TopSeedColor)
	}
}

func TestNewFromMap(t *testing.T) {
	p := NewFromMap(map[string]any{
		"topSeedColor":   "white",
		"maxiTournament": true,
		"forbiddenPairs": []any{[]any{"p1", "p2"}},
	})
	if *p.opts.TopSeedColor != "white" {
		t.Errorf("expected TopSeedColor 'white', got %q", *p.opts.TopSeedColor)
	}
	if p.opts.MaxiTournament == nil || !*p.opts.MaxiTournament {
		t.Error("expected MaxiTournament true")
	}
	if len(p.opts.ForbiddenPairs) != 1 {
		t.Errorf("expected 1 forbidden pair, got %d", len(p.opts.ForbiddenPairs))
	}
}

func TestParseOptionsNilMap(t *testing.T) {
	o := ParseOptions(nil)
	if o.TopSeedColor != nil {
		t.Error("nil map should produce nil TopSeedColor")
	}
}
