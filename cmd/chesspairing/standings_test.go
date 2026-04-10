// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/standings_test.go
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStandings_Text(t *testing.T) {
	// Need a TRF with completed rounds for standings
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--dutch", input}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("standings text: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Rank") {
		t.Errorf("should contain header, got: %s", out)
	}
}

func TestRunStandings_JSON(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--dutch", input, "--json"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("standings json: exit %d, stderr: %s", code, stderr.String())
	}
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := result["standings"]; !ok {
		t.Error("JSON should contain 'standings' field")
	}
}

func TestRunStandings_CustomTiebreakers(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--dutch", input, "--tiebreakers", "wins,buchholz"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("custom tiebreakers: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestRunStandings_MissingSystem(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"some-file.trf"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing system: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunStandings_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runStandings(nil, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("no args: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunStandings_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--help"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("help: got exit %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "standings") {
		t.Errorf("help should describe standings command")
	}
}
