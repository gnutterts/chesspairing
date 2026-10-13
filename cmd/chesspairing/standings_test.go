// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/standings_test.go
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStandings_Text(t *testing.T) {
	// Need a TRF with completed rounds for standings
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--dutch", input}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("standings text: exit %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Rank") {
		t.Errorf("should contain header, got: %s", out)
	}
}

func TestRunStandings_JSON(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--dutch", input, "--json"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("standings json: exit %d, stderr: %s", code, stderr.String())
	}
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := result["standings"]; !ok {
		t.Error("JSON should contain 'standings' field")
	}
}

func TestRunStandings_CustomTiebreakers(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--dutch", input, "--tiebreakers", "wins,buchholz"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("custom tiebreakers: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestRunStandings_MissingSystem(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"some-file.trf"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing system: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunStandings_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runStandings(nil, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("no args: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunStandings_NoSystemWithTiebreakers(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runStandings([]string{input, "--tiebreakers", "wins,buchholz"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("no system with --tiebreakers: exit %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Rank") {
		t.Errorf("output should contain Rank header, got: %s", stdout.String())
	}
}

func TestRunStandings_InferredRoundRobinWithdrawal(t *testing.T) {
	input := "012 Withdrawal\n092 Round Robin\nXXR 3\n"
	input += "001    1      Player One                        2000 NED                         1.0    1  0002 w 1  0003 w -  0004 w -\n"
	input += "001    2      Player Two                        1900 NED                         0.0    2  0001 b 0  0004 w 1  0003 w 1\n"
	input += "001    3      Player Three                      1800 NED                         0.0    3  0004 w 1  0001 b +  0002 b 0\n"
	input += "001    4      Player Four                       1700 NED                         0.0    4  0003 b 0  0002 b 0  0001 b +\n"
	input += "### chesspairing:withdrawn player=1 after-round=1\n"
	path := filepath.Join(t.TempDir(), "roundrobin.trf")
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := runStandings([]string{path, "--tiebreakers", "buchholz", "--json"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("standings: exit %d, stderr: %s", code, stderr.String())
	}
	var result struct {
		Standings []struct {
			PlayerID string `json:"playerId"`
			Rank     int    `json:"rank"`
		} `json:"standings"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Standings) != 4 {
		t.Fatalf("standings count = %d, want 4", len(result.Standings))
	}
	last := result.Standings[3]
	if last.PlayerID != "1" || last.Rank != 0 {
		t.Errorf("withdrawn player last and unranked: got %+v", last)
	}
}

func TestRunStandings_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--help"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("help: got exit %d, want %d", code, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "standings") {
		t.Errorf("help should describe standings command")
	}
}

func TestRunStandings_MultipleSystemFlags(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--dutch", "--burstein", input}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("multiple systems: exit %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "warning") || !strings.Contains(stderr.String(), "multiple system flags") {
		t.Errorf("should warn about multiple system flags, stderr: %s", stderr.String())
	}
}

func TestRunStandings_OutputFile(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	outFile := filepath.Join(t.TempDir(), "standings.txt")
	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--dutch", input, "-o", outFile}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("standings -o: exit %d, stderr: %s", code, stderr.String())
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
	if !strings.Contains(string(data), "Rank") {
		t.Errorf("output file should contain Rank header, got: %s", string(data))
	}
}

func TestRunStandings_OutputFileJSON(t *testing.T) {
	input := filepath.Join("..", "..", "trf", "testdata", "basic.trf")
	if _, err := os.Stat(input); err != nil {
		t.Skip("test fixture not available")
	}

	outFile := filepath.Join(t.TempDir(), "standings.json")
	var stdout, stderr bytes.Buffer
	code := runStandings([]string{"--dutch", input, "-o", outFile, "--json"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Fatalf("standings -o json: exit %d, stderr: %s", code, stderr.String())
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output file is not valid JSON: %v", err)
	}
	if _, ok := result["standings"]; !ok {
		t.Error("JSON should contain 'standings' field")
	}
}
