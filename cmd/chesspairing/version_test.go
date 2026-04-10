// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/version_test.go
package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunVersion_Text(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runVersion(nil, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("got exit %d, want %d", code, ExitSuccess)
	}
	out := stdout.String()
	if !strings.Contains(out, "chesspairing") {
		t.Errorf("should contain program name, got: %s", out)
	}
	if !strings.Contains(out, version) {
		t.Errorf("should contain version %q, got: %s", version, out)
	}
	// Should mention supported systems
	if !strings.Contains(out, "dutch") {
		t.Errorf("should list dutch system, got: %s", out)
	}
}

func TestRunVersion_JSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runVersion([]string{"--json"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("got exit %d, want %d", code, ExitSuccess)
	}
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["version"] != version {
		t.Errorf("version: got %v, want %s", result["version"], version)
	}
	systems, ok := result["pairingSystems"].([]any)
	if !ok || len(systems) == 0 {
		t.Errorf("pairingSystems should be non-empty array")
	}
}

func TestRunVersion_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runVersion([]string{"--help"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("help: got exit %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "version") {
		t.Errorf("help should describe version command")
	}
}
