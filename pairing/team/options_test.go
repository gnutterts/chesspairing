package team

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPairerImplementsInterface(t *testing.T) {
	var _ chesspairing.Pairer = (*Pairer)(nil)
}

func TestNewWithDefaults(t *testing.T) {
	p := New(Options{})
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.opts.ColorPreferenceType == nil || *p.opts.ColorPreferenceType != "A" {
		t.Errorf("expected ColorPreferenceType default 'A', got %v", p.opts.ColorPreferenceType)
	}
	if p.opts.PrimaryScore == nil || *p.opts.PrimaryScore != "match" {
		t.Errorf("expected PrimaryScore default 'match', got %v", p.opts.PrimaryScore)
	}
}

func TestNewFromMap(t *testing.T) {
	p := NewFromMap(map[string]any{
		"topSeedColor":        "white",
		"totalRounds":         float64(9),
		"colorPreferenceType": "B",
		"primaryScore":        "game",
		"forbiddenPairs":      []any{[]any{"t1", "t2"}},
	})
	if *p.opts.TopSeedColor != "white" {
		t.Errorf("expected TopSeedColor 'white', got %q", *p.opts.TopSeedColor)
	}
	if p.opts.TotalRounds == nil || *p.opts.TotalRounds != 9 {
		t.Errorf("expected TotalRounds 9, got %v", p.opts.TotalRounds)
	}
	if *p.opts.ColorPreferenceType != "B" {
		t.Errorf("expected ColorPreferenceType 'B', got %q", *p.opts.ColorPreferenceType)
	}
	if *p.opts.PrimaryScore != "game" {
		t.Errorf("expected PrimaryScore 'game', got %q", *p.opts.PrimaryScore)
	}
	if len(p.opts.ForbiddenPairs) != 1 {
		t.Errorf("expected 1 forbidden pair, got %d", len(p.opts.ForbiddenPairs))
	}
}

func TestNewFromMap_ColorPrefNone(t *testing.T) {
	p := NewFromMap(map[string]any{
		"colorPreferenceType": "none",
	})
	if *p.opts.ColorPreferenceType != "none" {
		t.Errorf("expected ColorPreferenceType 'none', got %q", *p.opts.ColorPreferenceType)
	}
}

func TestParseOptionsNilMap(t *testing.T) {
	o := ParseOptions(nil)
	if o.TopSeedColor != nil {
		t.Error("nil map should produce nil TopSeedColor")
	}
}

func TestOptionsWithDefaults(t *testing.T) {
	var o Options
	d := o.WithDefaults()
	if d.ColorPreferenceType == nil || *d.ColorPreferenceType != "A" {
		t.Error("default ColorPreferenceType should be 'A'")
	}
	if d.PrimaryScore == nil || *d.PrimaryScore != "match" {
		t.Error("default PrimaryScore should be 'match'")
	}
}

func TestResolveColorPrefType(t *testing.T) {
	tests := []struct {
		input string
		want  ColorPrefType
	}{
		{"A", ColorPrefTypeA},
		{"a", ColorPrefTypeA},
		{"B", ColorPrefTypeB},
		{"b", ColorPrefTypeB},
		{"none", ColorPrefTypeNone},
		{"None", ColorPrefTypeNone},
		{"NONE", ColorPrefTypeNone},
		{"", ColorPrefTypeA},
		{"invalid", ColorPrefTypeA},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := resolveColorPrefType(tt.input)
			if got != tt.want {
				t.Errorf("resolveColorPrefType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
