// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/validate_test.go
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunValidate_Text(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runValidate([]string{input}, &stdout, &stderr)
	// Should succeed (exit 0) even if there are warnings
	if code != ExitSuccess && code != ExitInvalidInput {
		t.Fatalf("validate text: exit %d, stderr: %s", code, stderr.String())
	}
	// Should produce some output
	if stdout.Len() == 0 {
		t.Error("expected validation output on stdout")
	}
}

func TestRunValidate_JSON(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runValidate([]string{input, "--json"}, &stdout, &stderr)
	if code != ExitSuccess && code != ExitInvalidInput {
		t.Fatalf("validate json: exit %d, stderr: %s", code, stderr.String())
	}
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, stdout.String())
	}
	if _, ok := result["valid"]; !ok {
		t.Error("JSON should contain 'valid' field")
	}
}

func TestRunValidate_MissingFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runValidate([]string{"nonexistent.trf"}, &stdout, &stderr)
	if code != ExitFileAccess {
		t.Errorf("missing file: got exit %d, want %d", code, ExitFileAccess)
	}
}

func TestRunValidate_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runValidate(nil, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("no args: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunValidate_Profiles(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	for _, profile := range []string{"minimal", "standard", "strict"} {
		t.Run(profile, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runValidate([]string{input, "--profile", profile}, &stdout, &stderr)
			if code != ExitSuccess && code != ExitInvalidInput {
				t.Errorf("profile %s: exit %d, stderr: %s", profile, code, stderr.String())
			}
			if stdout.Len() == 0 {
				t.Errorf("profile %s: expected output", profile)
			}
		})
	}
}

func TestRunValidate_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runValidate([]string{"--help"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("help: got exit %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "validate") {
		t.Errorf("help should describe validate command")
	}
}
