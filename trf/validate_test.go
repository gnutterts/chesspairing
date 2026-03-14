package trf

import (
	"testing"
)

func TestValidate_general_empty(t *testing.T) {
	doc := &Document{}

	issues := doc.Validate(ValidateGeneral)

	var hasNoPlayers bool
	for _, iss := range issues {
		if iss.Field == "001" && iss.Severity == SeverityError {
			hasNoPlayers = true
		}
	}
	if !hasNoPlayers {
		t.Error("expected error for missing players")
	}
}

func TestValidate_general_duplicateStartNumbers(t *testing.T) {
	doc := &Document{
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice"},
			{StartNumber: 1, Name: "Bob"},
		},
	}

	issues := doc.Validate(ValidateGeneral)

	var hasDup bool
	for _, iss := range issues {
		if iss.Field == "001.startNumber" && iss.Severity == SeverityError {
			hasDup = true
		}
	}
	if !hasDup {
		t.Error("expected error for duplicate start numbers")
	}
}

func TestValidate_general_numPlayersMismatch(t *testing.T) {
	doc := &Document{
		NumPlayers: 5,
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice"},
			{StartNumber: 2, Name: "Bob"},
		},
	}

	issues := doc.Validate(ValidateGeneral)

	var hasMismatch bool
	for _, iss := range issues {
		if iss.Field == "062" && iss.Severity == SeverityWarning {
			hasMismatch = true
		}
	}
	if !hasMismatch {
		t.Error("expected warning for NumPlayers mismatch")
	}
}

func TestValidate_general_valid(t *testing.T) {
	doc := &Document{
		NumPlayers: 2,
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice"},
			{StartNumber: 2, Name: "Bob"},
		},
	}

	issues := doc.Validate(ValidateGeneral)

	for _, iss := range issues {
		if iss.Severity == SeverityError {
			t.Errorf("unexpected error: %s: %s", iss.Field, iss.Message)
		}
	}
}

func TestValidate_pairingEngine(t *testing.T) {
	doc := &Document{
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice"},
		},
	}

	issues := doc.Validate(ValidatePairingEngine)

	fields := make(map[string]bool)
	for _, iss := range issues {
		if iss.Severity == SeverityError {
			fields[iss.Field] = true
		}
	}

	for _, want := range []string{"XXR", "XXC", "092"} {
		if !fields[want] {
			t.Errorf("expected error for missing %s", want)
		}
	}
}

func TestValidate_pairingEngine_valid(t *testing.T) {
	doc := &Document{
		TotalRounds:    7,
		InitialColor:   "white1",
		TournamentType: "Swiss Dutch",
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice"},
		},
	}

	issues := doc.Validate(ValidatePairingEngine)

	for _, iss := range issues {
		if iss.Severity == SeverityError {
			t.Errorf("unexpected error: %s: %s", iss.Field, iss.Message)
		}
	}
}

func TestValidate_fide_missingHeaders(t *testing.T) {
	doc := &Document{
		TotalRounds:    7,
		InitialColor:   "white1",
		TournamentType: "Swiss Dutch",
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice", Rating: 2000, Federation: "NED", FideID: "12345"},
		},
	}

	issues := doc.Validate(ValidateFIDE)

	fields := make(map[string]bool)
	for _, iss := range issues {
		if iss.Severity == SeverityError {
			fields[iss.Field] = true
		}
	}

	for _, want := range []string{"012", "042", "052", "122"} {
		if !fields[want] {
			t.Errorf("expected error for missing %s", want)
		}
	}
}

func TestValidate_fide_playerWarnings(t *testing.T) {
	doc := &Document{
		Name:           "Test",
		StartDate:      "2026/01/01",
		EndDate:        "2026/01/07",
		TimeControl:    "90/40+30+30",
		TotalRounds:    7,
		InitialColor:   "white1",
		TournamentType: "Swiss Dutch",
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice", Rating: 2000, Federation: "NED", FideID: "12345"},
			{StartNumber: 2, Name: "", Rating: 1800, Federation: "", FideID: ""},      // missing name, federation, fide ID (rated)
			{StartNumber: 3, Name: "Carol", Rating: 0, Federation: "NED", FideID: ""}, // unrated, no FIDE ID — no warning
		},
	}

	issues := doc.Validate(ValidateFIDE)

	// Should have no errors.
	for _, iss := range issues {
		if iss.Severity == SeverityError {
			t.Errorf("unexpected error: %s: %s", iss.Field, iss.Message)
		}
	}

	// Should warn about player 2: missing name, missing federation, rated but no FIDE ID.
	warnings := make(map[string]bool)
	for _, iss := range issues {
		if iss.Severity == SeverityWarning {
			warnings[iss.Field] = true
		}
	}

	if !warnings["player.2.name"] {
		t.Error("expected warning for player 2 missing name")
	}
	if !warnings["player.2.federation"] {
		t.Error("expected warning for player 2 missing federation")
	}
	if !warnings["player.2.fideID"] {
		t.Error("expected warning for player 2 rated but no FIDE ID")
	}

	// Player 3 is unrated — no FIDE ID warning.
	if warnings["player.3.fideID"] {
		t.Error("unexpected warning for unrated player 3 missing FIDE ID")
	}
}

func TestValidate_fide_fullyValid(t *testing.T) {
	doc := &Document{
		Name:           "FIDE Open 2026",
		StartDate:      "2026/01/01",
		EndDate:        "2026/01/07",
		TimeControl:    "90/40+30+30",
		TotalRounds:    7,
		InitialColor:   "white1",
		TournamentType: "Swiss Dutch",
		NumPlayers:     1,
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice", Rating: 2000, Federation: "NED", FideID: "12345"},
		},
	}

	issues := doc.Validate(ValidateFIDE)

	if len(issues) != 0 {
		for _, iss := range issues {
			t.Errorf("%s [%v]: %s", iss.Field, iss.Severity, iss.Message)
		}
	}
}
