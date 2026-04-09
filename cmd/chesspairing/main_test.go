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

func TestArgDispatch_GenerateSubcommand(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "gen.trf")
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "generate", "--dutch", "-o", outFile, "-s", "42"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("generate subcommand: got exit %d, want %d, stderr: %s", code, ExitSuccess, stderr.String())
	}
}

func TestArgDispatch_Help(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"--help", []string{"chesspairing", "--help"}},
		{"-h", []string{"chesspairing", "-h"}},
		{"help", []string{"chesspairing", "help"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr)
			if code != ExitSuccess {
				t.Errorf("%s: got exit %d, want %d", tt.name, code, ExitSuccess)
			}
			combined := stdout.String() + stderr.String()
			if !strings.Contains(combined, "Usage") {
				t.Errorf("%s: output should contain 'Usage', got: %s", tt.name, combined)
			}
		})
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

func TestLegacy_UnknownFlag_L(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "--dutch", "-p", "input.trf", "-l", "check.txt"}, &stdout, &stderr)
	// -l is no longer recognized, should error during arg parsing
	if code != ExitInvalidInput {
		t.Errorf("-l flag: got exit %d, want %d", code, ExitInvalidInput)
	}
	if !strings.Contains(stderr.String(), "unexpected argument") {
		t.Errorf("-l flag: expected 'unexpected argument' in stderr, got: %s", stderr.String())
	}
}

func TestGenerate_InvalidConfigValue(t *testing.T) {
	// Write a config file with an invalid integer value
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "bad.cfg")
	if err := os.WriteFile(cfgFile, []byte("PlayersNumber=not-a-number\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	outFile := filepath.Join(dir, "out.trf")
	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "--dutch", "-g", cfgFile, "-o", outFile, "-s", "1"}, &stdout, &stderr)
	if code == ExitSuccess {
		t.Error("expected failure for invalid config value, got ExitSuccess")
	}
	if !strings.Contains(stderr.String(), "PlayersNumber") {
		t.Errorf("stderr should mention the bad key, got: %s", stderr.String())
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

func TestArgDispatch_PairSubcommand(t *testing.T) {
	input := filepath.Join("testdata", "pair-input.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "pair", "--dutch", input}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("pair subcommand: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestArgDispatch_CheckSubcommand(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"chesspairing", "check", "--dutch", input}, &stdout, &stderr)
	// Should succeed or mismatch — not crash
	if code != ExitSuccess && code != ExitNoPairing {
		t.Fatalf("check subcommand: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestPrintUsage_ContainsAllSubcommands(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	usage := buf.String()
	for _, cmd := range []string{"pair", "check", "generate", "validate", "standings", "convert", "version", "tiebreakers"} {
		if !strings.Contains(usage, cmd) {
			t.Errorf("usage should mention %q", cmd)
		}
	}
}
