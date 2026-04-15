// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/generate_test.go
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGenerate_BasicWithSeed(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "tournament.trf")
	var stdout, stderr bytes.Buffer
	code := runGenerate(
		[]string{"--dutch", "-o", outFile, "-s", "42"},
		&stdout, &stderr,
	)
	if code != ExitSuccess {
		t.Fatalf("generate: exit %d, stderr: %s", code, stderr.String())
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("output file is empty")
	}
	if !strings.Contains(string(data), "001") {
		t.Error("output should contain player lines (001)")
	}
}

func TestRunGenerate_Deterministic(t *testing.T) {
	outFile1 := filepath.Join(t.TempDir(), "t1.trf")
	outFile2 := filepath.Join(t.TempDir(), "t2.trf")

	var stdout1, stderr1 bytes.Buffer
	code1 := runGenerate(
		[]string{"--dutch", "-o", outFile1, "-s", "hello-world"},
		&stdout1, &stderr1,
	)
	var stdout2, stderr2 bytes.Buffer
	code2 := runGenerate(
		[]string{"--dutch", "-o", outFile2, "-s", "hello-world"},
		&stdout2, &stderr2,
	)

	if code1 != ExitSuccess || code2 != ExitSuccess {
		t.Fatalf("generate failed: code1=%d code2=%d", code1, code2)
	}

	data1, _ := os.ReadFile(outFile1)
	data2, _ := os.ReadFile(outFile2)
	if string(data1) != string(data2) {
		t.Error("same seed should produce identical output")
	}
}

func TestRunGenerate_DifferentSeeds(t *testing.T) {
	outFile1 := filepath.Join(t.TempDir(), "t1.trf")
	outFile2 := filepath.Join(t.TempDir(), "t2.trf")

	var stdout1, stderr1 bytes.Buffer
	runGenerate([]string{"--dutch", "-o", outFile1, "-s", "seed-a"}, &stdout1, &stderr1)
	var stdout2, stderr2 bytes.Buffer
	runGenerate([]string{"--dutch", "-o", outFile2, "-s", "seed-b"}, &stdout2, &stderr2)

	data1, _ := os.ReadFile(outFile1)
	data2, _ := os.ReadFile(outFile2)
	if string(data1) == string(data2) {
		t.Error("different seeds should produce different output")
	}
}

