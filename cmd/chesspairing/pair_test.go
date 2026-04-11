// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPair_Stdout(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", input}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair stdout: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines of output, got %d", len(lines))
	}
}

func TestRunPair_ToFile(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	outFile := filepath.Join(t.TempDir(), "output.txt")
	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", input, "-o", outFile}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair to file: exit %d, stderr: %s", code, stderr.String())
	}
	if stdout.Len() > 0 {
		t.Errorf("stdout should be empty when writing to file, got: %s", stdout.String())
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("output file is empty")
	}
}

func TestRunPair_MissingSystem(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runPair([]string{"input.trf"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing system: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunPair_MissingInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing input: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunPair_FileNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", "nonexistent.trf"}, &stdout, &stderr)
	if code != ExitFileAccess {
		t.Errorf("missing file: got exit %d, want %d", code, ExitFileAccess)
	}
}

func TestRunPair_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runPair(nil, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("no args: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunPair_JSON(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", input, "--json"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair json: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "{") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestRunPair_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--help"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("help: got exit %d, want %d", code, ExitSuccess)
	}
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "pair") {
		t.Errorf("help should describe the pair command, got: %s", combined)
	}
}

func TestRunPair_FormatWide(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", input, "--format", "wide"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair wide: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Board") {
		t.Errorf("wide format should contain Board header, got: %s", out)
	}
	if !strings.Contains(out, "Kasparov") {
		t.Errorf("wide format should contain player names, got: %s", out)
	}
}

func TestRunPair_DashW(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", input, "-w"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair -w: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Board") {
		t.Errorf("-w should produce wide format, got: %s", out)
	}
}

func TestRunPair_FormatBoard(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", input, "--format", "board"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair board: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Board") {
		t.Errorf("board format should contain Board prefix, got: %s", out)
	}
}

func TestRunPair_FormatXML(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", input, "--format", "xml"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair xml: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "<?xml") {
		t.Errorf("xml format should contain XML declaration, got: %s", out)
	}
	if !strings.Contains(out, "<pairings") {
		t.Errorf("xml format should contain pairings element, got: %s", out)
	}
}

func TestRunPair_FormatOverridesShorthand(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	// --format should win over -w
	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", input, "-w", "--format", "board"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("format override: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	// Board format, not wide
	if strings.Contains(out, "Rtg") {
		t.Errorf("--format board should override -w, but got wide output: %s", out)
	}
	if !strings.Contains(out, "Board") {
		t.Errorf("should contain Board prefix from board format, got: %s", out)
	}
}

func TestRunPair_InvalidFormat(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", input, "--format", "csv"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("invalid format: got exit %d, want %d", code, ExitInvalidInput)
	}
	if !strings.Contains(stderr.String(), "unknown format") {
		t.Errorf("should report unknown format, got: %s", stderr.String())
	}
}

func TestRunPair_MultipleSystemFlags(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runPair([]string{"--dutch", "--burstein", input}, &stdout, &stderr)
	// Should succeed (uses the last system flag) but warn
	if code != ExitSuccess {
		t.Fatalf("multiple systems: exit %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "warning") || !strings.Contains(stderr.String(), "multiple system flags") {
		t.Errorf("should warn about multiple system flags, stderr: %s", stderr.String())
	}
}
