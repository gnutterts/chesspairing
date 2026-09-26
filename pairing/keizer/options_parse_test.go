package keizer

import "testing"

func TestParseOptionsPrefersNestedScoringOptions(t *testing.T) {
	nested := ParseOptions(map[string]any{
		"selfVictory":    true,
		"scoringOptions": map[string]any{"selfVictory": false},
	})
	flat := ParseOptions(map[string]any{"selfVictory": false})

	if nested.ScoringOptions == nil || nested.ScoringOptions.SelfVictory == nil || *nested.ScoringOptions.SelfVictory {
		t.Fatalf("nested selfVictory = %#v, want false", nested.ScoringOptions)
	}
	if flat.ScoringOptions == nil || flat.ScoringOptions.SelfVictory == nil || *flat.ScoringOptions.SelfVictory {
		t.Fatalf("flat selfVictory = %#v, want false", flat.ScoringOptions)
	}
}

func TestParseOptionsCombinesFlatAndNestedScoringOptions(t *testing.T) {
	m := map[string]any{
		"frozen":         true,
		"scoringOptions": map[string]any{"absenceLimit": 3},
	}
	opts := ParseOptions(m)
	assertCombinedScoringOptions(t, opts)

	strict, err := ParseOptionsStrict(m)
	if err != nil {
		t.Fatalf("ParseOptionsStrict: %v", err)
	}
	assertCombinedScoringOptions(t, strict)
}

func assertCombinedScoringOptions(t *testing.T, opts Options) {
	t.Helper()
	if opts.ScoringOptions == nil {
		t.Fatal("ScoringOptions = nil, want combined flat and nested values")
	}
	if opts.ScoringOptions.Frozen == nil || !*opts.ScoringOptions.Frozen {
		t.Fatalf("flat frozen = %v, want true", opts.ScoringOptions.Frozen)
	}
	if opts.ScoringOptions.AbsenceLimit == nil || *opts.ScoringOptions.AbsenceLimit != 3 {
		t.Fatalf("nested absenceLimit = %v, want 3", opts.ScoringOptions.AbsenceLimit)
	}
}

func TestParseOptionsStrictRejectsUnknownNestedScoringOption(t *testing.T) {
	_, err := ParseOptionsStrict(map[string]any{
		"scoringOptions": map[string]any{"unknownOption": true},
	})
	if err == nil {
		t.Fatal("ParseOptionsStrict accepted an unknown nested scoring option")
	}
}

func TestParseOptionsIgnoresUnknownOptions(t *testing.T) {
	opts := ParseOptions(map[string]any{"selfVictory": false, "unknownOption": true})
	if opts.ScoringOptions == nil || opts.ScoringOptions.SelfVictory == nil || *opts.ScoringOptions.SelfVictory {
		t.Fatalf("selfVictory = %#v, want false", opts.ScoringOptions)
	}
}

func TestParseOptionsStrictRejectsInvalidInitialOrder(t *testing.T) {
	_, err := ParseOptionsStrict(map[string]any{"initialOrder": "rating-id"})
	if err == nil || err.Error() != `invalid Keizer initialOrder "rating-id"` {
		t.Fatalf("error = %v, want invalid initialOrder error", err)
	}
}
