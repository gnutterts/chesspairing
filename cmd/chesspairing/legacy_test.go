// cmd/chesspairing/legacy_test.go
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLegacy_PairToStdout(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runLegacy([]string{"--dutch", input, "-p"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair to stdout: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// First line should be the number of pairings
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines of output, got %d", len(lines))
	}
}

func TestRunLegacy_PairToFile(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	outFile := filepath.Join(t.TempDir(), "output.trf")
	var stdout, stderr bytes.Buffer
	code := runLegacy([]string{"--dutch", input, "-p", outFile}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair to file: exit %d, stderr: %s", code, stderr.String())
	}
	// stdout should be empty when writing to file
	if stdout.Len() > 0 {
		t.Errorf("stdout should be empty when writing to file, got: %s", stdout.String())
	}
	// File should exist and contain pairing data
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("output file is empty")
	}
}

func TestRunLegacy_MissingSystemFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runLegacy([]string{"input.trf", "-p"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing system: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunLegacy_MissingInputFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runLegacy([]string{"--dutch", "nonexistent.trf", "-p"}, &stdout, &stderr)
	if code != ExitFileAccess {
		t.Errorf("missing file: got exit %d, want %d", code, ExitFileAccess)
	}
}

func TestRunLegacy_MissingModeFlag(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runLegacy([]string{"--dutch", input}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("no mode flag: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunLegacy_IgnoredFlags(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	// -w and -q should be accepted and ignored
	var stdout, stderr bytes.Buffer
	code := runLegacy([]string{"--dutch", input, "-p", "-w", "-q", "1000000"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("-w -q ignored: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestRunLegacy_CheckMode(t *testing.T) {
	// Use a TRF with existing rounds — check re-pairs and compares
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runLegacy([]string{"--dutch", input, "-c"}, &stdout, &stderr)
	// Should succeed (match) or return 1 (mismatch) — not crash
	if code != ExitSuccess && code != ExitNoPairing {
		t.Fatalf("check mode: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestRunLegacy_CheckMode_MissingFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runLegacy([]string{"--dutch", "nonexistent.trf", "-c"}, &stdout, &stderr)
	if code != ExitFileAccess {
		t.Errorf("check missing file: got exit %d, want %d", code, ExitFileAccess)
	}
}

func TestRunLegacy_DashR_WithPair(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runLegacy([]string{"-r", "--dutch", input, "-p"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("-r with pair: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	// Should contain version info AND pairing output
	if !strings.Contains(out, "chesspairing") {
		t.Errorf("should contain version info, got: %s", out)
	}
}
