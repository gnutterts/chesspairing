package burstein

import (
	"fmt"
	"testing"
)

func TestParseOptions_Empty(t *testing.T) {
	t.Parallel()

	opts := ParseOptions(nil)
	if opts.TopSeedColor != nil {
		t.Errorf("expected nil TopSeedColor, got %v", *opts.TopSeedColor)
	}
	if opts.TotalRounds != nil {
		t.Errorf("expected nil TotalRounds, got %v", *opts.TotalRounds)
	}
	if opts.ForbiddenPairs != nil {
		t.Errorf("expected nil ForbiddenPairs, got %v", opts.ForbiddenPairs)
	}
}

func TestParseOptions_WithValues(t *testing.T) {
	t.Parallel()

	m := map[string]any{
		"topSeedColor": "white",
		"totalRounds":  float64(9),
		"forbiddenPairs": []any{
			[]any{"p1", "p2"},
		},
	}

	opts := ParseOptions(m)

	if opts.TopSeedColor == nil || *opts.TopSeedColor != "white" {
		t.Errorf("expected topSeedColor=white, got %v", opts.TopSeedColor)
	}
	if opts.TotalRounds == nil || *opts.TotalRounds != 9 {
		t.Errorf("expected totalRounds=9, got %v", opts.TotalRounds)
	}
	if len(opts.ForbiddenPairs) != 1 || opts.ForbiddenPairs[0][0] != "p1" || opts.ForbiddenPairs[0][1] != "p2" {
		t.Errorf("expected forbiddenPairs=[[p1,p2]], got %v", opts.ForbiddenPairs)
	}
}

func TestWithDefaults(t *testing.T) {
	t.Parallel()

	opts := Options{}
	d := opts.WithDefaults()

	if d.TopSeedColor == nil || *d.TopSeedColor != "auto" {
		t.Errorf("expected default topSeedColor=auto, got %v", d.TopSeedColor)
	}
}

func TestSeedingRounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		totalRounds int
		want        int
	}{
		{name: "0 rounds", totalRounds: 0, want: 0},
		{name: "1 round", totalRounds: 1, want: 0},
		{name: "2 rounds", totalRounds: 2, want: 1},
		{name: "3 rounds", totalRounds: 3, want: 1},
		{name: "4 rounds", totalRounds: 4, want: 2},
		{name: "5 rounds", totalRounds: 5, want: 2},
		{name: "6 rounds", totalRounds: 6, want: 3},
		{name: "7 rounds", totalRounds: 7, want: 3},
		{name: "8 rounds", totalRounds: 8, want: 4},
		{name: "9 rounds", totalRounds: 9, want: 4},
		{name: "10 rounds", totalRounds: 10, want: 4},
		{name: "11 rounds", totalRounds: 11, want: 4},
		{name: "20 rounds", totalRounds: 20, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SeedingRounds(tt.totalRounds)
			if got != tt.want {
				t.Errorf("SeedingRounds(%d) = %d, want %d", tt.totalRounds, got, tt.want)
			}
		})
	}
}

func TestIsSeedingRound(t *testing.T) {
	t.Parallel()

	// 9 rounds → 4 seeding rounds.
	tests := []struct {
		round int
		want  bool
	}{
		{1, true},
		{2, true},
		{3, true},
		{4, true},
		{5, false},
		{6, false},
		{9, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("round_%d", tt.round), func(t *testing.T) {
			t.Parallel()
			got := IsSeedingRound(tt.round, 9)
			if got != tt.want {
				t.Errorf("IsSeedingRound(%d, 9) = %v, want %v", tt.round, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	p := New(Options{})
	if p == nil {
		t.Fatal("expected non-nil Pairer")
	}
	if p.opts.TopSeedColor == nil || *p.opts.TopSeedColor != "auto" {
		t.Error("expected defaults applied")
	}
}

func TestNewFromMap(t *testing.T) {
	t.Parallel()

	p := NewFromMap(map[string]any{
		"topSeedColor": "black",
	})
	if p == nil {
		t.Fatal("expected non-nil Pairer")
	}
	if p.opts.TopSeedColor == nil || *p.opts.TopSeedColor != "black" {
		t.Error("expected topSeedColor=black")
	}
}
