// cmd/chesspairing/tiebreakers_test.go
package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunTiebreakers_Text(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runTiebreakers(nil, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("got exit %d, want %d", code, ExitSuccess)
	}
	out := stdout.String()
	// Should have at least one tiebreaker listed
	if !strings.Contains(out, "buchholz") {
		t.Errorf("should contain buchholz, got: %s", out)
	}
	// Should have multiple lines
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 10 {
		t.Errorf("expected at least 10 tiebreakers, got %d lines", len(lines))
	}
}

func TestRunTiebreakers_JSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runTiebreakers([]string{"--json"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("got exit %d, want %d", code, ExitSuccess)
	}
	var result []map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, stdout.String())
	}
	if len(result) < 10 {
		t.Errorf("expected at least 10 tiebreakers, got %d", len(result))
	}
	// Each entry should have "id" and "name"
	for i, entry := range result {
		if entry["id"] == "" {
			t.Errorf("tiebreaker[%d] missing id", i)
		}
		if entry["name"] == "" {
			t.Errorf("tiebreaker[%d] missing name", i)
		}
	}
}

func TestRunTiebreakers_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runTiebreakers([]string{"--help"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("help: got exit %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "tiebreakers") {
		t.Errorf("help should describe tiebreakers command")
	}
}
