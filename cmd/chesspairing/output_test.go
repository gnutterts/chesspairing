// cmd/chesspairing/output_test.go
package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/trf"
)

func TestFormatPairList(t *testing.T) {
	result := &cp.PairingResult{
		Pairings: []cp.GamePairing{
			{Board: 1, WhiteID: "1", BlackID: "4"},
			{Board: 2, WhiteID: "2", BlackID: "3"},
		},
		Byes: []cp.ByeEntry{
			{PlayerID: "5", Type: cp.ByePAB},
		},
	}
	// Build a minimal player ID→start number map
	playerNumbers := map[string]int{"1": 1, "2": 2, "3": 3, "4": 4, "5": 5}

	var buf bytes.Buffer
	formatPairList(&buf, result, playerNumbers)
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	// First line: number of pairings (not including byes)
	if lines[0] != "2" {
		t.Errorf("first line: got %q, want %q", lines[0], "2")
	}
	// Pairing lines: "white black"
	if lines[1] != "1 4" {
		t.Errorf("pairing 1: got %q, want %q", lines[1], "1 4")
	}
	if lines[2] != "2 3" {
		t.Errorf("pairing 2: got %q, want %q", lines[2], "2 3")
	}
	// Bye line: "player 0"
	if lines[3] != "5 0" {
		t.Errorf("bye: got %q, want %q", lines[3], "5 0")
	}
}

func TestFormatStandingsText(t *testing.T) {
	standings := []cp.Standing{
		{Rank: 1, PlayerID: "1", DisplayName: "Fischer, Robert", Score: 6.5,
			TieBreakers: []cp.NamedValue{{ID: "buchholz", Name: "Buchholz", Value: 32.0}},
			GamesPlayed: 9, Wins: 6, Draws: 1, Losses: 2},
		{Rank: 2, PlayerID: "2", DisplayName: "Karpov, Anatoly", Score: 6.0,
			TieBreakers: []cp.NamedValue{{ID: "buchholz", Name: "Buchholz", Value: 30.0}},
			GamesPlayed: 9, Wins: 5, Draws: 2, Losses: 2},
	}
	var buf bytes.Buffer
	formatStandingsText(&buf, standings)
	out := buf.String()
	if !strings.Contains(out, "Fischer") {
		t.Errorf("should contain Fischer, got: %s", out)
	}
	if !strings.Contains(out, "6.5") {
		t.Errorf("should contain 6.5, got: %s", out)
	}
}

func TestFormatStandingsJSON(t *testing.T) {
	standings := []cp.Standing{
		{Rank: 1, PlayerID: "1", DisplayName: "Fischer, Robert", Score: 6.5,
			TieBreakers: []cp.NamedValue{{ID: "buchholz", Name: "Buchholz", Value: 32.0}},
			GamesPlayed: 9, Wins: 6, Draws: 1, Losses: 2},
	}
	var buf bytes.Buffer
	formatStandingsJSON(&buf, standings, "standard", []string{"buchholz"})
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	stArr, ok := result["standings"].([]any)
	if !ok || len(stArr) != 1 {
		t.Fatalf("expected 1 standing, got %v", result["standings"])
	}
}

func TestFormatValidationText(t *testing.T) {
	issues := []trf.ValidationIssue{
		{Field: "XXR", Severity: trf.SeverityError, Message: "missing total rounds"},
		{Field: "player.2.rating", Severity: trf.SeverityWarning, Message: "no rating"},
	}
	var buf bytes.Buffer
	formatValidationText(&buf, "test.trf", issues)
	out := buf.String()
	if !strings.Contains(out, "1 error") {
		t.Errorf("should report 1 error, got: %s", out)
	}
	if !strings.Contains(out, "1 warning") {
		t.Errorf("should report 1 warning, got: %s", out)
	}
}

func TestFormatValidationJSON(t *testing.T) {
	issues := []trf.ValidationIssue{
		{Field: "XXR", Severity: trf.SeverityError, Message: "missing total rounds"},
	}
	var buf bytes.Buffer
	formatValidationJSON(&buf, issues, "standard", "trfx")
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["valid"] != false {
		t.Errorf("expected valid=false")
	}
}
