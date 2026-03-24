// cmd/chesspairing/main_test.go
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArgDispatch_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("no args: got exit %d, want %d", code, ExitInvalidInput)
	}
	if !strings.Contains(stderr.String(), "usage") && !strings.Contains(stderr.String(), "Usage") {
		t.Errorf("no args: stderr should contain usage info, got: %s", stderr.String())
	}
}

func TestArgDispatch_VersionSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "version"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("version: got exit %d, want %d", code, ExitSuccess)
	}
}

func TestArgDispatch_UnknownSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "foobar"}, &stdout, &stderr)
	// Unknown first arg that isn't a system flag → invalid input
	if code != ExitInvalidInput {
		t.Errorf("unknown: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestArgDispatch_DashR(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "-r"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("-r: got exit %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "chesspairing") {
		t.Errorf("-r: stdout should contain program name, got: %s", stdout.String())
	}
}

func TestEndToEnd_PairAndStandings(t *testing.T) {
	// Generate a tournament, then compute its standings
	outFile := filepath.Join(t.TempDir(), "tournament.trf")
	var stdout, stderr bytes.Buffer

	// Generate
	code := run([]string{"chesspairing", "--dutch", "-g", "-o", outFile, "-s", "integration-test"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("generate: exit %d, stderr: %s", code, stderr.String())
	}

	// Standings
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"chesspairing", "standings", "--dutch", outFile}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("standings: exit %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Rank") {
		t.Errorf("standings should contain header, got: %s", stdout.String())
	}

	// Validate
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"chesspairing", "validate", outFile}, &stdout, &stderr)
	if code != ExitSuccess && code != ExitInvalidInput {
		t.Fatalf("validate: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestEndToEnd_ConvertRoundTrip(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	outFile := filepath.Join(t.TempDir(), "converted.trf")
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "convert", input, "-o", outFile}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("convert: exit %d, stderr: %s", code, stderr.String())
	}

	// Validate the converted file
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"chesspairing", "validate", outFile}, &stdout, &stderr)
	if code != ExitSuccess && code != ExitInvalidInput {
		t.Fatalf("validate converted: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestEndToEnd_VersionJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "version", "--json"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("version json: exit %d", code)
	}
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestEndToEnd_TiebreakersJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "tiebreakers", "--json"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("tiebreakers json: exit %d", code)
	}
	var result []any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}
