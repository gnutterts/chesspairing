// cmd/chesspairing/convert_test.go
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunConvert_RoundTrip(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	outFile := filepath.Join(t.TempDir(), "output.trf")
	var stdout, stderr bytes.Buffer
	code := runConvert([]string{input, "-o", outFile}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("convert: exit %d, stderr: %s", code, stderr.String())
	}

	// Output file should exist and be non-empty
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("output file is empty")
	}
}

func TestRunConvert_MissingInputFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runConvert([]string{"nonexistent.trf", "-o", "out.trf"}, &stdout, &stderr)
	if code != ExitFileAccess {
		t.Errorf("missing file: got exit %d, want %d", code, ExitFileAccess)
	}
}

func TestRunConvert_MissingOutputFlag(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runConvert([]string{input}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing -o: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunConvert_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runConvert(nil, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("no args: got exit %d, want %d", code, ExitInvalidInput)
	}
}
