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
