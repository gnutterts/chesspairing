package dubov

import (
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

func TestOptionsWithDefaults(t *testing.T) {
	var o Options
	d := o.WithDefaults()
	if d.TopSeedColor == nil || *d.TopSeedColor != "auto" {
		t.Error("default TopSeedColor should be auto")
	}
}

func TestParseOptions(t *testing.T) {
	m := map[string]any{
		"topSeedColor": "black",
		"totalRounds":  float64(9),
		"forbiddenPairs": []any{
			[]any{"p1", "p2"},
		},
	}
	o := ParseOptions(m)
	if o.TopSeedColor == nil || *o.TopSeedColor != "black" {
		t.Errorf("expected black, got %v", o.TopSeedColor)
	}
	if o.TotalRounds == nil || *o.TotalRounds != 9 {
		t.Errorf("expected 9, got %v", o.TotalRounds)
	}
	if len(o.ForbiddenPairs) != 1 {
		t.Errorf("expected 1 forbidden pair, got %d", len(o.ForbiddenPairs))
	}
}

func TestParseOptionsNil(t *testing.T) {
	o := ParseOptions(nil)
	if o.TopSeedColor != nil {
		t.Error("nil map should produce nil TopSeedColor")
	}
}

func TestNewFromMap(t *testing.T) {
	p := NewFromMap(map[string]any{"topSeedColor": "white"})
	if p == nil {
		t.Fatal("NewFromMap returned nil")
	}
}

func TestPairerInterface(t *testing.T) {
	var _ chesspairing.Pairer = (*Pairer)(nil)
}