func TestRunGenerate_ConfigFile(t *testing.T) {
	cfgFile := filepath.Join(t.TempDir(), "config.txt")
	if err := os.WriteFile(cfgFile, []byte("PlayersNumber=10\nRoundsNumber=3\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	outFile := filepath.Join(t.TempDir(), "tournament.trf")
	var stdout, stderr bytes.Buffer
	code := runGenerate(
		[]string{"--dutch", "--config", cfgFile, "-o", outFile, "-s", "99"},
		&stdout, &stderr,
	)
	if code != ExitSuccess {
		t.Fatalf("generate with config: exit %d, stderr: %s", code, stderr.String())
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	content := string(data)
	count := strings.Count(content, "\n001 ")
	if count < 10 {
		t.Errorf("expected 10 players, found %d '001' lines", count)
	}
}

func TestRunGenerate_MissingOutputFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runGenerate([]string{"--dutch", "-s", "42"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing -o: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunGenerate_StringSeed(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "tournament.trf")
	var stdout, stderr bytes.Buffer
	code := runGenerate(
		[]string{"--dutch", "-o", outFile, "-s", "my-memorable-seed"},
		&stdout, &stderr,
	)
	if code != ExitSuccess {
		t.Fatalf("string seed: exit %d, stderr: %s", code, stderr.String())
	}
}

func TestRunGenerate_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runGenerate([]string{"--help"}, &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("help: got exit %d, want %d", code, ExitSuccess)
	}
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "generate") {
		t.Errorf("help should describe the generate command, got: %s", combined)
	}
}

func TestRunGenerate_MissingSystem(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "out.trf")
	var stdout, stderr bytes.Buffer
	code := runGenerate([]string{"-o", outFile, "-s", "42"}, &stdout, &stderr)
	if code != ExitInvalidInput {
		t.Errorf("missing system: got exit %d, want %d", code, ExitInvalidInput)
	}
}

func TestRunGenerate_UnexpectedPositionalArgs(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "out.trf")
	var stdout, stderr bytes.Buffer
	code := runGenerate(
		[]string{"--dutch", "-o", outFile, "-s", "42", "extra-arg"},
		&stdout, &stderr,
	)
	if code != ExitSuccess {
		t.Fatalf("generate with extra arg: exit %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "warning") || !strings.Contains(stderr.String(), "extra-arg") {
		t.Errorf("should warn about unexpected arg, stderr: %s", stderr.String())
	}
}

func TestRunGenerate_MultipleSystemFlags(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "out.trf")
	var stdout, stderr bytes.Buffer
	code := runGenerate(
		[]string{"--dutch", "--burstein", "-o", outFile, "-s", "42"},
		&stdout, &stderr,
	)
	if code != ExitSuccess {
		t.Fatalf("multiple systems: exit %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "warning") || !strings.Contains(stderr.String(), "multiple system flags") {
		t.Errorf("should warn about multiple system flags, stderr: %s", stderr.String())
	}
}

func TestRunGenerate_UnknownConfigKey(t *testing.T) {
	cfgFile := filepath.Join(t.TempDir(), "config.txt")
	if err := os.WriteFile(cfgFile, []byte("PlayersNumber=10\nRoundsNumber=3\nBogusKey=42\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	outFile := filepath.Join(t.TempDir(), "tournament.trf")
	var stdout, stderr bytes.Buffer
	code := runGenerate(
		[]string{"--dutch", "--config", cfgFile, "-o", outFile, "-s", "99"},
		&stdout, &stderr,
	)
	if code != ExitSuccess {
		t.Fatalf("unknown config key: exit %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "BogusKey") {
		t.Errorf("should warn about unknown key, stderr: %s", stderr.String())
	}
}

// TestLoadRTGConfig_AllKeys exercises every recognised key in the parser,
// verifies comments and blank lines are skipped, and confirms unknown keys
// emit a warning rather than failing.
func TestLoadRTGConfig_AllKeys(t *testing.T) {
	body := `# Tournament setup
PlayersNumber = 12
RoundsNumber=7

DrawPercentage=15
ForfeitRate=10
RetiredRate=5
HalfPointByeRate=3
HighestRating=2800
LowestRating=1200

# Scoring
PointsForWin=1.0
PointsForDraw=0.5
PointsForLoss=0.0
PointsForZPB=1.0
PointsForForfeitLoss=0.0
PointsForPAB=0.5

# Comment after settings
NotARealKey=ignored
`
	cfgFile := filepath.Join(t.TempDir(), "rtg.cfg")
	if err := os.WriteFile(cfgFile, []byte(body), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := defaultRTGConfig()
	var stderr bytes.Buffer
	if err := loadRTGConfig(cfgFile, &cfg, &stderr); err != nil {
		t.Fatalf("loadRTGConfig: %v", err)
	}

	want := rtgConfig{
		PlayersNumber:        12,
		RoundsNumber:         7,
		DrawPercentage:       15,
		ForfeitRate:          10,
		RetiredRate:          5,
		HalfPointByeRate:     3,
		HighestRating:        2800,
		LowestRating:         1200,
		PointsForWin:         1.0,
		PointsForDraw:        0.5,
		PointsForLoss:        0.0,
		PointsForZPB:         1.0,
		PointsForForfeitLoss: 0.0,
		PointsForPAB:         0.5,
	}
	if cfg != want {
		t.Errorf("config mismatch:\n got: %+v\nwant: %+v", cfg, want)
	}
	if !strings.Contains(stderr.String(), "NotARealKey") {
		t.Errorf("expected warning for unknown key, got: %q", stderr.String())
	}
}

func TestLoadRTGConfig_MissingFile(t *testing.T) {
	cfg := defaultRTGConfig()
	var stderr bytes.Buffer
	err := loadRTGConfig(filepath.Join(t.TempDir(), "no-such-file"), &cfg, &stderr)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadRTGConfig_InvalidValues(t *testing.T) {
	cases := map[string]string{
		"int_key":   "PlayersNumber=not-a-number\n",
		"float_key": "PointsForWin=not-a-float\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			cfgFile := filepath.Join(t.TempDir(), "rtg.cfg")
			if err := os.WriteFile(cfgFile, []byte(body), 0644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
			cfg := defaultRTGConfig()
			var stderr bytes.Buffer
			if err := loadRTGConfig(cfgFile, &cfg, &stderr); err == nil {
				t.Errorf("expected parse error for %s", name)
			}
		})
	}
}

func TestLoadRTGConfig_MalformedLinesSkipped(t *testing.T) {
	// Lines without an `=` must be ignored, not error out.
	body := "this line has no equals\nPlayersNumber=8\nanother bad line\n"
	cfgFile := filepath.Join(t.TempDir(), "rtg.cfg")
	if err := os.WriteFile(cfgFile, []byte(body), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	cfg := defaultRTGConfig()
	var stderr bytes.Buffer
	if err := loadRTGConfig(cfgFile, &cfg, &stderr); err != nil {
		t.Fatalf("loadRTGConfig: %v", err)
	}
	if cfg.PlayersNumber != 8 {
		t.Errorf("PlayersNumber = %d, want 8", cfg.PlayersNumber)
	}
}
