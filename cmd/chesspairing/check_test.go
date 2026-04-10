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

func TestRunCheck_Match(t *testing.T) {
	// Use a TRF with completed rounds
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"--dutch", input}, &stdout, &stderr)
	// Either matches (0) or mismatches (1) — should not crash
	if code != ExitSuccess && code != ExitNoPairing {
		t.Fatalf("check: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestRunCheck_MissingSystem(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"input.trf"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing system: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunCheck_MissingInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"--dutch"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing input: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunCheck_FileNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"--dutch", "nonexistent.trf"}, &stdout, &stderr)
	if code != ExitFileAccess {
		t.Errorf("missing file: got exit %d, want %d", code, ExitFileAccess)
	}
}

func TestRunCheck_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheck(nil, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("no args: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunCheck_JSON(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"--dutch", input, "--json"}, &stdout, &stderr)
	if code != ExitSuccess && code != ExitNoPairing {
		t.Fatalf("check json: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "{") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestRunCheck_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"--help"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("help: got exit %d, want %d", code, ExitSuccess)
	}
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "check") {
		t.Errorf("help should describe the check command, got: %s", combined)
	}
}
