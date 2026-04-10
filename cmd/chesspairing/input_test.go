// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenInput_File(t *testing.T) {
	// Create a temp file with known content
	dir := t.TempDir()
	path := filepath.Join(dir, "test.trf")
	if err := os.WriteFile(path, []byte("012 Test Tournament\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	rc, err := openInput(path)
	if err != nil {
		t.Fatalf("openInput(%q): %v", path, err)
	}
	defer func() { _ = rc.Close() }()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != "012 Test Tournament\n" {
		t.Errorf("got %q, want %q", string(data), "012 Test Tournament\n")
	}
}

func TestOpenInput_Stdin(t *testing.T) {
	rc, err := openInput("-")
	if err != nil {
		t.Fatalf("openInput(-): %v", err)
	}
	// Should return a wrapper around os.Stdin — we can't read from it in tests,
	// but we can verify it's not nil and Close doesn't error
	if rc == nil {
		t.Fatal("expected non-nil ReadCloser for stdin")
	}
	_ = rc.Close() // should be a no-op for stdin
}

func TestOpenInput_MissingFile(t *testing.T) {
	_, err := openInput("/nonexistent/path/file.trf")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestOpenInput_EmptyName(t *testing.T) {
	_, err := openInput("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !strings.Contains(err.Error(), "no input file") {
		t.Errorf("error should mention 'no input file', got: %v", err)
	}
}
