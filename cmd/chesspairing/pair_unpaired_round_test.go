// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// A TRF whose last column holds the requested byes of the round to pair, as
// written by bbpPairings and JaVaFo, must be paired like those programs do.
func TestPairUnpairedRoundColumn(t *testing.T) {
	dir := filepath.Join("..", "..", "internal", "harness", "testdata", "regress")
	files, err := filepath.Glob(filepath.Join(dir, "*.trf"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no regression files: %v", err)
	}
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".trf")
		t.Run(name, func(t *testing.T) {
			meta, err := os.ReadFile(filepath.Join(dir, name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			var want struct {
				BBP []string `json:"bbp"`
			}
			if err := json.Unmarshal(meta, &want); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			if code := runPair([]string{"--dutch", filepath.Join(dir, name+".trf")}, &stdout, &stderr); code != ExitSuccess {
				t.Fatalf("exit code %d: %s", code, stderr.String())
			}
			var got []string
			for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n")[1:] {
				if strings.HasSuffix(line, " 0") {
					continue // a bye
				}
				got = append(got, strings.Join(strings.Fields(line), " "))
			}
			sort.Strings(got)
			var wantPairs []string
			for _, pair := range want.BBP {
				if !strings.HasSuffix(pair, " 0") {
					wantPairs = append(wantPairs, pair)
				}
			}
			sort.Strings(wantPairs)
			if strings.Join(got, "|") != strings.Join(wantPairs, "|") {
				t.Errorf("pairs = %v, bbpPairings gives %v", got, wantPairs)
			}
		})
	}
}
